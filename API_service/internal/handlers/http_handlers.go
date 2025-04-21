package handlers

import (
	"github.com/gorilla/mux"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/middleware"
)

type Handler struct {
	APIMiddleware *middleware.APIMiddleware
}

func NewHandler(midleware *middleware.APIMiddleware) *Handler {
	return &Handler{APIMiddleware: midleware}
}
func (h *Handler) InitRoutes() *mux.Router {
	m := mux.NewRouter()
	m.HandleFunc("/reg", h.APIMiddleware.Middleware()).Methods("POST")
	m.HandleFunc("/auth", h.APIMiddleware.Middleware()).Methods("POST")
	m.HandleFunc("/delete", h.APIMiddleware.Middleware()).Methods("DELETE")
	return m
}
