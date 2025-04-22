package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/middleware"
)

type Handler struct {
	GRPCclient *client.GrpcClient
}

func NewHandler(grpc *client.GrpcClient) *Handler {
	return &Handler{
		GRPCclient: grpc,
	}
}
func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.LoggingMiddleware())
	r.POST("/reg", middleware.NotAuthorityMiddleware(h.GRPCclient), h.RegistrationHandler)
	r.POST("/login", middleware.AuthorityMiddleware(h.GRPCclient), h.LoginHandler)
	return r
}
