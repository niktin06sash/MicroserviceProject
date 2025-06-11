package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiterEntry struct {
	Limiter  *rate.Limiter
	LastUsed time.Time
}
type Middleware struct {
	grpcClient   SessionClient
	logproducer  LogProducer
	rateLimiters sync.Map
	stopclean    chan (struct{})
}
type LogProducer interface {
	NewAPILog(c *http.Request, level, place, traceid, msg string)
}

const RateLimiter = "Middleware-RateLimiter"
const Not_Authority = "Middleware-Not-Authority"
const Authority = "Middleware-Authority"

func NewMiddleware(grpcClient SessionClient, logproducer LogProducer) *Middleware {
	m := &Middleware{
		grpcClient:   grpcClient,
		logproducer:  logproducer,
		rateLimiters: sync.Map{},
		stopclean:    make(chan struct{}),
	}
	go cleanLimit(m)
	return m
}
