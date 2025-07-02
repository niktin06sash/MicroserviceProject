package handlers

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/niktin06sash/MicroserviceProject/API_service/docs"
	pb "github.com/niktin06sash/MicroserviceProject/Photo_service/proto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	LoadPhoto(ctx context.Context, userid string, photodata []byte) (*pb.LoadPhotoResponse, error)
	DeletePhoto(ctx context.Context, userid string, photoid string) (*pb.DeletePhotoResponse, error)
	GetPhotos(ctx context.Context, userid string) (*pb.GetPhotosResponse, error)
	GetPhoto(ctx context.Context, userid string, photoid string) (*pb.GetPhotoResponse, error)
}

const Logout = "API-Logout"
const DeleteUser = "API-DeleteUser"
const Login = "API-Login"
const Registration = "API-Registration"
const Update = "API-Update"
const LoadPhoto = "API-LoadPhoto"
const DeletePhoto = "API-DeletePhoto"
const GetMyPhotoById = "API-GetMyPhotoById"
const GetPhotoById = "API-GetPhotoById"
const GetUserProfileById = "API-GetUserProfileById"
const GetMyProfile = "API-GetMyProfile"

type Handler struct {
	Middleware  Middleware
	Routes      map[string]string
	logproducer LogProducer
	photoclient PhotoClient
	httpclient  *http.Client
}

func NewHandler(middleware Middleware, photoclient PhotoClient, logproducer LogProducer, routes map[string]string) *Handler {
	return &Handler{
		Middleware:  middleware,
		Routes:      routes,
		logproducer: logproducer,
		photoclient: photoclient,
		httpclient: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   5 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				IdleConnTimeout:       30 * time.Second,
			},
		},
	}
}
func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.New()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.Use(h.Middleware.Logging())
	r.Use(h.Middleware.RateLimiter())
	r.POST("/api/auth/register", h.Middleware.AuthorizedNot(), h.Registration)
	r.POST("/api/auth/login", h.Middleware.AuthorizedNot(), h.Login)
	r.DELETE("/api/users/del", h.Middleware.Authorized(), h.DeleteUser)
	r.DELETE("/api/users/logout", h.Middleware.Authorized(), h.Logout)
	r.PATCH("/api/me/update", h.Middleware.Authorized(), h.Update)

	r.GET("/api/me", h.Middleware.Authorized(), h.GetMyProfile)
	r.GET("/api/users/:id", h.Middleware.Authorized(), h.GetUserProfileById)
	r.POST("/api/me/photos", h.Middleware.Authorized(), h.LoadPhoto)
	r.DELETE("/api/me/photos/:photo_id", h.Middleware.Authorized(), h.DeletePhoto)
	r.GET("/api/me/photos/:photo_id", h.Middleware.Authorized(), h.GetMyPhotoById)
	r.GET("/api/users/:id/photos/:photo_id", h.Middleware.Authorized(), h.GetPhotoById)
	return r
}
