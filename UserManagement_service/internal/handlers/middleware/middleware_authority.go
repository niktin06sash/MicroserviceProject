package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
)

func Middleware_Authorized(isauthority bool) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			traceID := r.Context().Value("traceID").(string)
			maparesponse := make(map[string]string)
			switch isauthority {
			case false:
				userID := r.Header.Get("X-User-ID")
				if userID != "" {
					logRequest(r, "Not-Authority", traceID, true, "Not-Required User-ID")
					maparesponse["UserId"] = erro.ErrorNotRequiredUserID.Error()
					br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
					response.SendResponse(w, br, traceID)
					return
				}
				sessionID := r.Header.Get("X-Session-ID")
				if sessionID != "" {
					logRequest(r, "Not-Authority", traceID, true, "Not-Required Session-ID")
					maparesponse["SessionId"] = erro.ErrorNotRequiredUserID.Error()
					br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
					response.SendResponse(w, br, traceID)
					return
				}
				logRequest(r, "Non-Authority", traceID, false, "")
			case true:
				userID := r.Header.Get("X-User-ID")
				if userID == "" {
					logRequest(r, "Authority", traceID, true, "Required User-ID")
					maparesponse["UserId"] = erro.ErrorRequiredUserID.Error()
					br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
					response.SendResponse(w, br, traceID)
					return
				}
				sessionID := r.Header.Get("X-Session-ID")
				if sessionID == "" {
					logRequest(r, "Authority", traceID, true, "Required Session-ID")
					maparesponse["SessionId"] = erro.ErrorRequiredSessionID.Error()
					br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
					response.SendResponse(w, br, traceID)
					return
				}
				ctx := context.WithValue(r.Context(), "userID", userID)
				ctx = context.WithValue(ctx, "sessionID", sessionID)
				r = r.WithContext(ctx)
				logRequest(r, "Authority", traceID, false, "")
			}
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
	}
}
