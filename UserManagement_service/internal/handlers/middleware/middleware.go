package middleware

import (
	"net/http"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
)

type Middleware struct {
	KafkaProducer kafka.KafkaProducerService
}

func NewMiddleware(kafka kafka.KafkaProducerService) *Middleware {
	return &Middleware{KafkaProducer: kafka}
}

type MiddlewareService interface {
	Logging(next http.Handler) http.Handler
	Authorized(next http.Handler) http.Handler
	AuthorizedNot(next http.Handler) http.Handler
}
