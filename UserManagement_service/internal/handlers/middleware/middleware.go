package middleware

type Middleware struct {
	LogProducer LogProducer
}

const Authority = "Middleware-Authority"
const Not_Authority = "Middleware-Not-Authority"
const Logging = "Middleware-Logging"

func NewMiddleware(logproducer LogProducer) *Middleware {
	return &Middleware{LogProducer: logproducer}
}

type LogProducer interface {
	NewUserLog(level, place, traceid, msg string)
}
