package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
)

func (m *Middleware) Authorized() gin.HandlerFunc {
	return func(c *gin.Context) {
		var place = "Authority"
		maparesponse := make(map[string]string)
		traceID := c.MustGet("traceID").(string)
		start := c.MustGet("starttime").(time.Time)
		sessionID, err := c.Cookie("session")
		if err != nil {
			m.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, "Required session in cookie")
			maparesponse["ClientError"] = "Required session in cookie"
			response.SendResponse(c, http.StatusUnauthorized, false, nil, maparesponse, traceID, place, m.KafkaProducer)
			c.Abort()
			duration := time.Since(start).Seconds()
			metrics.APIErrorsTotal.WithLabelValues("ClientError").Inc()
			metrics.APIRequestDuration.WithLabelValues(place).Observe(duration)
			return
		}
		grpcresponse, errv := retryAuthorized(c, m, sessionID, traceID, place)
		if errv != nil {
			switch errv.GetTypeError() {
			case erro.ClientErrorType:
				m.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, "Unauthorized-request for authorized users")
				maparesponse["ClientError"] = errv.Error()
				response.SendResponse(c, http.StatusUnauthorized, false, nil, maparesponse, traceID, place, m.KafkaProducer)
				c.Abort()
				duration := time.Since(start).Seconds()
				metrics.APIErrorsTotal.WithLabelValues("ClientError").Inc()
				metrics.APIRequestDuration.WithLabelValues(place).Observe(duration)
				metrics.APIBackendRequestsTotal.WithLabelValues("Session-Service").Inc()
				return

			case erro.ServerErrorType:
				maparesponse["InternalServerError"] = errv.Error()
				response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse, traceID, "Authority", m.KafkaProducer)
				c.Abort()
				duration := time.Since(start).Seconds()
				metrics.APIErrorsTotal.WithLabelValues("InternalServerError").Inc()
				metrics.APIRequestDuration.WithLabelValues(place).Observe(duration)
				metrics.APIBackendRequestsTotal.WithLabelValues("Session-Service").Inc()
				return
			}
		}
		m.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelInfo, place, traceID, "Successful authorization verification")
		c.Set("userID", grpcresponse.UserID)
		c.Set("sessionID", sessionID)
		c.Next()
	}
}

func (m *Middleware) AuthorizedNot() gin.HandlerFunc {
	return func(c *gin.Context) {
		var place = "Not-Authority"
		maparesponse := make(map[string]string)
		traceID := c.MustGet("traceID").(string)
		start := c.MustGet("starttime").(time.Time)
		sessionID, err := c.Cookie("session")
		if err != nil || sessionID == "" {
			m.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelInfo, place, traceID, "Successful unauthorization verification")
			c.Next()
			return
		}
		_, errv := retryAuthorized_Not(c, m, sessionID, traceID, place)
		if errv != nil {
			switch errv.GetTypeError() {
			case erro.ClientErrorType:
				m.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, "Authorized-request for unauthorized users")
				maparesponse["ClientError"] = errv.Error()
				response.SendResponse(c, http.StatusForbidden, false, nil, maparesponse, traceID, place, m.KafkaProducer)
				duration := time.Since(start).Seconds()
				metrics.APIErrorsTotal.WithLabelValues("ClientError").Inc()
				metrics.APIRequestDuration.WithLabelValues(place).Observe(duration)
				metrics.APIBackendRequestsTotal.WithLabelValues("Session-Service").Inc()
				c.Abort()
				return

			case erro.ServerErrorType:
				maparesponse["InternalServerError"] = errv.Error()
				response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse, traceID, place, m.KafkaProducer)
				duration := time.Since(start).Seconds()
				metrics.APIErrorsTotal.WithLabelValues("InternalServerError").Inc()
				metrics.APIRequestDuration.WithLabelValues(place).Observe(duration)
				metrics.APIBackendRequestsTotal.WithLabelValues("Session-Service").Inc()
				c.Abort()
				return
			}
		}
		m.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelInfo, place, traceID, "Successful unauthorization verification")
		c.Next()
	}
}
