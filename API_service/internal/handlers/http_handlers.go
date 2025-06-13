package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/niktin06sash/MicroserviceProject/API_service/docs"
	pb "github.com/niktin06sash/MicroserviceProject/Photo_service/proto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
)

type Middleware interface {
	RateLimiter() gin.HandlerFunc
	Logging() gin.HandlerFunc
	Authorized() gin.HandlerFunc
	AuthorizedNot() gin.HandlerFunc
}
type LogProducer interface {
	NewAPILog(c *http.Request, level, place, traceid, msg string)
}
type PhotoClient interface {
	LoadPhoto(ctx context.Context, in *pb.LoadPhotoRequest, opts ...grpc.CallOption) (*pb.LoadPhotoResponse, error)
	DeletePhoto(ctx context.Context, in *pb.DeletePhotoRequest, opts ...grpc.CallOption) (*pb.DeletePhotoResponse, error)
	GetPhotos(ctx context.Context, in *pb.GetPhotosRequest, opts ...grpc.CallOption) (*pb.GetPhotosResponse, error)
	GetPhoto(ctx context.Context, in *pb.GetPhotoRequest, opts ...grpc.CallOption) (*pb.GetPhotoResponse, error)
}

const ProxyHTTP = "API-ProxyHTTP"

type Handler struct {
	Middleware  Middleware
	Routes      map[string]string
	logproducer LogProducer
	photoclient PhotoClient
}

func NewHandler(middleware Middleware, photoclient PhotoClient, logproducer LogProducer, routes map[string]string) *Handler {
	return &Handler{
		Middleware:  middleware,
		Routes:      routes,
		logproducer: logproducer,
		photoclient: photoclient,
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
	return r
}
