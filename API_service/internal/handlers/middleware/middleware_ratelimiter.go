package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"golang.org/x/time/rate"
)

func (m *Middleware) RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.MustGet("traceID").(string)
		ip := c.Request.RemoteAddr
		limiter := getLimit(m, ip)
		if !limiter.Allow() {
			logRequest(c.Request, "RateLimiter", traceID, false, "Too many requests")
			response.SendResponse(c, http.StatusTooManyRequests, false, nil, map[string]string{"ClientError": "Too Many Requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func getLimit(m *Middleware, ip string) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()
	if limiter, exist := m.rateLimiters[ip]; exist {
		return limiter
	}
	limiter := rate.NewLimiter(5, 3)
	m.rateLimiters[ip] = limiter
	return limiter
}
