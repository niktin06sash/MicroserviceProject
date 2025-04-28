package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/client"
	"golang.org/x/time/rate"
)

type Middleware struct {
	grpcClient   client.GrpcClientService
	rateLimiters map[string]*rate.Limiter
	mu           sync.Mutex
}
type MiddlewareService interface {
	RateLimiter() gin.HandlerFunc
	Logging() gin.HandlerFunc
	Authorized() gin.HandlerFunc
	AuthorizedNot() gin.HandlerFunc
}

func NewMiddleware(grpcClient client.GrpcClientService) *Middleware {
	return &Middleware{
		grpcClient:   grpcClient,
		rateLimiters: make(map[string]*rate.Limiter),
	}
}
