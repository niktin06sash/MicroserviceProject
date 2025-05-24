package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/kafka"
	"golang.org/x/time/rate"
)

type RateLimiterEntry struct {
	Limiter  *rate.Limiter
	LastUsed time.Time
}
type Middleware struct {
	grpcClient    client.GrpcClientService
	KafkaProducer kafka.KafkaProducerService
	rateLimiters  sync.Map
	stopclean     chan (struct{})
}
type MiddlewareService interface {
	RateLimiter() gin.HandlerFunc
	Logging() gin.HandlerFunc
	Authorized() gin.HandlerFunc
	AuthorizedNot() gin.HandlerFunc
}

const RateLimiter = "Middleware-RateLimiter"
const Not_Authority = "Middleware-Not-Authority"
const Authority = "Middleware-Authority"

func NewMiddleware(grpcClient client.GrpcClientService, kafkaProducer kafka.KafkaProducerService) *Middleware {
	m := &Middleware{
		grpcClient:    grpcClient,
		KafkaProducer: kafkaProducer,
		rateLimiters:  sync.Map{},
		stopclean:     make(chan struct{}),
	}
	go cleanLimit(m)
	return m
}
