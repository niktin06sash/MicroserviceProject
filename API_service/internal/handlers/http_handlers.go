package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/middleware"
)

type Handler struct {
	GRPCclient *client.GrpcClient
	Routes     map[string]string
}

func NewHandler(grpc *client.GrpcClient, routes map[string]string) *Handler {
	return &Handler{
		GRPCclient: grpc,
		Routes:     routes,
	}
}
func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.LoggingMiddleware())
	r.POST("/reg", middleware.NotAuthorityMiddleware(h.GRPCclient), h.ProxyHTTP)
	r.POST("/auth", middleware.NotAuthorityMiddleware(h.GRPCclient), h.ProxyHTTP)
	r.DELETE("/del", middleware.NotAuthorityMiddleware(h.GRPCclient), h.ProxyHTTP)
	r.POST("/logout", middleware.AuthorityMiddleware(h.GRPCclient), h.ProxtGrpc)
	return r
}
