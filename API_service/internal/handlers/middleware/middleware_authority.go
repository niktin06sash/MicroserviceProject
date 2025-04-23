package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"google.golang.org/grpc/metadata"
)

func Middleware_Authorized(grpcClient client.GrpcClientService) gin.HandlerFunc {
	return func(c *gin.Context) {
		maparesponse := make(map[string]string)
		traceID := c.MustGet("traceID").(string)
		cookie, err := c.Cookie("session")
		if err != nil {
			logRequest(c.Request, "Authority", traceID, true, err.Error())
			maparesponse["ClientError"] = "Required Session in Cookie"
			response.SendResponse(c, http.StatusUnauthorized, false, nil, maparesponse)
			c.Abort()
			return
		}

		sessionID := cookie
		md := metadata.Pairs("traceID", traceID)
		ctx := metadata.NewOutgoingContext(c.Request.Context(), md)
		err = response.CheckContext(c.Request.Context(), traceID, "Authority")
		if err != nil {
			maparesponse["InternalServerError"] = "Context canceled or deadline exceeded"
			response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse)
			c.Abort()
			return
		}
		grpcresponse, err := grpcClient.ValidateSession(ctx, sessionID)
		if err != nil || !grpcresponse.Success {
			logRequest(c.Request, "Authority", traceID, true, err.Error())
			maparesponse["ClientError"] = "Invalid Session Data!"
			response.SendResponse(c, http.StatusUnauthorized, false, nil, maparesponse)
			c.Abort()
			return
		}
		logRequest(c.Request, "Authority", traceID, false, "Successful authorization verification")
		c.Set("userID", grpcresponse.UserID)
		c.Set("sessionID", sessionID)
		c.Next()
	}
}

func Middleware_AuthorizedNot(grpcClient client.GrpcClientService) gin.HandlerFunc {
	return func(c *gin.Context) {
		maparesponse := make(map[string]string)
		traceID := c.MustGet("traceID").(string)
		err := response.CheckContext(c.Request.Context(), traceID, "Not-Authority")
		if err != nil {
			maparesponse["InternalServerError"] = "Context canceled or deadline exceeded"
			response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse)
			c.Abort()
			return
		}
		cookie, err := c.Cookie("session")
		if err != nil {
			logRequest(c.Request, "Not-Authority", traceID, false, "Successful unauthorization verification")
			c.Next()
			return
		}

		sessionID := cookie
		md := metadata.Pairs("traceID", traceID)
		ctx := metadata.NewOutgoingContext(c.Request.Context(), md)
		err = response.CheckContext(c.Request.Context(), traceID, "Not-Authority")
		if err != nil {
			maparesponse["InternalServerError"] = "Context canceled or deadline exceeded"
			response.SendResponse(c, http.StatusInternalServerError, false, nil, maparesponse)
			c.Abort()
			return
		}
		grpcresponse, err := grpcClient.ValidateSession(ctx, sessionID)
		if err == nil || grpcresponse.Success {
			logRequest(c.Request, "Not-Authority", traceID, true, err.Error())
			maparesponse["ClientError"] = "Invalid Session Data!"
			response.SendResponse(c, http.StatusForbidden, false, nil, maparesponse)
			c.Abort()
			return
		}
		logRequest(c.Request, "Not-Authority", traceID, false, "Successful unauthorization verification")
		c.Next()
	}
}
