package handlers

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gorilla/mux"
)

type Handler struct {
	Services    UserService
	Middlewares MiddlewareService
	LogProducer LogProducer
}

const Registration = "API-Registration"
const Authentication = "API-Authentication"
const DeleteAccount = "API-DeleteAccount"
const Logout = "API-Logout"
const Update = "API-Update"
const MyProfile = "API-MyProfile"
const GetUserById = "API-GetUserById"
const MyFriends = "API-MyFriends"

type MiddlewareService interface {
	Logging(next http.Handler) http.Handler
	Authorized(next http.Handler) http.Handler
	AuthorizedNot(next http.Handler) http.Handler
}
type LogProducer interface {
	NewUserLog(level, place, traceid, msg string)
}

func NewHandler(services UserService, middleware MiddlewareService, logproducer LogProducer) *Handler {
	return &Handler{Services: services, Middlewares: middleware, LogProducer: logproducer}
}
func (h *Handler) InitRoutes() *mux.Router {
	m := mux.NewRouter()
	m.Handle("/metrics", promhttp.Handler())
	authNotGroup := m.PathPrefix("/").Subrouter()
	authNotGroup.Use(func(next http.Handler) http.Handler {
		return h.Middlewares.Logging(h.Middlewares.AuthorizedNot(next))
	})
	authGroup := m.PathPrefix("/").Subrouter()
	authGroup.Use(func(next http.Handler) http.Handler {
		return h.Middlewares.Logging(h.Middlewares.Authorized(next))
	})
	authNotGroup.HandleFunc("/auth/register", h.Registration).Methods("POST")
	authNotGroup.HandleFunc("/auth/login", h.Login).Methods("POST")
	authGroup.HandleFunc("/users/del", h.DeleteAccount).Methods("DELETE")
	authGroup.HandleFunc("/users/logout", h.Logout).Methods("DELETE")
	authGroup.HandleFunc("/me/update", h.Update).Methods("PATCH")
	authGroup.HandleFunc("/me", h.MyProfile).Methods("GET")
	authGroup.HandleFunc("/users/id/{id}", h.GetUserById).Methods("GET")
	return m
}
