package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/kafka"
)

func (m *Middleware) Authorized() gin.HandlerFunc {
	return func(c *gin.Context) {
		maparesponse := make(map[string]string)
		traceID := c.MustGet("traceID").(string)
		sessionID, err := c.Cookie("session")
		if err != nil {
			m.KafkaProducer.NewAPILog(kafka.APILog{
				Level:     kafka.LogLevelWarn,
				Place:     "Authority",
				TraceID:   traceID,
				IP:        c.Request.RemoteAddr,
				Method:    c.Request.Method,
				Path:      c.Request.URL.Path,
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   err.Error(),
			})
			maparesponse["ClientError"] = "Required Session in Cookie"
			response.SendResponse(c, http.StatusUnauthorized, false, nil, maparesponse, traceID, "Authority", m.KafkaProducer)
			c.Abort()
			return
		}
		grpcresponse, errv := retryAuthorized(c, m, sessionID, traceID, "Authority")
		if errv != nil {
			switch errv.GetTypeError() {
			case erro.ClientErrorType:
				m.KafkaProducer.NewAPILog(kafka.APILog{
					Level:     kafka.LogLevelWarn,
					Place:     "Authority",
					TraceID:   traceID,
					IP:        c.Request.RemoteAddr,
					Method:    c.Request.Method,
					Path:      c.Request.URL.Path,
					Timestamp: time.Now().Format(time.RFC3339),
					Message:   "Unauthorized-request for authorized users",
				})
				maparesponse["ClientError"] = errv.Error()
				response.SendResponse(c, http.StatusUnauthorized, false, nil, maparesponse, traceID, "Authority", m.KafkaProducer)
				c.Abort()
				return

			case erro.ServerErrorType:
				m.KafkaProducer.NewAPILog(kafka.APILog{
					Level:     kafka.LogLevelError,
					Place:     "Authority",
					TraceID:   traceID,
					IP:        c.Request.RemoteAddr,
					Method:    c.Request.Method,
					Path:      c.Request.URL.Path,
					Timestamp: time.Now().Format(time.RFC3339),
					Message:   "Internal server error during authorization",
				})
				maparesponse["InternalServerError"] = errv.Error()
				response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse, traceID, "Authority", m.KafkaProducer)
				c.Abort()
				return
			}
		}
		m.KafkaProducer.NewAPILog(kafka.APILog{
			Level:     kafka.LogLevelInfo,
			Place:     "Authority",
			TraceID:   traceID,
			IP:        c.Request.RemoteAddr,
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Timestamp: time.Now().Format(time.RFC3339),
			Message:   "Successful authorization verification",
		})
		c.Set("userID", grpcresponse.UserID)
		c.Set("sessionID", sessionID)
		c.Next()
	}
}

func (m *Middleware) AuthorizedNot() gin.HandlerFunc {
	return func(c *gin.Context) {
		maparesponse := make(map[string]string)
		traceID := c.MustGet("traceID").(string)
		err := response.CheckContext(c.Request.Context(), traceID, "Not-Authority")
		if err != nil {
			maparesponse["InternalServerError"] = "Context canceled or deadline exceeded"
			response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse, traceID, "Not-Authority", m.KafkaProducer)
			c.Abort()
			return
		}
		sessionID, err := c.Cookie("session")
		if err != nil {
			m.KafkaProducer.NewAPILog(kafka.APILog{
				Level:     kafka.LogLevelInfo,
				Place:     "Not-Authority",
				TraceID:   traceID,
				IP:        c.Request.RemoteAddr,
				Method:    c.Request.Method,
				Path:      c.Request.URL.Path,
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Successful unauthorization verification",
			})
			c.Next()
			return
		}
		_, errv := retryAuthorized_Not(c, m, sessionID, traceID, "Not-Authority")
		if errv != nil {
			switch errv.GetTypeError() {
			case erro.ClientErrorType:
				m.KafkaProducer.NewAPILog(kafka.APILog{
					Level:     kafka.LogLevelWarn,
					Place:     "Not-Authority",
					TraceID:   traceID,
					IP:        c.Request.RemoteAddr,
					Method:    c.Request.Method,
					Path:      c.Request.URL.Path,
					Timestamp: time.Now().Format(time.RFC3339),
					Message:   "Authorized-request for unauthorized users",
				})
				maparesponse["ClientError"] = errv.Error()
				response.SendResponse(c, http.StatusForbidden, false, nil, maparesponse, traceID, "Not-Authority", m.KafkaProducer)
				c.Abort()
				return

			case erro.ServerErrorType:
				m.KafkaProducer.NewAPILog(kafka.APILog{
					Level:     kafka.LogLevelError,
					Place:     "Not-Authority",
					TraceID:   traceID,
					IP:        c.Request.RemoteAddr,
					Method:    c.Request.Method,
					Path:      c.Request.URL.Path,
					Timestamp: time.Now().Format(time.RFC3339),
					Message:   "Internal server error during authorization",
				})
				maparesponse["InternalServerError"] = errv.Error()
				response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse, traceID, "Not-Authority", m.KafkaProducer)
				c.Abort()
				return
			}
		}
		m.KafkaProducer.NewAPILog(kafka.APILog{
			Level:     kafka.LogLevelInfo,
			Place:     "Not-Authority",
			TraceID:   traceID,
			IP:        c.Request.RemoteAddr,
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Timestamp: time.Now().Format(time.RFC3339),
			Message:   "Successful unauthorization verification",
		})
		c.Next()
	}
}
