package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] Failed to read request body: %v", requestID, err)
		return erro.ErrorReadAll
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	for attempt := 0; attempt < 3; attempt++ {
		rw := &customResponseWriter{ResponseWriter: w}
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		next.ServeHTTP(rw, r)
		if rw.Status() >= 500 && rw.Status() < 600 {
			log.Printf("[WARN] [UserManagement] [RequestID: %s] Warn: %d Attempt failed! Retrying...", requestID, attempt+1)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		return nil
	}

	return erro.ErrorInternalServer
}
func Middleware_Retry(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value("requestID").(string)
		maparesponse := make(map[string]string)
		err := retryRequest(next, w, r, requestID)
		if err != nil {
			logRequest(r, "Retry", requestID, true, "All attempts failed!")
			maparesponse["Retry"] = err.Error()
			br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
			response.SendResponse(w, br, requestID)
			return
		}
	}
}
