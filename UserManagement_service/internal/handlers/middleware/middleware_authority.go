package middleware

import (
	"context"
	"net/http"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
)

func (m *Middleware) Authorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var place = "Middleware-Authority"
		traceID := r.Context().Value("traceID").(string)
		maparesponse := make(map[string]string)
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			m.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Required User-ID")
			maparesponse["InternalServerError"] = erro.ErrorRequiredUserID.Error()
			response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusInternalServerError, traceID, place, m.KafkaProducer)
			return
		}
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			m.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Required Session-ID")
			maparesponse["InternalServerError"] = erro.ErrorRequiredSessionID.Error()
			response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusInternalServerError, traceID, place, m.KafkaProducer)
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
		var place = "Middleware-Not-Authority"
		traceID := r.Context().Value("traceID").(string)
		maparesponse := make(map[string]string)
		userID := r.Header.Get("X-User-ID")
		if userID != "" {
			m.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Not-Required User-ID")
			maparesponse["InternalServerError"] = erro.ErrorNotRequiredUserID.Error()
			response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusInternalServerError, traceID, place, m.KafkaProducer)
			return
		}
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID != "" {
			m.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Not-Required Session-ID")
			maparesponse["InternalServerError"] = erro.ErrorNotRequiredUserID.Error()
			response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusInternalServerError, traceID, place, m.KafkaProducer)
			return
		}
		m.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, "Successful unauthorization verification")
		next.ServeHTTP(w, r)
	})
}
