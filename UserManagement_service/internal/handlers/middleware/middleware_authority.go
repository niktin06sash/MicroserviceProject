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
			requestID := r.Context().Value("requestID").(string)
			maparesponse := make(map[string]string)
			switch isauthority {
			case false:
				userID := r.Header.Get("X-User-ID")
				if userID != "" {
					logRequest(r, "Not-Authority", requestID, true, "Not-Required User-ID")
					maparesponse["UserId"] = erro.ErrorNotRequiredUserID.Error()
					br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
					response.SendResponse(w, br)
					return
				}
				sessionID := r.Header.Get("X-Session-ID")
				if sessionID != "" {
					logRequest(r, "Not-Authority", requestID, true, "Not-Required Session-ID")
					maparesponse["SessionId"] = erro.ErrorNotRequiredUserID.Error()
					br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
					response.SendResponse(w, br)
					return
				}
				logRequest(r, "Non-Authority", requestID, false, "")
			case true:
				userID := r.Header.Get("X-User-ID")
				if userID == "" {
					logRequest(r, "Authority", requestID, true, "Required User-ID")
					maparesponse["UserId"] = erro.ErrorRequiredUserID.Error()
					br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
					response.SendResponse(w, br)
					return
				}
				sessionID := r.Header.Get("X-Session-ID")
				if sessionID == "" {
					logRequest(r, "Authority", requestID, true, "Required Session-ID")
					maparesponse["SessionId"] = erro.ErrorRequiredSessionID.Error()
					br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
					response.SendResponse(w, br)
					return
				}
				ctx := context.WithValue(r.Context(), "userID", userID)
				ctx = context.WithValue(ctx, "sessionID", sessionID)
				r = r.WithContext(ctx)
				logRequest(r, "Authority", requestID, false, "")
			}
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
	}
}
