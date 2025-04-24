package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
)

// ProxyHTTP handles HTTP requests and proxies them to the target service.
// @Summary Register a new user
// @Description Register a new user by sending user data to the target service.
// @Tags User Management
// @Accept json
// @Produce json
// @Param request body PersonReg true "User registration data"
// @Success 200 {object} HTTPResponse "User successfully registered"
// @Failure 400 {object} HTTPResponse "Invalid input data"
// @Failure 500 {object} HTTPResponse "Internal server error"
// @Router /reg [post]
//
// @Summary Authenticate a user
// @Description Authenticate a user by sending credentials to the target service.
// @Tags User Management
// @Accept json
// @Produce json
// @Param request body PersonAuth true "User credentials"
// @Success 200 {object} HTTPResponse "Authentication successful"
// @Failure 400 {object} HTTPResponse "Invalid input data"
// @Failure 500 {object} HTTPResponse "Internal server error"
// @Router /auth [post]
//
// @Summary Delete a user
// @Description Delete a user by sending a DELETE request with session and user ID.
// @Tags User Management
// @Accept json
// @Produce json
// @Param session cookie string true "Session ID for authorization"
// @Param request body PersonDelete true "User credentials"
// @Success 200 {object} HTTPResponse "User successfully deleted"
// @Failure 400 {object} HTTPResponse "Invalid input data"
// @Failure 500 {object} HTTPResponse "Internal server error"
// @Router /del [delete]
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
			Status:  http.StatusBadGateway,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("[ERROR] [API-Service] [ProxyHTTP] [TraceID: %v] Failed to send response: %v", traceID, err)
		}
	}
	log.Printf("[ERROR] [API-Service] [ProxyHTTP] [TraceID: %v] Successful HTTP-request to %s", traceID, targetURL)
	proxy.ServeHTTP(c.Writer, c.Request)
}
func (h *Handler) ProxyGrpc(c *gin.Context) {

}
