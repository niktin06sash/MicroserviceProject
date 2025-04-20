package middleware

import (
	"log"
	"net/http"
	"time"

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
func Middleware_Retry(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value("requestID").(string)
		err := retryRequest(next, w, r, requestID)
		if err != nil {
			logRequest(r, "Retry", requestID, true, "All attempts failed!")
			//рассмотреть возможность такого же ответа клиенту в случае ошибки как в обработчиках(utils?)
			http.Error(w, "Not-Required User-ID", http.StatusBadRequest)
			return
		}
	}
}
