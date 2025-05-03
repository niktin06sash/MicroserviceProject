package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/kafka"
)

func (h *Handler) ProxyHTTP(c *gin.Context) {
	var place = "ProxyHTTP"
	traceID := c.MustGet("traceID").(string)
	maparesponse := make(map[string]string)
	reqpath := c.Request.URL.Path
	targetURL, ok := h.Routes[reqpath]
	if !ok {
		maparesponse["ClientError"] = "Page not found"
		response.SendResponse(c, http.StatusBadRequest, false, nil, maparesponse, traceID, place, h.KafkaProducer)
		return
	}
	target, err := url.Parse(targetURL)
	if err != nil {
		strerr := fmt.Sprintf("URL-Parse Error: %s", err)
		h.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, strerr)
		maparesponse["InternalServerError"] = "Url-Parse Error"
		response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse, traceID, "ProxyHTTP", h.KafkaProducer)
		return
	}
	err = response.CheckContext(c.Request.Context(), traceID, "ProxyHTTP")
	if err != nil {
		fmterr := fmt.Sprintf("Context error: %v", err)
		h.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, fmterr)
		maparesponse["InternalServerError"] = "Context canceled or deadline exceeded"
		response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse, traceID, "ProxyHTTP", h.KafkaProducer)
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
		maparesponse := map[string]string{"InternalServerError": "Proxy Error"}
		response := response.HTTPResponse{
			Success: false,
			Data:    nil,
			Errors:  maparesponse,
			Status:  http.StatusInternalServerError,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			strerr := fmt.Sprintf("Failed to send response: %s", err)
			h.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, strerr)
		}
	}
	resp := fmt.Sprintf("Successful HTTP-request to %s", targetURL)
	h.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelInfo, place, traceID, resp)
	proxy.ServeHTTP(c.Writer, c.Request)
}
