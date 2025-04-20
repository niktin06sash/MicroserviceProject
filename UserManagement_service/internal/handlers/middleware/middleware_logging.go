package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func logRequest(r *http.Request, place string, requestID string, isError bool, errorMessage string) {
	ip := r.RemoteAddr
	method := r.Method
	path := r.URL.Path
	if isError {
		log.Printf("[ERROR] [UserManagement] [%s] [RequestID: %s] [IP: %s] [Method: %s] [Path: %s] Error: %s", place, requestID, ip, method, path, errorMessage)
	} else {
		log.Printf("[INFO] [UserManagement] [%s] [RequestID: %s] [IP: %s] [Method: %s] [Path: %s]", place, requestID, ip, method, path)
	}
}
func Middleware_Logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			log.Println("[WARN] [UserManagement] Warn: Required Request-ID")
			requestID = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), "requestID", requestID)
		r = r.WithContext(ctx)
		logRequest(r, "Logging", requestID, false, "")
		next.ServeHTTP(w, r)
	}
}
