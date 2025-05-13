package response

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
)

type HTTPResponse struct {
	Success bool              `json:"success"`
	Errors  map[string]string `json:"errors"`
	Data    map[string]any    `json:"data,omitempty"`
	Status  int               `json:"status"`
}

func SendResponse(ctx context.Context, w http.ResponseWriter, success bool, data map[string]any, errors map[string]string, status int, traceid string, place string, kafkaprod kafka.KafkaProducerService) {
	start := ctx.Value("starttime").(time.Time)
	w.Header().Set("Content-Type", "application/json")
	if ctx.Err() != nil {
		fmterr := fmt.Sprintf("Context error: %v", ctx.Err())
		kafkaprod.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		w.WriteHeader(http.StatusInternalServerError)
		badreq := HTTPResponse{
			Success: false,
			Errors:  map[string]string{"InternalServerError": "Context deadline exceeded"},
			Status:  http.StatusInternalServerError,
		}
		json.NewEncoder(w).Encode(badreq)
		metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
		duration := time.Since(start).Seconds()
		metrics.UserRequestDuration.WithLabelValues(place).Observe(duration)
		return
	}
	resp := HTTPResponse{
		Success: success,
		Errors:  errors,
		Data:    data,
		Status:  status,
	}
	w.WriteHeader(resp.Status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmterr := fmt.Sprintf("Failed to encode response: %v", err)
		kafkaprod.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		w.WriteHeader(http.StatusInternalServerError)
		badreq := HTTPResponse{
			Success: false,
			Errors:  map[string]string{"InternalServerError": "EncoderResponse Error"},
			Status:  http.StatusInternalServerError,
		}
		json.NewEncoder(w).Encode(badreq)
		metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
		duration := time.Since(start).Seconds()
		metrics.UserRequestDuration.WithLabelValues(place).Observe(duration)
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
func ConvertErrorsToString(errors map[string]error) map[string]string {
	stringMap := make(map[string]string)
	for key, err := range errors {
		if err != nil {
			stringMap[key] = err.Error()
		}
	}
	return stringMap
}
