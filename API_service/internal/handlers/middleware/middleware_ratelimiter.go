package middleware

import (
	"log"
	"net/http"
	"time"

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
func (m *Middleware) Stop() {
	close(m.stopclean)
}
func getLimit(m *Middleware, ip string) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()
	if entry, exist := m.rateLimiters[ip]; exist {
		entry.LastUsed = time.Now()
		return entry.Limiter
	}
	limiter := rate.NewLimiter(5, 3)
	m.rateLimiters[ip] = &RateLimiterEntry{
		Limiter:  limiter,
		LastUsed: time.Now(),
	}
	return limiter
}
func cleanLimit(m *Middleware) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-m.stopclean:
			log.Println("[INFO] [API-Service] [RateLimiter] Successful completion of RateLimiter")
			return
		case <-ticker.C:
			m.mu.Lock()
			defer m.mu.Unlock()
			for ip, entry := range m.rateLimiters {
				if time.Since(entry.LastUsed) >= 5*time.Second {
					delete(m.rateLimiters, ip)
				}
			}
		}
	}
}
