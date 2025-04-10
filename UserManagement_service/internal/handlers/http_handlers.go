package handlers

import (
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"

	"github.com/google/uuid"
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
	UserID        uuid.UUID         `json:"userid"`
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}
func (h *Handler) InitRoutes() *mux.Router {
	m := mux.NewRouter()
	m.HandleFunc("/reg", h.Registration).Methods("POST")
	m.HandleFunc("/auth", h.Authentication).Methods("POST")
	m.HandleFunc("/delete", h.DeleteAccount).Methods("POST")
	return m
}
