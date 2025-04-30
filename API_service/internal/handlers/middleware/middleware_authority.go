package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
)

func (m *Middleware) Authorized() gin.HandlerFunc {
	return func(c *gin.Context) {
		maparesponse := make(map[string]string)
		traceID := c.MustGet("traceID").(string)
		sessionID, err := c.Cookie("session")
		if err != nil {
			logRequest(c.Request, "Authority", traceID, true, err.Error())
			maparesponse["ClientError"] = "Required Session in Cookie"
			response.SendResponse(c, http.StatusUnauthorized, false, nil, maparesponse, traceID, "Authority")
			c.Abort()
			return
		}
		grpcresponse, errv := retryAuthorized(c, m.grpcClient, sessionID, traceID, "Authority")
		if errv != nil {
			switch errv.GetTypeError() {
			case erro.ClientErrorType:
				logRequest(c.Request, "Authority", traceID, true, "Unauthorized-request for authorized users")
				maparesponse["ClientError"] = errv.Error()
				response.SendResponse(c, http.StatusUnauthorized, false, nil, maparesponse, traceID, "Authority")
				c.Abort()
				return

			case erro.ServerErrorType:
				logRequest(c.Request, "Authority", traceID, true, "Internal server error during authorization")
				maparesponse["InternalServerError"] = errv.Error()
				response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse, traceID, "Authority")
				c.Abort()
				return
			}
		}
		logRequest(c.Request, "Authority", traceID, false, "Successful authorization verification")
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
			response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse, traceID, "Not-Authority")
			c.Abort()
			return
		}
		sessionID, err := c.Cookie("session")
		if err != nil {
			logRequest(c.Request, "Not-Authority", traceID, false, "Successful unauthorization verification")
			c.Next()
			return
		}
		_, errv := retryAuthorized_Not(c, m.grpcClient, sessionID, traceID, "Not-Authority")
		if errv != nil {
			switch errv.GetTypeError() {
			case erro.ClientErrorType:
				logRequest(c.Request, "Authority", traceID, true, "Authorized-request for unauthorized users")
				maparesponse["ClientError"] = errv.Error()
				response.SendResponse(c, http.StatusForbidden, false, nil, maparesponse, traceID, "Not-Authority")
				c.Abort()
				return

			case erro.ServerErrorType:
				logRequest(c.Request, "Authority", traceID, true, "Internal server error during authorization")
				maparesponse["InternalServerError"] = errv.Error()
				response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse, traceID, "Not-Authority")
				c.Abort()
				return
			}
		}
		logRequest(c.Request, "Not-Authority", traceID, false, "Successful unauthorization verification")
		c.Next()
	}
}
