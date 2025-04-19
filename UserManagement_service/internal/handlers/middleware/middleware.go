package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
)

type customResponseWriter struct {
	http.ResponseWriter
	reqstatuscode int
}

func (crw *customResponseWriter) WriteHeader(code int) {
	crw.reqstatuscode = code
	crw.ResponseWriter.WriteHeader(code)
}
func (crw *customResponseWriter) Status() int {
	return crw.reqstatuscode
}
func retryRequest(next http.HandlerFunc, w http.ResponseWriter, r *http.Request, requestID string) error {
	for attempt := 0; attempt < 3; attempt++ {
		rw := &customResponseWriter{ResponseWriter: w}
		next.ServeHTTP(rw, r)

		if rw.Status() >= 500 && rw.Status() < 600 {
			log.Printf("[WARN] [UserManagement] [RequestID: %s] Warn: %d Attempt failed! Retrying...", requestID, attempt)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		return nil
	}

	return erro.ErrorAllRetryFailed
}
func logRequest(r *http.Request, requestID string, isError bool, errorMessage string) {
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
			log.Println("[WARN] [UserManagement] Warn: Required Request-ID")
			requestID = uuid.New().String()
		}
		userID := r.Header.Get("X-User-ID")
		if userID != "" {
			logRequest(r, requestID, true, "Not-Required User-ID")
			//рассмотреть возможность такого же ответа клиенту в случае ошибки как в обработчиках(utils?)
			http.Error(w, "Not-Required User-ID", http.StatusBadRequest)
			return
		}
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID != "" {
			logRequest(r, requestID, true, "Not-Required Session-ID")
			//рассмотреть возможность такого же ответа клиенту в случае ошибки как в обработчиках(utils?)
			http.Error(w, "Not-Required Session-ID", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), "requestID", requestID)
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		r = r.WithContext(ctx)
		logRequest(r, requestID, false, "")
		if err := retryRequest(next, w, r, requestID); err != nil {
			logRequest(r, requestID, true, err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
func UserManagementMiddleware_Authority(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			log.Println("[WARN] [UserManagement] Warn: Required Request-ID")
			//рассмотреть возможность такого же ответа клиенту в случае ошибки как в обработчиках(utils?)
			http.Error(w, "Required User-ID", http.StatusBadRequest)
			return
		}
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			logRequest(r, requestID, true, "Required Session-ID")
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
		logRequest(r, requestID, false, "")
		if err := retryRequest(next, w, r, requestID); err != nil {
			logRequest(r, requestID, true, err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
