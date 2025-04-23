package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/middleware"
)

type Handler struct {
	GRPCclient client.GrpcClientService
	Routes     map[string]string
}

func NewHandler(grpc *client.GrpcClient, routes map[string]string) *Handler {
	return &Handler{
		GRPCclient: grpc,
		Routes:     routes,
	}
}
func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.New()
	r.Use(middleware.Middleware_Logging())
	r.POST("/reg", middleware.Middleware_AuthorizedNot(h.GRPCclient), h.ProxyHTTP)
	r.POST("/auth", middleware.Middleware_AuthorizedNot(h.GRPCclient), h.ProxyHTTP)
	r.DELETE("/del", middleware.Middleware_Authorized(h.GRPCclient), h.ProxyHTTP)
	r.POST("/logout", middleware.Middleware_Authorized(h.GRPCclient), h.ProxtGrpc)
	return r
}
