package service

import (
	"context"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
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
func (s *SessionService) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	const place = UseCase_CreateSession
	traceID := ctx.Value("traceID").(string)
	if req.UserID == "" {
		s.kafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, "Required userID")
		return &pb.CreateSessionResponse{Success: false}, status.Errorf(codes.InvalidArgument, erro.UserIdRequired)
	}
	if _, err := uuid.Parse(req.UserID); err != nil {
		fmterr := fmt.Sprintf("UUID-Parse userID error: %v", err)
		s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, fmterr)
		return &pb.CreateSessionResponse{Success: false}, status.Error(codes.InvalidArgument, erro.UserIdInvalid)
	}
	sessionID := uuid.New()
	expiryTime := time.Now().Add(24 * time.Hour)
	newsession := model.Session{
		SessionID:      sessionID.String(),
		UserID:         req.UserID,
		ExpirationTime: expiryTime,
	}
	response := s.repo.SetSession(ctx, newsession)
	return &pb.CreateSessionResponse{Success: response.Success, SessionID: response.SessionId, ExpiryTime: response.ExpirationTime.Unix()}, response.Errors
}
func (s *SessionService) ValidateSession(ctx context.Context, req *pb.ValidateSessionRequest) (*pb.ValidateSessionResponse, error) {
	const place = UseCase_ValidateSession
	traceID := ctx.Value("traceID").(string)
	flag := ctx.Value("flagvalidate").(string)
	switch flag {
	case "true":
		if _, err := uuid.Parse(req.SessionID); err != nil {
			fmterr := fmt.Sprintf("UUID-Parse sessionID error: %v", err)
			s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, fmterr)
			return &pb.ValidateSessionResponse{Success: false}, status.Error(codes.InvalidArgument, erro.SessionIdInvalid)
		}
	case "false":
		if _, err := uuid.Parse(req.SessionID); err != nil {
			s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Request for an unauthorized users with invalid sessionID format")
			return &pb.ValidateSessionResponse{Success: true}, nil
		}
	}
	response := s.repo.GetSession(ctx, req.SessionID)
	return &pb.ValidateSessionResponse{Success: response.Success, UserID: response.UserID}, response.Errors
}
func (s *SessionService) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	const place = UseCase_DeleteSession
	traceID := ctx.Value("traceID").(string)
	if _, err := uuid.Parse(req.SessionID); err != nil {
		fmterr := fmt.Sprintf("UUID-Parse sessionID error: %v", err)
		s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, fmterr)
		return &pb.DeleteSessionResponse{Success: false}, status.Error(codes.InvalidArgument, erro.SessionIdInvalid)
	}
	response := s.repo.DeleteSession(ctx, req.SessionID)
	return &pb.DeleteSessionResponse{Success: response.Success}, response.Errors
}
