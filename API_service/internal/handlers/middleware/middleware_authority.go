package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"google.golang.org/grpc/metadata"
)

func AuthorityMiddleware(grpcClient *client.GrpcClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		maparesponse := make(map[string]string)
		traceID := c.MustGet("traceID").(string)
		cookie, err := c.Cookie("session")
		if err != nil {
			logRequest(c.Request, "Authority", traceID, true, err.Error())
			maparesponse["ClientError"] = "Required Session in Cookie"
			response.SendResponse(c, http.StatusUnauthorized, false, nil, maparesponse)
			return
		}

		sessionID := cookie
		md := metadata.Pairs("traceID", traceID)
		ctxGRPC := metadata.NewOutgoingContext(c.Request.Context(), md)
		grpcresponse, err := grpcClient.ValidateSession(ctxGRPC, sessionID)
		if err != nil || !grpcresponse.Success {
			logRequest(c.Request, "Authority", traceID, true, err.Error())
			maparesponse["ClientError"] = "Invalid Session Data!"
			response.SendResponse(c, http.StatusUnauthorized, false, nil, maparesponse)
			return
		}
		c.Set("userID", grpcresponse.UserID)
		c.Set("sessionID", sessionID)
		c.Next()
	}
}

func NotAuthorityMiddleware(grpcClient *client.GrpcClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		maparesponse := make(map[string]string)
		traceID := c.MustGet("traceID").(string)
		cookie, err := c.Cookie("session")
		if err != nil {
			logRequest(c.Request, "Not-Authority", traceID, false, "")
			c.Next()
			return
		}

		sessionID := cookie
		md := metadata.Pairs("traceID", traceID)
		ctxGRPC := metadata.NewOutgoingContext(c.Request.Context(), md)
		grpcresponse, err := grpcClient.ValidateSession(ctxGRPC, sessionID)
		if err == nil || grpcresponse.Success {
			logRequest(c.Request, "Not-Authority", traceID, true, err.Error())
			maparesponse["ClientError"] = "Invalid Session Data!"
			response.SendResponse(c, http.StatusForbidden, false, nil, maparesponse)
			return
		}

		c.Next()
	}
}
