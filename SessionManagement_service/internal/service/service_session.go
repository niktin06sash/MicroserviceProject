package service

import (
	"context"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/logger"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"go.uber.org/zap"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type SessionService struct {
	repo   repository.RedisSessionRepos
	logger *logger.SessionLogger
}

func NewSessionService(repo repository.RedisSessionRepos, log *logger.SessionLogger) *SessionService {
	return &SessionService{repo: repo, logger: log}
}
func validateContext(ctx context.Context, logger *logger.SessionLogger) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Error("Metadata not found in context", zap.Error(erro.ErrorMissingMetadata))
		return "", erro.ErrorMissingMetadata
	}
	requestIDs := md.Get("requestID")
	if len(requestIDs) == 0 || requestIDs[0] == "" {
		logger.Error("Request ID not found in metadata", zap.Error(erro.ErrorRequiredRequestID))
		return "", erro.ErrorRequiredRequestID
	}
	requestID := requestIDs[0]
	if ctx.Err() != nil {
		logger.Error("Context cancelled",
			zap.String("requestID", requestID),
			zap.Error(ctx.Err()),
		)
		return "", status.Errorf(codes.DeadlineExceeded, "request timed out")
	}

	return requestID, nil
}
func (s *SessionService) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	requestID, err := validateContext(ctx, s.logger)
	if err != nil {
		return nil, err
	}
	if req.UserID == "" {
		s.logger.Error("Required userID",
			zap.String("requestID", requestID),
			zap.Error(erro.ErrorRequiredUserId),
		)
		return nil, status.Errorf(codes.InvalidArgument, "userID is required")
	}

	if _, err := uuid.Parse(req.UserID); err != nil {
		s.logger.Error("UUID-Parse userID",
			zap.String("requestID", requestID),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.InvalidArgument, "invalid userID format: %v", err)
	}

	sessionID := uuid.New()
	expiryTime := time.Now().Add(24 * time.Hour)

	newsession := model.Session{
		SessionID:      sessionID.String(),
		UserID:         req.UserID,
		ExpirationTime: expiryTime,
	}
	response := s.repo.SetSession(ctx, newsession)
	if !response.Success {
		return nil, response.Errors
	}
	return &pb.CreateSessionResponse{
		Success:    true,
		SessionID:  sessionID.String(),
		ExpiryTime: expiryTime.Unix(),
	}, nil
}
func (s *SessionService) ValidateSession(ctx context.Context, req *pb.ValidateSessionRequest) (*pb.ValidateSessionResponse, error) {
	requestID, err := validateContext(ctx, s.logger)
	if err != nil {
		return nil, err
	}
	if req.SessionID == "" {
		s.logger.Error("Required sessionID",
			zap.String("requestID", requestID),
			zap.Error(erro.ErrorRequiredSessionId),
		)
		return nil, status.Errorf(codes.InvalidArgument, "sessionID is required")
	}
	response := s.repo.GetSession(ctx, req.SessionID)
	if !response.Success {
		return nil, response.Errors
	}
	return &pb.ValidateSessionResponse{
		Success: true,
		UserID:  response.UserID,
	}, nil
}
func (s *SessionService) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	requestID, err := validateContext(ctx, s.logger)
	if err != nil {
		return nil, err
	}
	if req.SessionID == "" {
		s.logger.Error("Required sessionID",
			zap.String("requestID", requestID),
			zap.Error(erro.ErrorRequiredSessionId),
		)
		return nil, status.Errorf(codes.InvalidArgument, "sessionID is required")
	}

	response := s.repo.DeleteSession(ctx, req.SessionID)
	if !response.Success {
		return nil, response.Errors
	}
	return &pb.DeleteSessionResponse{
		Success: true,
	}, nil
}
