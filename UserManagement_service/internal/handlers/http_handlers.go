package handlers

import (
	"context"
	"net/http"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gorilla/mux"
)

const Registration = "API-Registration"
const Login = "API-Login"
const DeleteAccount = "API-DeleteAccount"
const Logout = "API-Logout"
const Update = "API-Update"
const MyProfile = "API-MyProfile"
const GetUserById = "API-GetUserById"
const MyFriends = "API-MyFriends"

const (
	userID      = "userID"
	sessionID   = "sessionID"
	getID       = "getID"
	update_type = "update_type"
)

type MiddlewareService interface {
	Logging(next http.Handler) http.Handler
	Authorized(next http.Handler) http.Handler
	AuthorizedNot(next http.Handler) http.Handler
}
type LogProducer interface {
	NewUserLog(level, place, traceid, msg string)
}
type UserService interface {
	RegistrateAndLogin(ctx context.Context, req *model.RegistrationRequest) *service.ServiceResponse
	AuthenticateAndLogin(ctx context.Context, req *model.AuthenticationRequest) *service.ServiceResponse
	DeleteAccount(ctx context.Context, req *model.DeletionRequest, sessionID string, useridstr string) *service.ServiceResponse
	Logout(ctx context.Context, sessionID string) *service.ServiceResponse
	UpdateAccount(ctx context.Context, req *model.UpdateRequest, useridstr string, updateType string) *service.ServiceResponse
	GetProfileById(ctx context.Context, getid string) *service.ServiceResponse
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
