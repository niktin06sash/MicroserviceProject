package server

import (
	"context"
	"log"
	"net"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/configs"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"

	"google.golang.org/grpc"
)

type GrpcServer struct {
	server *grpc.Server
}

func NewGrpcServer(config configs.ServerConfig, service pb.SessionServiceServer) *GrpcServer {
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(config.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(config.MaxSendMsgSize),
	}
	s := &GrpcServer{
		server: grpc.NewServer(opts...),
	}
	pb.RegisterSessionServiceServer(s.server, service)
	return s
}
func (s *GrpcServer) Run(config configs.ServerConfig) error {
	lis, err := net.Listen("tcp", ":"+config.Port)
	if err != nil {
		log.Printf("[DEBUG] [Session-Service] Error create connection on port %s: %v", config.Port, err)
		return err
	}
	log.Printf("[DEBUG] [Session-Service] Starting gRPC-server on port: %s", config.Port)
	return s.server.Serve(lis)
}
func (s *GrpcServer) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
		log.Println("[DEBUG] [Session-Service] Starting gRPC-server gracefully stopped")
		return nil
	case <-ctx.Done():
		log.Println("[DEBUG] [Session-Service] Graceful shutdown timed out, forcefully stopping the server")
		s.server.Stop()
		return ctx.Err()
	}
}
