package middleware

import (
	"context"
	"net/http"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
)

func (m *Middleware) Authorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const place = Authority
		traceID := r.Context().Value("traceID").(string)
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			m.LogProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Required User-ID")
			response.BadResponse(r, w, http.StatusInternalServerError, erro.UserServiceUnavalaible, traceID, place, m.LogProducer)
			metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return
		}
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			m.LogProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Required Session-ID")
			response.BadResponse(r, w, http.StatusInternalServerError, erro.UserServiceUnavalaible, traceID, place, m.LogProducer)
			metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return
		}
		ctx := context.WithValue(r.Context(), "userID", userID)
		ctx = context.WithValue(ctx, "sessionID", sessionID)
		r = r.WithContext(ctx)
		m.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, "Successful authorization verification")
		next.ServeHTTP(w, r)
	})
}
func (m *Middleware) AuthorizedNot(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const place = Not_Authority
		traceID := r.Context().Value("traceID").(string)
		userID := r.Header.Get("X-User-ID")
		if userID != "" {
			m.LogProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Not-Required User-ID")
			response.BadResponse(r, w, http.StatusInternalServerError, erro.UserServiceUnavalaible, traceID, place, m.LogProducer)
			metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return
		}
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID != "" {
			m.LogProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Not-Required Session-ID")
			response.BadResponse(r, w, http.StatusInternalServerError, erro.UserServiceUnavalaible, traceID, place, m.LogProducer)
			metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return
		}
		m.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, "Successful unauthorization verification")
		next.ServeHTTP(w, r)
	})
}
