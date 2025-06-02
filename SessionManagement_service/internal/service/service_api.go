package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/kafka"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type SessionAPI struct {
	pb.UnimplementedSessionServiceServer
	sessionService SessionAuthentication
	kafkaProducer  LogProducer
}
type SessionAuthentication interface {
	CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error)
	ValidateSession(ctx context.Context, req *pb.ValidateSessionRequest) (*pb.ValidateSessionResponse, error)
	DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error)
}
type LogProducer interface {
	NewSessionLog(level, place, traceid, msg string)
}

const API_CreateSession = "API-CreateSession"
const API_ValidateSession = "API-ValidateSession"
const API_DeleteSession = "API-DeleteSession"

const UseCase_CreateSession = "UseCase-CreateSession"
const UseCase_ValidateSession = "UseCase-ValidateSession"
const UseCase_DeleteSession = "UseCase-DeleteSession"

func NewSessionAPI(repos SessionRepos, kafka LogProducer) *SessionAPI {
	return &SessionAPI{
		sessionService: NewSessionService(repos, kafka),
		kafkaProducer:  kafka,
	}
}
func (s *SessionAPI) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	const place = API_CreateSession
	traceID := s.getTraceIdFromMetadata(ctx, place)
	s.kafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "New request has been received")
	ctx = context.WithValue(ctx, "traceID", traceID)
	resp, err := s.sessionService.CreateSession(ctx, req)
	s.kafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	return resp, err
}

func (s *SessionAPI) ValidateSession(ctx context.Context, req *pb.ValidateSessionRequest) (*pb.ValidateSessionResponse, error) {
	const place = API_ValidateSession
	traceID := s.getTraceIdFromMetadata(ctx, place)
	flag := s.getFlagValidate(ctx, place, traceID)
	if flag == "" {
		return &pb.ValidateSessionResponse{Success: false}, status.Errorf(codes.Internal, erro.SessionServiceUnavalaible)
	}
	s.kafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "New request has been received")
	ctx = context.WithValue(ctx, "traceID", traceID)
	ctx = context.WithValue(ctx, "flagvalidate", flag)
	resp, err := s.sessionService.ValidateSession(ctx, req)
	s.kafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	return resp, err
}

func (s *SessionAPI) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	const place = API_DeleteSession
	traceID := s.getTraceIdFromMetadata(ctx, place)
	s.kafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "New request has been received")
	ctx = context.WithValue(ctx, "traceID", traceID)
	resp, err := s.sessionService.DeleteSession(ctx, req)
	s.kafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	return resp, err
}
func (s *SessionAPI) getTraceIdFromMetadata(ctx context.Context, place string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, "", "Metadata not found in context")
		newtrace := uuid.New()
		return newtrace.String()
	}
	traceIDs := md.Get("traceID")
	if len(traceIDs) == 0 || traceIDs[0] == "" {
		s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, "", "Trace ID not found in context")
		newtrace := uuid.New()
		return newtrace.String()
	}
	return traceIDs[0]
}
func (s *SessionAPI) getFlagValidate(ctx context.Context, place string, traceID string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Metadata not found in context")
		return ""
	}
	flagvalidates := md.Get("flagvalidate")
	if len(flagvalidates) == 0 || flagvalidates[0] == "" {
		s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "FlagValidate not found in context")
		return ""
	}
	return flagvalidates[0]
}
