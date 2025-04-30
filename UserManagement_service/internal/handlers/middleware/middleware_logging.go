package middleware

import (
	"log"
	"net/http"
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
