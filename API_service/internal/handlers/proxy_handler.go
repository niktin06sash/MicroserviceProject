package handlers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
)

type LogProducer interface {
	NewAPILog(c *http.Request, level, place, traceid, msg string)
}

func (h *Handler) ProxyHTTP(c *gin.Context) {
	const place = ProxyHTTP
	traceID := c.MustGet("traceID").(string)
	start := c.MustGet("starttime").(time.Time)
	reqpath := c.Request.URL.Path
	normalizedPath := metrics.NormalizePath(reqpath)
	targetURL, ok := h.Routes[normalizedPath]
	if !ok {
		response.SendResponse(c, http.StatusBadRequest, response.HTTPResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.PageNotFound}}, traceID, place, h.KafkaProducer)
		metrics.APIErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return
	}
	target, err := url.Parse(targetURL)
	if err != nil {
		strerr := fmt.Sprintf("URL-Parse Error: %s", err)
		h.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, strerr)
		response.SendResponse(c, http.StatusInternalServerError, response.HTTPResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.APIServiceUnavalaible}}, traceID, place, h.KafkaProducer)
		metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return
	}
	err = response.CheckContext(c.Request.Context(), traceID, place)
	if err != nil {
		fmterr := fmt.Sprintf("Context error: %v", err)
		h.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, fmterr)
		response.SendResponse(c, http.StatusInternalServerError, response.HTTPResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.RequestTimedOut}}, traceID, place, h.KafkaProducer)
		metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
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
		if userID := c.Param("id"); userID != "" {
			req.URL.Path = fmt.Sprintf("%s/%s", target.Path, userID)
		} else {
			req.URL.Path = target.Path
		}
		req.URL.RawQuery = c.Request.URL.RawQuery
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
		response.SendResponse(c, http.StatusInternalServerError, response.HTTPResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.APIServiceUnavalaible}}, traceID, place, h.KafkaProducer)
		metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
	}
	resp := fmt.Sprintf("Successful HTTP-request to %s", targetURL)
	h.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelInfo, place, traceID, resp)
	proxy.ServeHTTP(c.Writer, c.Request)
	duration := time.Since(start).Seconds()
	metrics.APISuccessfulRequestDuration.WithLabelValues(place, normalizedPath).Observe(duration)
	metrics.APIBackendRequestsTotal.WithLabelValues("User-Service").Inc()
	metrics.APITotalSuccessfulRequests.WithLabelValues(place, normalizedPath).Inc()
}
