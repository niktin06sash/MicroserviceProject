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

type SessionRepos interface {
	SetSession(ctx context.Context, session model.Session) *repository.RepositoryResponse
	GetSession(ctx context.Context, sessionID string) *repository.RepositoryResponse
	DeleteSession(ctx context.Context, sessionID string) *repository.RepositoryResponse
}
type SessionService struct {
	repo          SessionRepos
	kafkaProducer kafka.KafkaProducerService
}

func NewSessionService(repo SessionRepos, kafka kafka.KafkaProducerService) *SessionService {
	return &SessionService{repo: repo, kafkaProducer: kafka}
}
func (s *SessionService) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	const place = UseCase_CreateSession
	traceID := ctx.Value("traceID").(string)
	if req.UserID == "" {
		s.kafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, "Required userID")
		return &pb.CreateSessionResponse{Success: false}, status.Errorf(codes.Internal, erro.UserIdRequired)
	}
	if _, err := uuid.Parse(req.UserID); err != nil {
		fmterr := fmt.Sprintf("UUID-Parse userID error: %v", err)
		s.kafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
		return &pb.CreateSessionResponse{Success: false}, status.Error(codes.Internal, erro.UserIdInvalid)
	}
	sessionID := uuid.New()
	expiryTime := time.Now().Add(24 * time.Hour)
	newsession := model.Session{
		SessionID:      sessionID.String(),
		UserID:         req.UserID,
		ExpirationTime: expiryTime,
	}
	data, err := requestToDB(ctx, func(ctx context.Context) *repository.RepositoryResponse { return s.repo.SetSession(ctx, newsession) }, traceID, s.kafkaProducer)
	if err != nil {
		return nil, err
	}
	sessionid := data[repository.KeySessionId].(string)
	exp := data[repository.KeySessionId].(time.Time).Unix()
	return &pb.CreateSessionResponse{Success: true, SessionID: sessionid, ExpiryTime: exp}, nil
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
	data, err := requestToDB(ctx, func(ctx context.Context) *repository.RepositoryResponse { return s.repo.GetSession(ctx, req.SessionID) }, traceID, s.kafkaProducer)
	if err != nil {
		return nil, err
	}
	userID := data[repository.KeyUserId].(string)
	return &pb.ValidateSessionResponse{Success: true, UserID: userID}, nil
}
func (s *SessionService) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	const place = UseCase_DeleteSession
	traceID := ctx.Value("traceID").(string)
	if _, err := uuid.Parse(req.SessionID); err != nil {
		fmterr := fmt.Sprintf("UUID-Parse sessionID error: %v", err)
		s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, fmterr)
		return &pb.DeleteSessionResponse{Success: false}, status.Error(codes.Internal, erro.SessionIdInvalid)
	}
	_, err := requestToDB(ctx, func(ctx context.Context) *repository.RepositoryResponse { return s.repo.GetSession(ctx, req.SessionID) }, traceID, s.kafkaProducer)
	if err != nil {
		return nil, err
	}
	return &pb.DeleteSessionResponse{Success: true}, nil
}
func requestToDB(ctx context.Context, operation func(context.Context) *repository.RepositoryResponse, traceid string, kafkaprod kafka.KafkaProducerService) (map[string]any, error) {
	response := operation(ctx)
	if response.Errors != nil {
		st, _ := status.FromError(response.Errors)
		switch st.Code() {
		case codes.Internal, codes.Unavailable, codes.Canceled:
			kafkaprod.NewSessionLog(kafka.LogLevelError, response.Place, traceid, st.Message())
			return nil, status.Error(st.Code(), erro.SessionServiceUnavalaible)
		case codes.InvalidArgument:
			kafkaprod.NewSessionLog(kafka.LogLevelWarn, response.Place, traceid, st.Message())
			return nil, response.Errors
		}
	}
	kafkaprod.NewSessionLog(kafka.LogLevelInfo, response.Place, traceid, response.SuccessMessage)
	return response.Data, nil
}
