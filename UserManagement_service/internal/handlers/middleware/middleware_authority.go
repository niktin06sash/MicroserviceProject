package middleware

import (
	"context"
	"net/http"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
)

func (m *Middleware) Authorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const place = Authority
		traceID := r.Context().Value("traceID").(string)
		maparesponse := make(map[string]string)
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			m.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Required User-ID")
			maparesponse[erro.ErrorType] = erro.ServerErrorType
			maparesponse[erro.ErrorMessage] = erro.UserServiceUnavalaible
			response.SendResponse(r, w, response.HTTPResponse{Success: false, Data: nil, Errors: maparesponse}, http.StatusInternalServerError, traceID, place, m.KafkaProducer)
			metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return
		}
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			m.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Required Session-ID")
			maparesponse[erro.ErrorType] = erro.ServerErrorType
			maparesponse[erro.ErrorMessage] = erro.UserServiceUnavalaible
			response.SendResponse(r, w, response.HTTPResponse{Success: false, Data: nil, Errors: maparesponse}, http.StatusInternalServerError, traceID, place, m.KafkaProducer)
			metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return
		}
		ctx := context.WithValue(r.Context(), "userID", userID)
		ctx = context.WithValue(ctx, "sessionID", sessionID)
		r = r.WithContext(ctx)
		m.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, "Successful authorization verification")
		next.ServeHTTP(w, r)
	})
}
func (m *Middleware) AuthorizedNot(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const place = Not_Authority
		traceID := r.Context().Value("traceID").(string)
		maparesponse := make(map[string]string)
		userID := r.Header.Get("X-User-ID")
		if userID != "" {
			m.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Not-Required User-ID")
			maparesponse[erro.ErrorType] = erro.ServerErrorType
			maparesponse[erro.ErrorMessage] = erro.UserServiceUnavalaible
			response.SendResponse(r, w, response.HTTPResponse{Success: false, Data: nil, Errors: maparesponse}, http.StatusInternalServerError, traceID, place, m.KafkaProducer)
			metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return
		}
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID != "" {
			m.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Not-Required Session-ID")
			maparesponse[erro.ErrorType] = erro.ServerErrorType
			maparesponse[erro.ErrorMessage] = erro.UserServiceUnavalaible
			response.SendResponse(r, w, response.HTTPResponse{Success: false, Data: nil, Errors: maparesponse}, http.StatusInternalServerError, traceID, place, m.KafkaProducer)
			metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return
		}
		m.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, "Successful unauthorization verification")
		next.ServeHTTP(w, r)
	})
}
