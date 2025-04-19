package middleware

import (
	"context"
	"log"
	"net/http"
	"time"
)

func LogRequest(r *http.Request, requestID string, isError bool, errorMessage string) {
	ip := r.RemoteAddr
	method := r.Method
	path := r.URL.Path

	if isError {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] [IP: %s] [Method: %s] [Path: %s] Error: %s", requestID, ip, method, path, errorMessage)
	} else {
		log.Printf("[INFO] [UserManagement] [RequestID: %s] [IP: %s] [Method: %s] [Path: %s]", requestID, ip, method, path)
	}
}
func UserManagementMiddleware_NonAuthority(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			LogRequest(r, "", true, "Missing X-Request-ID")
			//рассмотреть возможность такого же ответа клиенту в случае ошибки как в обработчиках(utils?)
			http.Error(w, "Required Request-ID", http.StatusBadRequest)
			return
		}
		userID := r.Header.Get("X-User-ID")
		if userID != "" {
			LogRequest(r, requestID, true, "Not-Required User-ID")
			//рассмотреть возможность такого же ответа клиенту в случае ошибки как в обработчиках(utils?)
			http.Error(w, "Not-Required User-ID", http.StatusBadRequest)
			return
		}
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID != "" {
			LogRequest(r, requestID, true, "Not-Required Session-ID")
			//рассмотреть возможность такого же ответа клиенту в случае ошибки как в обработчиках(utils?)
			http.Error(w, "Not-Required Session-ID", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), "requestID", requestID)
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		r = r.WithContext(ctx)
		LogRequest(r, requestID, false, "")
		next.ServeHTTP(w, r)
	}
}
func UserManagementMiddleware_Authority(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			LogRequest(r, "", true, "Missing X-Request-ID")
			//рассмотреть возможность такого же ответа клиенту в случае ошибки как в обработчиках(utils?)
			http.Error(w, "Required Request-ID", http.StatusBadRequest)
			return
		}
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			LogRequest(r, requestID, true, "Required User-ID")
			//рассмотреть возможность такого же ответа клиенту в случае ошибки как в обработчиках(utils?)
			http.Error(w, "Required User-ID", http.StatusBadRequest)
			return
		}
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			LogRequest(r, requestID, true, "Required Session-ID")
			//рассмотреть возможность такого же ответа клиенту в случае ошибки как в обработчиках(utils?)
			http.Error(w, "Required Session-ID", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), "requestID", requestID)
		ctx = context.WithValue(ctx, "userID", userID)
		ctx = context.WithValue(ctx, "sessionID", sessionID)
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		r = r.WithContext(ctx)
		LogRequest(r, requestID, false, "")
		next.ServeHTTP(w, r)
	}
}
