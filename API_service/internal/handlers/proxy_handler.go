package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
)

func (h *Handler) ProxyHTTP(c *gin.Context) {
	traceID := c.MustGet("traceID").(string)
	maparesponse := make(map[string]string)
	reqpath := c.Request.URL.Path
	targetURL, ok := h.Routes[reqpath]
	if !ok {
		maparesponse["ClientError"] = "Page not found"
		response.SendResponse(c, http.StatusBadRequest, false, nil, maparesponse, traceID, "ProxyHTTP")
		return
	}
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Printf("[ERROR] [API-Service] [ProxyHTTP] [TraceID: %s] URL-Parse Error: %s", traceID, err)
		maparesponse["InternalServerError"] = "Url-Parse Error"
		response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse, traceID, "ProxyHTTP")
		return
	}
	err = response.CheckContext(c.Request.Context(), traceID, "ProxyHTTP")
	if err != nil {
		maparesponse["InternalServerError"] = "Context canceled or deadline exceeded"
		response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse, traceID, "ProxyHTTP")
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		deadline, ok := c.Request.Context().Deadline()
		if !ok {
			log.Println("[WARN] [API-Gateway] Failed to get deadline from context")
			deadline = time.Now().Add(20 * time.Second)
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
		log.Printf("[ERROR] [API-Service] [ProxyHTTP] [TraceID: %v] Proxy error: %s", traceID, err)
		if req.Context().Err() != nil {
			log.Printf("[WARN] [API-Service] [ProxyHTTP] [TraceID: %v] Context canceled or deadline exceeded", traceID)
		}
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
			log.Printf("[ERROR] [API-Service] [ProxyHTTP] [TraceID: %v] Failed to send response: %v", traceID, err)
		}
	}
	log.Printf("[ERROR] [API-Service] [ProxyHTTP] [TraceID: %v] Successful HTTP-request to %s", traceID, targetURL)
	proxy.ServeHTTP(c.Writer, c.Request)
}
