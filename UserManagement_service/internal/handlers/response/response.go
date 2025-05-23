package response

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
)

type HTTPResponse struct {
	Success bool              `json:"success"`
	Errors  map[string]string `json:"errors,omitempty"`
	Data    map[string]any    `json:"data,omitempty"`
}

func SendResponse(ctx context.Context, w http.ResponseWriter, resp HTTPResponse, status int, traceid string, place string, kafkaprod kafka.KafkaProducerService) {
	start := ctx.Value("starttime").(time.Time)
	w.Header().Set("Content-Type", "application/json")
	if ctx.Err() != nil {
		fmterr := fmt.Sprintf("Context error: %v", ctx.Err())
		kafkaprod.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		w.WriteHeader(http.StatusInternalServerError)
		badresp := HTTPResponse{
			Success: false,
			Errors:  map[string]string{erro.ServerErrorType: erro.RequestTimedOut},
		}
		json.NewEncoder(w).Encode(badresp)
		metrics.UserErrorsTotal.WithLabelValues(string(erro.ServerErrorType)).Inc()
		duration := time.Since(start).Seconds()
		metrics.UserRequestDuration.WithLabelValues(place).Observe(duration)
		return
	}
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmterr := fmt.Sprintf("Failed to encode response: %v", err)
		kafkaprod.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		w.WriteHeader(http.StatusInternalServerError)
		badreq := HTTPResponse{
			Success: false,
			Errors:  map[string]string{erro.ServerErrorType: erro.UserServiceUnavalaible},
		}
		json.NewEncoder(w).Encode(badreq)
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
	}
	kafkaprod.NewUserLog(kafka.LogLevelInfo, place, traceid, "Succesfull send response to client")
	metrics.UserTotalSuccessfulRequests.WithLabelValues(place).Inc()
	duration := time.Since(start).Seconds()
	metrics.UserRequestDuration.WithLabelValues(place).Observe(duration)
}
func AddSessionCookie(w http.ResponseWriter, sessionID string, expireTime time.Time) {
	maxAge := int(time.Until(expireTime).Seconds())
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, //временное использование так как пока протокол http
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	})
}
func DeleteSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, //временное использование так как пока протокол http
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}
