package middleware

type Middleware struct {
	KafkaProducer LogProducer
}

const Authority = "Middleware-Authority"
const Not_Authority = "Middleware-Not-Authority"
const Logging = "Middleware-Logging"

func NewMiddleware(kafka LogProducer) *Middleware {
	return &Middleware{KafkaProducer: kafka}
}

type LogProducer interface {
	NewUserLog(level, place, traceid, msg string)
}
