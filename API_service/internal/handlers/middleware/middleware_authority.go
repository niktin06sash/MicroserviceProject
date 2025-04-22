package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/client"
	"google.golang.org/grpc/metadata"
)

func AuthorityMiddleware(grpcClient *client.GrpcClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.MustGet("traceID").(string)
		cookie, err := c.Cookie("session")
		if err != nil {
			logRequest(c.Request, "Authority", traceID, true, err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		sessionID := cookie
		md := metadata.Pairs("traceID", traceID)
		ctxGRPC := metadata.NewOutgoingContext(c.Request.Context(), md)
		response, err := grpcClient.ValidateSession(ctxGRPC, sessionID)
		if err != nil || !response.Success {
			logRequest(c.Request, "Authority", traceID, true, err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Set("userID", response.UserID)
		c.Set("sessionID", sessionID)
		c.Next()
	}
}

func NotAuthorityMiddleware(grpcClient *client.GrpcClient) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		response, err := grpcClient.ValidateSession(ctxGRPC, sessionID)
		if err == nil || response.Success {
			logRequest(c.Request, "Not-Authority", traceID, true, err.Error())
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		c.Next()
	}
}
