package handlers

import (
	"net/http"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/middleware"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"

	"github.com/gorilla/mux"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}
func (h *Handler) InitRoutes() *mux.Router {
	m := mux.NewRouter()
	authNotGroup := m.PathPrefix("/").Subrouter()
	authNotGroup.Use(func(next http.Handler) http.Handler {
		return middleware.Middleware_Logging(middleware.Middleware_AuthorizedNot(next))
	})

	authGroup := m.PathPrefix("/").Subrouter()
	authGroup.Use(func(next http.Handler) http.Handler {
		return middleware.Middleware_Logging(middleware.Middleware_Authorized(next))
	})
	authNotGroup.HandleFunc("/reg", h.Registration).Methods("POST")
	authNotGroup.HandleFunc("/auth", h.Authentication).Methods("POST")
	authGroup.HandleFunc("/del", h.DeleteAccount).Methods("DELETE")
	return m
}
