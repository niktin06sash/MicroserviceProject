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
		log.Printf("[ERROR] [UserManagement] [%s] [TraceID: %s] [IP: %s] [Method: %s] [Path: %s] Error: %s", place, requestID, ip, method, path, errorMessage)
	} else {
		log.Printf("[INFO] [UserManagement] [%s] [TraceID: %s] [IP: %s] [Method: %s] [Path: %s]", place, requestID, ip, method, path)
	}
}
func Middleware_Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			log.Println("[WARN] [UserManagement] Warn: Required Request-ID")
			traceID = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), "traceID", traceID)
		r = r.WithContext(ctx)
		logRequest(r, "Logging", traceID, false, "")
		next.ServeHTTP(w, r)
	})
}
