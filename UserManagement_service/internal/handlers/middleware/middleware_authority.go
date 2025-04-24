package middleware

import (
	"context"
	"net/http"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
)

func Middleware_Authorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Context().Value("traceID").(string)
		maparesponse := make(map[string]string)
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			logRequest(r, "Authority", traceID, true, "Required User-ID")
			maparesponse["InternalServerError"] = erro.ErrorRequiredUserID.Error()
			br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
			response.SendResponse(w, br, traceID, "Authority")
			return
		}
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			logRequest(r, "Authority", traceID, true, "Required Session-ID")
			maparesponse["InternalServerError"] = erro.ErrorRequiredSessionID.Error()
			br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
			response.SendResponse(w, br, traceID, "Authority")
			return
		}
		ctx := context.WithValue(r.Context(), "userID", userID)
		ctx = context.WithValue(ctx, "sessionID", sessionID)
		r = r.WithContext(ctx)
		logRequest(r, "Authority", traceID, false, "Successful authorization verification")
		next.ServeHTTP(w, r)
	})
}
func Middleware_AuthorizedNot(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Context().Value("traceID").(string)
		maparesponse := make(map[string]string)
		userID := r.Header.Get("X-User-ID")
		if userID != "" {
			logRequest(r, "Not-Authority", traceID, true, "Not-Required User-ID")
			maparesponse["InternalServerError"] = erro.ErrorNotRequiredUserID.Error()
			br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
			response.SendResponse(w, br, traceID, "Not-Authority")
			return
		}
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID != "" {
			logRequest(r, "Not-Authority", traceID, true, "Not-Required Session-ID")
			maparesponse["InternalServerError"] = erro.ErrorNotRequiredUserID.Error()
			br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
			response.SendResponse(w, br, traceID, "Not-Authority")
			return
		}
		logRequest(r, "Non-Authority", traceID, false, "Successful unauthorization verification")
		next.ServeHTTP(w, r)
	})
}
