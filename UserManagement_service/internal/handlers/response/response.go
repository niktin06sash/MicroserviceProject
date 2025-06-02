package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
)

type HTTPResponse struct {
	Success bool              `json:"success"`
	Errors  map[string]string `json:"errors,omitempty"`
	Data    map[string]any    `json:"data,omitempty"`
}

func SendResponse(r *http.Request, w http.ResponseWriter, resp HTTPResponse, status int, traceid string, place string, kafkaprod kafka.KafkaProducerService) {
	ctx := r.Context()
	start := ctx.Value("starttime").(time.Time)
	w.Header().Set("Content-Type", "application/json")
	if ctx.Err() != nil {
		fmterr := fmt.Sprintf("Context error: %v", ctx.Err())
		kafkaprod.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		w.WriteHeader(http.StatusInternalServerError)
		badresp := HTTPResponse{
			Success: false,
			Errors:  map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.RequestTimedOut},
		}
		json.NewEncoder(w).Encode(badresp)
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		duration := time.Since(start).Seconds()
		metrics.UserBadRequestDuration.WithLabelValues(place, metrics.NormalizePath(r.URL.Path)).Observe(duration)
		metrics.UserTotalBadRequests.WithLabelValues(place, metrics.NormalizePath(r.URL.Path)).Inc()
		return
	}
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmterr := fmt.Sprintf("Failed to encode response: %v", err)
		kafkaprod.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		w.WriteHeader(http.StatusInternalServerError)
		badreq := HTTPResponse{
			Success: false,
			Errors:  map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.UserServiceUnavalaible},
		}
		json.NewEncoder(w).Encode(badreq)
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		duration := time.Since(start).Seconds()
		metrics.UserBadRequestDuration.WithLabelValues(place, metrics.NormalizePath(r.URL.Path)).Observe(duration)
		metrics.UserTotalBadRequests.WithLabelValues(place, metrics.NormalizePath(r.URL.Path)).Inc()
	}
	kafkaprod.NewUserLog(kafka.LogLevelInfo, place, traceid, "Succesfull send response to client")
	duration := time.Since(start).Seconds()
	if status == http.StatusOK {
		metrics.UserSuccessfulRequestDuration.WithLabelValues(place, metrics.NormalizePath(r.URL.Path)).Observe(duration)
		metrics.UserTotalSuccessfulRequests.WithLabelValues(place, metrics.NormalizePath(r.URL.Path)).Inc()
	} else {
		metrics.UserBadRequestDuration.WithLabelValues(place, metrics.NormalizePath(r.URL.Path)).Observe(duration)
		metrics.UserTotalBadRequests.WithLabelValues(place, metrics.NormalizePath(r.URL.Path)).Inc()
	}
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
