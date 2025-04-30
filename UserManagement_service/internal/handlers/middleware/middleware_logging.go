package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func logRequest(r *http.Request, place string, requestID string, isError bool, errorMessage string) {
	ip := r.RemoteAddr
	method := r.Method
	path := r.URL.Path
	if isError {
		log.Printf("[ERROR] [UserManagement] [%s] [TraceID: %s] [IP: %s] [Method: %s] [Path: %s] Error: %s", place, requestID, ip, method, path, errorMessage)
	} else {
		log.Printf("[INFO] [UserManagement] [%s] [TraceID: %s] [IP: %s] [Method: %s] [Path: %s] %s", place, requestID, ip, method, path, errorMessage)
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
		deadlinectx := r.Header.Get("X-Deadline")
		var deadline time.Time
		deadline, err := time.Parse(time.RFC3339, deadlinectx)
		if err != nil {
			log.Printf("[ERROR] [UserManagement] Failed to parse X-Deadline: %v", err)
			deadline = time.Now().Add(15 * time.Second)
		}
		ctx, cancel := context.WithDeadline(ctx, deadline)
		defer cancel()
		r = r.WithContext(ctx)
		logRequest(r, "Logging", traceID, false, "")
		next.ServeHTTP(w, r)
	})
}
