package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
	"golang.org/x/time/rate"
)

func (m *Middleware) RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		var place = "Middleware-RateLimiter"
		traceID := c.MustGet("traceID").(string)
		ip := c.Request.RemoteAddr
		limiter := getLimit(m, ip)
		if !limiter.Allow() {
			m.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, "Too many requests")
			response.SendResponse(c, http.StatusTooManyRequests, false, nil, map[string]string{"ClientError": "Too Many Requests"}, traceID, place, m.KafkaProducer)
			c.Abort()
			metrics.APIErrorsTotal.WithLabelValues("ClientError").Inc()
			metrics.APIRateLimitExceededTotal.WithLabelValues(c.Request.URL.Path).Inc()
			return
		}
		c.Next()
	}
}
func (m *Middleware) Stop() {
	close(m.stopclean)
}
func getLimit(m *Middleware, ip string) *rate.Limiter {
	if entry, exist := m.rateLimiters.Load(ip); exist {
		e := entry.(*RateLimiterEntry)
		return e.Limiter
	}
	limiter := rate.NewLimiter(0.25, 5)
	newEntry := &RateLimiterEntry{
		Limiter:  limiter,
		LastUsed: time.Now(),
	}
	m.rateLimiters.Store(ip, newEntry)
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
			log.Println("[INFO] [API-Service] [RateLimiter] Successful cleaning has started...")
			m.rateLimiters.Range(func(key, value any) bool {
				ip := key.(string)
				entry := value.(*RateLimiterEntry)
				if time.Since(entry.LastUsed) >= 5*time.Minute {
					m.rateLimiters.Delete(ip)
					log.Printf("[INFO] [API-Service] [RateLimiter] Deleted IP: %s", ip)
				}
				return true
			})
			log.Println("[INFO] [API-Service] [RateLimiter] Successful cleaning!")
		}
	}
}
