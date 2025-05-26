package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
)

func (m *Middleware) Authorized() gin.HandlerFunc {
	return func(c *gin.Context) {
		const place = Authority
		maparesponse := make(map[string]string)
		traceID := c.MustGet("traceID").(string)
		sessionID, err := c.Cookie("session")
		if err != nil {
			m.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, "Required session in cookie")
			maparesponse[erro.ClientErrorType] = erro.RequiredSession
			response.SendResponse(c, http.StatusUnauthorized, response.HTTPResponse{Success: false, Errors: maparesponse}, traceID, place, m.KafkaProducer)
			c.Abort()
			metrics.APIErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
			return
		}
		grpcresponse, errv := retryAuthorized(c, m, sessionID, traceID, place)
		if errv != nil {
			switch errv.Type {
			case erro.ClientErrorType:
				maparesponse[erro.ClientErrorType] = errv.Message
				response.SendResponse(c, http.StatusUnauthorized, response.HTTPResponse{Success: false, Errors: maparesponse}, traceID, place, m.KafkaProducer)
				c.Abort()
				return

			case erro.ServerErrorType:
				maparesponse[erro.ServerErrorType] = errv.Message
				response.SendResponse(c, http.StatusInternalServerError, response.HTTPResponse{Success: false, Errors: maparesponse}, traceID, place, m.KafkaProducer)
				c.Abort()
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
		const place = Not_Authority
		maparesponse := make(map[string]string)
		traceID := c.MustGet("traceID").(string)
		sessionID, err := c.Cookie("session")
		if err != nil || sessionID == "" {
			m.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelInfo, place, traceID, "Successful unauthorization verification")
			c.Next()
			return
		}
		_, errv := retryAuthorized_Not(c, m, sessionID, traceID, place)
		if errv != nil {
			switch errv.Type {
			case erro.ClientErrorType:
				maparesponse[erro.ClientErrorType] = errv.Message
				response.SendResponse(c, http.StatusForbidden, response.HTTPResponse{Success: false, Errors: maparesponse}, traceID, place, m.KafkaProducer)
				c.Abort()
				return

			case erro.ServerErrorType:
				maparesponse[erro.ServerErrorType] = errv.Message
				response.SendResponse(c, http.StatusInternalServerError, response.HTTPResponse{Success: false, Errors: maparesponse}, traceID, place, m.KafkaProducer)
				c.Abort()
				return
			}
		}
		m.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelInfo, place, traceID, "Successful unauthorization verification")
		c.Next()
	}
}
