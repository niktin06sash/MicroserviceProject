package server

import (
	"context"
	"log"
	"net/http"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(config configs.ServerConfig, handler http.Handler) *Server {
	server := &Server{}
	server.httpServer = &http.Server{
		Addr:           ":" + config.Port,
		Handler:        handler,
		MaxHeaderBytes: config.MaxHeaderBytes,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
	}
	return server
}
func (s *Server) Run() error {
	log.Printf("[DEBUG] [User-Service] Starting server on port: %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
