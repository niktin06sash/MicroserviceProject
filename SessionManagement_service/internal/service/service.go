package service

import (
	"context"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
)

type Service struct {
	pb.UnimplementedSessionServiceServer
	sessionService *SessionService
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		sessionService: NewSessionService(repos.RedisSessionRepos),
	}
}

func (s *Service) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	return s.sessionService.CreateSession(ctx, req)
}

func (s *Service) ValidateSession(ctx context.Context, req *pb.ValidateSessionRequest) (*pb.ValidateSessionResponse, error) {
	return s.sessionService.ValidateSession(ctx, req)
}

func (s *Service) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	return s.sessionService.DeleteSession(ctx, req)
}
