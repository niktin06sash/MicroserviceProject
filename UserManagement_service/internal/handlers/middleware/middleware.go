package middleware

import (
	"net/http"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
)

type Middleware struct {
	KafkaProducer kafka.KafkaProducerService
}

const Authority = "Middleware-Authority"
const Not_Authority = "Middleware-Not-Authority"
const Logging = "Middleware-Logging"

func NewMiddleware(kafka kafka.KafkaProducerService) *Middleware {
	return &Middleware{KafkaProducer: kafka}
}

type MiddlewareService interface {
	Logging(next http.Handler) http.Handler
	Authorized(next http.Handler) http.Handler
	AuthorizedNot(next http.Handler) http.Handler
}
