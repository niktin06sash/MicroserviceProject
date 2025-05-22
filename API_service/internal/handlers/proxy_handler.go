package handlers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
)

func (h *Handler) ProxyHTTP(c *gin.Context) {
	var place = "API-ProxyHTTP"
	traceID := c.MustGet("traceID").(string)
	start := c.MustGet("starttime").(time.Time)
	maparesponse := make(map[string]string)
	reqpath := c.Request.URL.Path
	targetURL, ok := h.Routes[reqpath]
	if !ok {
		maparesponse["ClientError"] = "Page not found"
		response.SendResponse(c, http.StatusBadRequest, false, nil, maparesponse, traceID, place, h.KafkaProducer)
		metrics.APIErrorsTotal.WithLabelValues("ClientError").Inc()
		return
	}
	target, err := url.Parse(targetURL)
	if err != nil {
		strerr := fmt.Sprintf("URL-Parse Error: %s", err)
		h.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, strerr)
		maparesponse["InternalServerError"] = "API-Service is unavailable"
		response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse, traceID, place, h.KafkaProducer)
		metrics.APIErrorsTotal.WithLabelValues("InternalServerError").Inc()
		return
	}
	err = response.CheckContext(c.Request.Context(), traceID, place)
	if err != nil {
		fmterr := fmt.Sprintf("Context error: %v", err)
		h.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, fmterr)
		maparesponse["InternalServerError"] = "Request timed out"
		response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse, traceID, place, h.KafkaProducer)
		metrics.APIErrorsTotal.WithLabelValues("InternalServerError").Inc()
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		deadline, ok := c.Request.Context().Deadline()
		if !ok {
			h.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, "Failed to get deadline from context")
			deadline = time.Now().Add(15 * time.Second)
		}
		req.Header.Set("X-Deadline", deadline.Format(time.RFC3339))
		req.Host = target.Host
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
		if userID, exists := c.Get("userID"); exists {
			req.Header.Set("X-User-ID", userID.(string))
		}
		if sessionID, exists := c.Get("sessionID"); exists {
			req.Header.Set("X-Session-ID", sessionID.(string))
		}
		traceID := c.MustGet("traceID").(string)
		req.Header.Set("X-Trace-ID", traceID)
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		traceID := c.MustGet("traceID").(string)
		strerr := fmt.Sprintf("Proxy error: %s", err)
		h.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, strerr)
		maparesponse := map[string]string{"InternalServerError": "API-Service is unavailable"}
		response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse, traceID, place, h.KafkaProducer)
		metrics.APIErrorsTotal.WithLabelValues("InternalServerError").Inc()
		return
	}
	resp := fmt.Sprintf("Successful HTTP-request to %s", targetURL)
	h.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelInfo, place, traceID, resp)
	proxy.ServeHTTP(c.Writer, c.Request)
	duration := time.Since(start).Seconds()
	metrics.APIRequestDuration.WithLabelValues(place, c.Request.URL.Path).Observe(duration)
	metrics.APIBackendRequestsTotal.WithLabelValues("User-Service").Inc()
	metrics.APITotalSuccessfulRequests.WithLabelValues(c.Request.URL.Path).Inc()
}
