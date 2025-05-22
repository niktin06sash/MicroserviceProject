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
	var place = "UseCase-CreateSession"
	traceID, err := s.validateContext(ctx, place)
	if err != nil {
		return &pb.CreateSessionResponse{Success: false}, err
	}
	if req.UserID == "" {
		s.kafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, "Required userID")
		return &pb.CreateSessionResponse{Success: false}, status.Errorf(codes.InvalidArgument, "UserID is required")
	}
	if _, err := uuid.Parse(req.UserID); err != nil {
		fmterr := fmt.Sprintf("UUID-Parse userID error: %v", err)
		s.kafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
		return &pb.CreateSessionResponse{Success: false}, status.Error(codes.InvalidArgument, "Invalid userID format")
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
		return &pb.CreateSessionResponse{Success: false}, response.Errors
	}
	return &pb.CreateSessionResponse{
		Success:    true,
		SessionID:  sessionID.String(),
		ExpiryTime: expiryTime.Unix(),
	}, nil
}
func (s *SessionService) ValidateSession(ctx context.Context, req *pb.ValidateSessionRequest) (*pb.ValidateSessionResponse, error) {
	var place = "UseCase-ValidateSession"
	traceID, err := s.validateContext(ctx, place)
	if err != nil {
		return &pb.ValidateSessionResponse{Success: false}, err
	}
	flag := ctx.Value("flagvalidate").(string)
	switch flag {
	case "true":
		if req.SessionID == "" {
			s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Required sessionID")
			return &pb.ValidateSessionResponse{Success: false}, status.Errorf(codes.InvalidArgument, "SessionID is required")
		}
		if _, err := uuid.Parse(req.SessionID); err != nil {
			s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Invalid sessionID format")
			return &pb.ValidateSessionResponse{Success: false}, status.Error(codes.InvalidArgument, "Invalid sessionID format")
		}
	case "false":
		if req.SessionID == "" {
			return &pb.ValidateSessionResponse{Success: true}, nil
		}
		if _, err := uuid.Parse(req.SessionID); err != nil {
			s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Request for an unauthorized page with invalid sessionID format")
			return &pb.ValidateSessionResponse{Success: true}, nil
		}
	}
	response := s.repo.GetSession(ctx, req.SessionID)
	if !response.Success {
		return &pb.ValidateSessionResponse{Success: false}, response.Errors
	}
	return &pb.ValidateSessionResponse{
		Success: true,
		UserID:  response.UserID,
	}, nil
}
func (s *SessionService) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	var place = "UseCase-DeleteSession"
	traceID, err := s.validateContext(ctx, place)
	if err != nil {
		return &pb.DeleteSessionResponse{Success: false}, err
	}
	if req.SessionID == "" {
		s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Required sessionID")
		return &pb.DeleteSessionResponse{Success: false}, status.Errorf(codes.InvalidArgument, "SessionID is required")
	}
	if _, err := uuid.Parse(req.SessionID); err != nil {
		s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "UUID-Parse sessionID error")
		return &pb.DeleteSessionResponse{Success: false}, status.Error(codes.InvalidArgument, "Invalid sessionID format")
	}
	response := s.repo.DeleteSession(ctx, req.SessionID)
	if !response.Success {
		return &pb.DeleteSessionResponse{Success: false}, response.Errors
	}
	return &pb.DeleteSessionResponse{
		Success: true,
	}, nil
}
