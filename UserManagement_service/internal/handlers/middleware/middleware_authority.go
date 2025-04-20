package middleware

import (
	"context"
	"net/http"
	"time"
)

func Middleware_Authorized(isauthority bool) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value("requestID").(string)
			switch isauthority {
			case false:
				userID := r.Header.Get("X-User-ID")
				if userID != "" {
					logRequest(r, "Non-Authority", requestID, true, "Not-Required User-ID")
					//рассмотреть возможность такого же ответа клиенту в случае ошибки как в обработчиках(utils?)
					http.Error(w, "Not-Required User-ID", http.StatusBadRequest)
					return
				}
				sessionID := r.Header.Get("X-Session-ID")
				if sessionID != "" {
					logRequest(r, "Non-Authority", requestID, true, "Not-Required Session-ID")
					//рассмотреть возможность такого же ответа клиенту в случае ошибки как в обработчиках(utils?)
					http.Error(w, "Not-Required Session-ID", http.StatusBadRequest)
					return
				}
				logRequest(r, "Non-Authority", requestID, false, "")
			case true:
				userID := r.Header.Get("X-User-ID")
				if userID == "" {
					logRequest(r, "Authority", requestID, true, "Required User-ID")
					//рассмотреть возможность такого же ответа клиенту в случае ошибки как в обработчиках(utils?)
					http.Error(w, "Required User-ID", http.StatusBadRequest)
					return
				}
				sessionID := r.Header.Get("X-Session-ID")
				if sessionID == "" {
					logRequest(r, "Authority", requestID, true, "Required Session-ID")
					//рассмотреть возможность такого же ответа клиенту в случае ошибки как в обработчиках(utils?)
					http.Error(w, "Required Session-ID", http.StatusBadRequest)
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
