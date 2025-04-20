package handlers

import (
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/middleware"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"

	"github.com/gorilla/mux"
)

const (
	jsonResponseType = "application/json"
)

type Handler struct {
	services *service.Service
}
type HTTPResponse struct {
	Success       bool              `json:"success"`
	Errors        map[string]string `json:"errors"`
	SessionId     string            `json:"sessionid"`
	ExpireSession time.Time         `json:"expiresession"`
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}
func (h *Handler) InitRoutes() *mux.Router {
	m := mux.NewRouter()
	m.HandleFunc("/reg", middleware.Middleware_Logging(middleware.Middleware_Authorized(false)(middleware.Middleware_Retry(h.Registration)))).Methods("POST")
	m.HandleFunc("/auth", middleware.Middleware_Logging(middleware.Middleware_Authorized(false)(middleware.Middleware_Retry(h.Authentication)))).Methods("POST")
	m.HandleFunc("/delete", middleware.Middleware_Logging(middleware.Middleware_Authorized(true)(middleware.Middleware_Retry(h.DeleteAccount)))).Methods("POST")
	return m
}
