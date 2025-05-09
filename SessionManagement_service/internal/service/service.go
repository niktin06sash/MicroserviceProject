package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"google.golang.org/grpc/metadata"
)

type SessionAPI struct {
	pb.UnimplementedSessionServiceServer
	sessionService SessionAuthentication
	kafkaProducer  kafka.KafkaProducerService
}
type SessionAuthentication interface {
	CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error)
	ValidateSession(ctx context.Context, req *pb.ValidateSessionRequest) (*pb.ValidateSessionResponse, error)
	DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error)
}

func NewSessionAPI(repos *repository.Repository, kafka kafka.KafkaProducerService) *SessionAPI {
	return &SessionAPI{
		sessionService: NewSessionService(repos.RedisSessionRepos, kafka),
		kafkaProducer:  kafka,
	}
}
func (s *SessionAPI) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	var place = "CreateSession"
	traceID := s.getTraceIdFromMetadata(ctx, place)
	ctx = context.WithValue(ctx, "traceID", traceID)
	resp, err := s.sessionService.CreateSession(ctx, req)
	s.kafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	return resp, err
}

func (s *SessionAPI) ValidateSession(ctx context.Context, req *pb.ValidateSessionRequest) (*pb.ValidateSessionResponse, error) {
	var place = "ValidateSession"
	traceID := s.getTraceIdFromMetadata(ctx, place)
	ctx = context.WithValue(ctx, "traceID", traceID)
	resp, err := s.sessionService.ValidateSession(ctx, req)
	s.kafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	return resp, err
}

func (s *SessionAPI) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	var place = "DeleteSession"
	traceID := s.getTraceIdFromMetadata(ctx, place)
	ctx = context.WithValue(ctx, "traceID", traceID)
	resp, err := s.sessionService.DeleteSession(ctx, req)
	s.kafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	return resp, err
}
func (s *SessionAPI) getTraceIdFromMetadata(ctx context.Context, place string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, "", "Metadata not found in context")
		newreq := uuid.New()
		return newreq.String()
	}
	traceIDs := md.Get("traceID")
	if len(traceIDs) == 0 || traceIDs[0] == "" {
		s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, "", "Trace ID not found in context")
		newtrace := uuid.New()
		return newtrace.String()
	}
	return traceIDs[0]
}
