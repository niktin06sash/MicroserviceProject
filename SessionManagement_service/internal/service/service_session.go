package service

import (
	"context"
	"log"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SessionService struct {
	repo repository.RedisSessionRepos
}

func NewSessionService(repo repository.RedisSessionRepos) *SessionService {
	return &SessionService{repo: repo}
}

func (s *SessionService) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	if ctx.Err() != nil {
		return nil, status.Errorf(codes.DeadlineExceeded, "request timed out")
	}
	if req.UserID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "userID is required")
	}

	if _, err := uuid.Parse(req.UserID); err != nil {
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
	log.Printf("Session created: sessionID=%s, userID=%s", sessionID, req.UserID)
	return &pb.CreateSessionResponse{
		Success:    true,
		SessionID:  sessionID.String(),
		ExpiryTime: expiryTime.Unix(),
	}, nil
}
func (s *SessionService) ValidateSession(ctx context.Context, req *pb.ValidateSessionRequest) (*pb.ValidateSessionResponse, error) {
	if ctx.Err() != nil {
		return nil, status.Errorf(codes.DeadlineExceeded, "request timed out")
	}
	if req.SessionID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "sessionID is required")
	}
	response := s.repo.GetSession(ctx, req.SessionID)
	if !response.Success {
		return nil, response.Errors
	}
	log.Printf("Session validated: sessionID=%s, userID=%s", req.SessionID, response.UserID)
	return &pb.ValidateSessionResponse{
		Success: true,
		UserID:  response.UserID,
	}, nil
}
func (s *SessionService) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	if ctx.Err() != nil {
		return nil, status.Errorf(codes.DeadlineExceeded, "request timed out")
	}
	if req.SessionID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "sessionID is required")
	}

	response := s.repo.DeleteSession(ctx, req.SessionID)
	if !response.Success {
		return nil, response.Errors
	}
	log.Printf("Session deleted: sessionID=%s", req.SessionID)
	return &pb.DeleteSessionResponse{
		Success: true,
	}, nil
}
