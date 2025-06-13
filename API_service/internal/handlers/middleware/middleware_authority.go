package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
)

type SessionClient interface {
	ValidateSession(ctx context.Context, sessionid string) (*pb.ValidateSessionResponse, error)
}

func (m *Middleware) Authorized() gin.HandlerFunc {
	return func(c *gin.Context) {
		const place = Authority
		traceID := c.MustGet("traceID").(string)
		sessionID, err := c.Cookie("session")
		if err != nil {
			m.logproducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, "Required session in cookie")
			response.BadResponse(c, http.StatusUnauthorized, erro.RequiredSession, traceID, place, m.logproducer)
			c.Abort()
			metrics.APIErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
			return
		}
		grpcresponse, errmap := retryAuthorized(c, m, sessionID, traceID, place)
		if errmap != nil {
			switch errmap[erro.ErrorType] {
			case erro.ClientErrorType:
				response.BadResponse(c, http.StatusUnauthorized, errmap[erro.ErrorMessage], traceID, place, m.logproducer)
				c.Abort()
				return

			case erro.ServerErrorType:
				response.BadResponse(c, http.StatusInternalServerError, errmap[erro.ErrorMessage], traceID, place, m.logproducer)
				c.Abort()
				return
			}
		}
		m.logproducer.NewAPILog(c.Request, kafka.LogLevelInfo, place, traceID, "Successful authorization verification")
		c.Set("userID", grpcresponse.UserID)
		c.Set("sessionID", sessionID)
		c.Next()
	}
}

func (m *Middleware) AuthorizedNot() gin.HandlerFunc {
	return func(c *gin.Context) {
		const place = Not_Authority
		traceID := c.MustGet("traceID").(string)
		sessionID, err := c.Cookie("session")
		if err != nil || sessionID == "" {
			m.logproducer.NewAPILog(c.Request, kafka.LogLevelInfo, place, traceID, "Successful unauthorization verification")
			c.Next()
			return
		}
		_, errmap := retryAuthorized_Not(c, m, sessionID, traceID, place)
		if errmap != nil {
			switch errmap[erro.ErrorType] {
			case erro.ClientErrorType:
				response.BadResponse(c, http.StatusForbidden, errmap[erro.ErrorMessage], traceID, place, m.logproducer)
				c.Abort()
				return

			case erro.ServerErrorType:
				response.BadResponse(c, http.StatusInternalServerError, errmap[erro.ErrorMessage], traceID, place, m.logproducer)
				c.Abort()
				return
			}
		}
		m.logproducer.NewAPILog(c.Request, kafka.LogLevelInfo, place, traceID, "Successful unauthorization verification")
		c.Next()
	}
}
