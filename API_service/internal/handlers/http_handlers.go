package handlers

import (
	"github.com/gin-gonic/gin"
	_ "github.com/niktin06sash/MicroserviceProject/API_service/docs"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Use(middleware.Middleware_Logging())
	r.POST("/api/reg", middleware.Middleware_AuthorizedNot(h.GRPCclient), h.Registration)
	r.POST("/api/auth", middleware.Middleware_AuthorizedNot(h.GRPCclient), h.Authentication)
	r.DELETE("/api/del", middleware.Middleware_Authorized(h.GRPCclient), h.DeleteUser)
	r.DELETE("/api/logout", middleware.Middleware_Authorized(h.GRPCclient), h.Logout)
	return r
}
