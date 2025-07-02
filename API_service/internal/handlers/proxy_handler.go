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

func (h *Handler) ProxyHTTP(c *gin.Context, place string) {
	traceID := c.MustGet("traceID").(string)
	start := c.MustGet("starttime").(time.Time)
	reqpath := c.Request.URL.Path
	normalizedPath := metrics.NormalizePath(reqpath)
	targetURL, ok := h.Routes[normalizedPath]
	if !ok {
		response.BadResponse(c, http.StatusBadRequest, erro.ClientError(erro.PageNotFound), traceID, place, h.logproducer)
		metrics.APIErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return
	}
	target, err := url.Parse(targetURL)
	if err != nil {
		strerr := fmt.Sprintf("URL-Parse Error: %s", err)
		h.logproducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, strerr)
		response.BadResponse(c, http.StatusInternalServerError, erro.ServerError(erro.APIServiceUnavalaible), traceID, place, h.logproducer)
		metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return
	}
	err = response.CheckContext(c, place, traceID, h.logproducer)
	if err != nil {
		response.BadResponse(c, http.StatusInternalServerError, erro.ServerError(erro.RequestTimedOut), traceID, place, h.logproducer)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		req.Host = target.Host
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
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
		h.logproducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, strerr)
		response.BadResponse(c, http.StatusInternalServerError, erro.ServerError(erro.APIServiceUnavalaible), traceID, place, h.logproducer)
		metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
	}
	h.logproducer.NewAPILog(c.Request, kafka.LogLevelInfo, place, traceID, fmt.Sprintf("Successful Proxy-HTTP-request to %s", targetURL))
	proxy.ServeHTTP(c.Writer, c.Request)
	duration := time.Since(start).Seconds()
	metrics.APISuccessfulRequestDuration.WithLabelValues(place, normalizedPath).Observe(duration)
	metrics.APIBackendRequestsTotal.WithLabelValues("User-Service").Inc()
	metrics.APITotalSuccessfulRequests.WithLabelValues(place, normalizedPath).Inc()
}
