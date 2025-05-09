package service

import (
	"context"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SessionService struct {
	repo          repository.RedisSessionRepos
	kafkaProducer kafka.KafkaProducerService
}

func NewSessionService(repo repository.RedisSessionRepos, kafka kafka.KafkaProducerService) *SessionService {
	return &SessionService{repo: repo, kafkaProducer: kafka}
}
func (s *SessionService) validateContext(ctx context.Context, place string) (string, error) {
	traceID := ctx.Value("traceID").(string)
	select {
	case <-ctx.Done():
		fmterr := fmt.Sprintf("Context error: %v", ctx.Err())
		s.kafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
		return "", status.Errorf(codes.Internal, "Request timed out")
	default:
		return traceID, nil
	}
}
func (s *SessionService) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	var place = "CreateSession"
	traceID, err := s.validateContext(ctx, place)
	if err != nil {
		return nil, err
	}
	if req.UserID == "" {
		s.kafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, "Required userID")
		return nil, status.Errorf(codes.InvalidArgument, "UserID is required")
	}
	if _, err := uuid.Parse(req.UserID); err != nil {
		fmterr := fmt.Sprintf("UUID-Parse userID error: %v", err)
		s.kafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
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
	var place = "ValidateSession"
	traceID, err := s.validateContext(ctx, place)
	if err != nil {
		return nil, err
	}
	if req.SessionID == "" {
		s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Required sessionID")
		return nil, status.Errorf(codes.InvalidArgument, "SessionID is required")
	}
	if _, err := uuid.Parse(req.SessionID); err != nil {
		s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "UUID-Parse sessionID error")
		return nil, status.Errorf(codes.InvalidArgument, "Invalid sessionID format: %v", err)
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
	var place = "DeleteSession"
	traceID, err := s.validateContext(ctx, place)
	if err != nil {
		return nil, err
	}
	if req.SessionID == "" {
		s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Required sessionID")
		return nil, status.Errorf(codes.InvalidArgument, "SessionID is required")
	}
	if _, err := uuid.Parse(req.SessionID); err != nil {
		s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "UUID-Parse sessionID error")
		return nil, status.Errorf(codes.InvalidArgument, "Invalid sessionID format: %v", err)
	}
	response := s.repo.DeleteSession(ctx, req.SessionID)
	if !response.Success {
		return nil, response.Errors
	}
	return &pb.DeleteSessionResponse{
		Success: true,
	}, nil
}
