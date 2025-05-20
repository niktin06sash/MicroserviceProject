package server

import (
	"context"
	"log"
	"net"

	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"

	"google.golang.org/grpc"
)

type GrpcServer struct {
	server  *grpc.Server
	service pb.SessionServiceServer
}

func NewGrpcServer(service pb.SessionServiceServer) *GrpcServer {
	return &GrpcServer{service: service}
}
func (s *GrpcServer) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	s.server = grpc.NewServer()
	pb.RegisterSessionServiceServer(s.server, s.service)
	log.Println("[DEBUG] [Session-Service] Starting gRPC-server on port:", port)
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
