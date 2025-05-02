package handlers

import (
	"github.com/gin-gonic/gin"
	_ "github.com/niktin06sash/MicroserviceProject/API_service/docs"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	Middleware middleware.MiddlewareService
	Routes     map[string]string
}

func NewHandler(middleware middleware.MiddlewareService, routes map[string]string) *Handler {
	return &Handler{
		Middleware: middleware,
		Routes:     routes,
	}
}
func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.New()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Use(h.Middleware.Logging())
	r.Use(h.Middleware.RateLimiter())
	r.POST("/api/reg", h.Middleware.AuthorizedNot(), h.Registration)
	r.POST("/api/auth", h.Middleware.AuthorizedNot(), h.Authentication)
	r.DELETE("/api/del", h.Middleware.Authorized(), h.DeleteUser)
	r.DELETE("/api/logout", h.Middleware.Authorized(), h.Logout)
	return r
}
