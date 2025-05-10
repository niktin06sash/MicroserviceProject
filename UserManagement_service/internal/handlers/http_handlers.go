package handlers

import (
	"net/http"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/middleware"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"

	"github.com/gorilla/mux"
)

type Handler struct {
	Services      *service.Service
	Middlewares   middleware.MiddlewareService
	KafkaProducer kafka.KafkaProducerService
}

func NewHandler(services *service.Service, middleware middleware.MiddlewareService, kafka kafka.KafkaProducerService) *Handler {
	return &Handler{Services: services, Middlewares: middleware, KafkaProducer: kafka}
}
func (h *Handler) InitRoutes() *mux.Router {
	m := mux.NewRouter()
	authNotGroup := m.PathPrefix("/").Subrouter()
	authNotGroup.Use(func(next http.Handler) http.Handler {
		return h.Middlewares.Logging(h.Middlewares.AuthorizedNot(next))
	})
	authGroup := m.PathPrefix("/").Subrouter()
	authGroup.Use(func(next http.Handler) http.Handler {
		return h.Middlewares.Logging(h.Middlewares.Authorized(next))
	})
	authNotGroup.HandleFunc("/reg", h.Registration).Methods("POST")
	authNotGroup.HandleFunc("/auth", h.Authentication).Methods("POST")
	authGroup.HandleFunc("/del", h.DeleteAccount).Methods("DELETE")
	authGroup.HandleFunc("/logout", h.Logout).Methods("DELETE")
	return m
}
