package handlers

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

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
		response.SendResponse(c, http.StatusNotFound, false, nil, maparesponse)
		return
	}
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Printf("[ERROR] [API-Service] [ProxyHTTP] [TraceID: %s] URL-Parse Error: %s", traceID, err)
		maparesponse["InternalServerError"] = "Url-Parse Error"
		response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse)
		return
	}
	err = response.CheckContext(c.Request.Context(), traceID, "ProxyHTTP")
	if err != nil {
		maparesponse["InternalServerError"] = "Context canceled or deadline exceeded"
		response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		req.Host = target.Host
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = c.Request.URL.Path
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
		log.Printf("[ERROR] [API-Service] [ProxyHTTP] [TraceID: %v] Proxy error: %s", traceID, err)
		maparesponse["InternalServerError"] = "Proxy Error"
		response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse)
	}
	log.Printf("[ERROR] [API-Service] [ProxyHTTP] [TraceID: %v] Successful HTTP-request to %s", traceID, target)
	proxy.ServeHTTP(c.Writer, c.Request)
}
func (h *Handler) ProxtGrpc(c *gin.Context) {

}
