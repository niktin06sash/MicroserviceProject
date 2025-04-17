package service

import (
	"context"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/logger"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type Service struct {
	pb.UnimplementedSessionServiceServer
	sessionService *SessionService
	logger         *logger.SessionLogger
}

func NewService(repos *repository.Repository, log *logger.SessionLogger) *Service {
	return &Service{
		sessionService: NewSessionService(repos.RedisSessionRepos, log),
		logger:         log,
	}
}
func (s *Service) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.logger.Error("CreateSession: Metadata not found in context", zap.Error(erro.ErrorMissingMetadata))
		return nil, erro.ErrorMissingMetadata
	}

	requestIDs := md.Get("requestID")
	if len(requestIDs) == 0 || requestIDs[0] == "" {
		s.logger.Error("CreateSession: Request ID not found in metadata", zap.Error(erro.ErrorRequiredRequestID))
		return nil, erro.ErrorRequiredRequestID
	}
	resp, err := s.sessionService.CreateSession(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *Service) ValidateSession(ctx context.Context, req *pb.ValidateSessionRequest) (*pb.ValidateSessionResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.logger.Error("ValidateSession: Metadata not found in context", zap.Error(erro.ErrorMissingMetadata))
		return nil, erro.ErrorMissingMetadata
	}

	requestIDs := md.Get("requestID")
	if len(requestIDs) == 0 || requestIDs[0] == "" {
		s.logger.Error("ValidateSession: Request ID not found in metadata", zap.Error(erro.ErrorRequiredRequestID))
		return nil, erro.ErrorRequiredRequestID
	}
	resp, err := s.sessionService.ValidateSession(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *Service) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.logger.Error("DeleteSession: Metadata not found in context", zap.Error(erro.ErrorMissingMetadata))
		return nil, erro.ErrorMissingMetadata
	}

	requestIDs := md.Get("requestID")
	if len(requestIDs) == 0 || requestIDs[0] == "" {
		s.logger.Error("DeleteSession: Request ID not found in metadata", zap.Error(erro.ErrorRequiredRequestID))
		return nil, erro.ErrorRequiredRequestID
	}
	resp, err := s.sessionService.DeleteSession(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
