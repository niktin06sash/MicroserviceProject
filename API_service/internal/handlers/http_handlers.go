package handlers

import (
	"github.com/gin-gonic/gin"
	_ "github.com/niktin06sash/MicroserviceProject/API_service/docs"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/middleware"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/kafka"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	Middleware    middleware.MiddlewareService
	Routes        map[string]string
	KafkaProducer kafka.KafkaProducerService
}

const ProxyHTTP = "API-ProxyHTTP"

func NewHandler(middleware middleware.MiddlewareService, kafkaproducer kafka.KafkaProducerService, routes map[string]string) *Handler {
	return &Handler{
		Middleware:    middleware,
		Routes:        routes,
		KafkaProducer: kafkaproducer,
	}
}
func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.New()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.Use(h.Middleware.Logging())
	r.Use(h.Middleware.RateLimiter())
	r.POST("/api/reg", h.Middleware.AuthorizedNot(), h.Registration)
	r.POST("/api/auth", h.Middleware.AuthorizedNot(), h.Authentication)
	r.DELETE("/api/del", h.Middleware.Authorized(), h.DeleteUser)
	r.DELETE("/api/logout", h.Middleware.Authorized(), h.Logout)
	r.PATCH("/api/me/update", h.Middleware.Authorized(), h.Update)
	r.GET("/api/me", h.Middleware.Authorized(), h.MyProfile)
	r.GET("/api/users/id/:id", h.Middleware.Authorized(), h.GetUserById)
	r.GET("/api/me/friends", h.Middleware.Authorized(), h.MyFriends)
	return r
}
