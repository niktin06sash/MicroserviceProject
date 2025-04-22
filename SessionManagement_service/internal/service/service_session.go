package service

import (
	"context"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/logger"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"go.uber.org/zap"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SessionService struct {
	repo   repository.RedisSessionRepos
	logger *logger.SessionLogger
}

func NewSessionService(repo repository.RedisSessionRepos, log *logger.SessionLogger) *SessionService {
	return &SessionService{repo: repo, logger: log}
}
func validateContext(ctx context.Context, logger *logger.SessionLogger, place string) (string, error) {
	traceID := ctx.Value("traceID").(string)
	select {
	case <-ctx.Done():
		logger.Error(fmt.Sprintf("[%s] Context time-out", place),
			zap.String("traceID", traceID),
			zap.Error(ctx.Err()),
		)
		return "", status.Errorf(codes.Internal, "Request timed out")
	default:
		return traceID, nil
	}
}
func (s *SessionService) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	traceID, err := validateContext(ctx, s.logger, "CreateSession")
	if err != nil {
		return nil, err
	}
	if req.UserID == "" {
		s.logger.Error("CreateSession: Required userID",
			zap.String("traceID", traceID),
			zap.Error(erro.ErrorRequiredUserId),
		)
		return nil, status.Errorf(codes.InvalidArgument, "UserID is required")
	}
	if _, err := uuid.Parse(req.UserID); err != nil {
		s.logger.Error("CreateSession: Error UUID-Parse userID",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid userID format: %v", err)
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
	traceID, err := validateContext(ctx, s.logger, "ValidateSession")
	if err != nil {
		return nil, err
	}
	if req.SessionID == "" {
		s.logger.Error("ValidateSession: Required sessionID",
			zap.String("traceID", traceID),
			zap.Error(erro.ErrorRequiredSessionId),
		)
		return nil, status.Errorf(codes.InvalidArgument, "SessionID is required")
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
	traceID, err := validateContext(ctx, s.logger, "DeleteSession")
	if err != nil {
		return nil, err
	}
	if req.SessionID == "" {
		s.logger.Error("DeleteSession: Required sessionID",
			zap.String("traceID", traceID),
			zap.Error(erro.ErrorRequiredSessionId),
		)
		return nil, status.Errorf(codes.InvalidArgument, "SessionID is required")
	}
	response := s.repo.DeleteSession(ctx, req.SessionID)
	if !response.Success {
		return nil, response.Errors
	}
	return &pb.DeleteSessionResponse{
		Success: true,
	}, nil
}
