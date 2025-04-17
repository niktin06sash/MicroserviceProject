package server

import (
	"context"
	"log"
	"net"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/logger"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"go.uber.org/zap"

	"google.golang.org/grpc"
)

type GrpcServer struct {
	logger  *logger.SessionLogger
	server  *grpc.Server
	service pb.SessionServiceServer
}

func NewGrpcServer(service pb.SessionServiceServer, log *logger.SessionLogger) *GrpcServer {
	return &GrpcServer{service: service, logger: log}
}
func (s *GrpcServer) Run(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return err
	}

	s.server = grpc.NewServer()
	pb.RegisterSessionServiceServer(s.server, s.service)
	s.logger.Info("SessionManagement: Starting gRPC server on port", zap.String("port", port))
	if err := s.server.Serve(lis); err != nil {
		s.logger.Fatal("SessionManagement: Failed to serve", zap.Error(err))
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
		s.logger.Info("SessionManagement: gRPC server gracefully stopped")
		return nil
	case <-ctx.Done():
		s.logger.Info("SessionManagement: Graceful shutdown timed out, forcefully stopping the server")
		s.server.Stop()
		return ctx.Err()
	}
}
