package service

import (
	"context"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/logger"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"go.uber.org/zap"
)

type Service struct {
	pb.UnimplementedSessionServiceServer
	sessionService *SessionService
	logger         *logger.SessionLogger
}

func NewService(repos *repository.Repository, log *logger.SessionLogger) *Service {
	return &Service{
		sessionService: NewSessionService(repos.RedisSessionRepos),
		logger:         log,
	}
}

func (s *Service) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	resp, err := s.sessionService.CreateSession(ctx, req)
	if err != nil {
		s.logger.Error("CreateSession Error", zap.Error(err))
	}
	return resp, nil
}

func (s *Service) ValidateSession(ctx context.Context, req *pb.ValidateSessionRequest) (*pb.ValidateSessionResponse, error) {
	resp, err := s.sessionService.ValidateSession(ctx, req)
	if err != nil {
		s.logger.Error("ValidateSession Error", zap.Error(err))
	}
	return resp, nil
}

func (s *Service) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	resp, err := s.sessionService.DeleteSession(ctx, req)
	if err != nil {
		s.logger.Error("DeleteSession Error", zap.Error(err))
	}
	return resp, nil
}
