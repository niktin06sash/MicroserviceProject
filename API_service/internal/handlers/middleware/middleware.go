package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/client"
	"golang.org/x/time/rate"
)

type RateLimiterEntry struct {
	Limiter  *rate.Limiter
	LastUsed time.Time
}
type Middleware struct {
	grpcClient   client.GrpcClientService
	rateLimiters sync.Map
	stopclean    chan (struct{})
}
type MiddlewareService interface {
	RateLimiter() gin.HandlerFunc
	Logging() gin.HandlerFunc
	Authorized() gin.HandlerFunc
	AuthorizedNot() gin.HandlerFunc
	Stop()
}

func NewMiddleware(grpcClient client.GrpcClientService) *Middleware {
	m := &Middleware{
		grpcClient:   grpcClient,
		rateLimiters: sync.Map{},
		stopclean:    make(chan struct{}),
	}
	go cleanLimit(m)
	return m
}
