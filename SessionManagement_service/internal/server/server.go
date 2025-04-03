package server

import (
	pb "SessionManagement_service/proto"
	"context"
	"log"
	"net"

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
		log.Fatalf("failed to listen: %v", err)
		return err
	}

	s.server = grpc.NewServer()
	pb.RegisterSessionServiceServer(s.server, s.service)

	log.Println("Starting gRPC server on port: " + port)
	if err := s.server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
		return err
	}
	return nil
}
func (s *GrpcServer) Shutdown(ctx context.Context) error {

	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Println("gRPC server gracefully stopped")
		return nil
	case <-ctx.Done():
		log.Println("Graceful shutdown timed out, forcefully stopping the server")
		s.server.Stop()
		return ctx.Err()
	}
}
