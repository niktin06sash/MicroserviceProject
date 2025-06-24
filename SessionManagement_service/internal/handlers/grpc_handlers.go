package handlers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type SessionAPI struct {
	pb.UnimplementedSessionServiceServer
	sessionService SessionAuthentication
	logproducer    LogProducer
}

func NewSessionAPI(service SessionAuthentication, logproducer LogProducer) *SessionAPI {
	return &SessionAPI{
		sessionService: service,
		logproducer:    logproducer,
	}
}
func (s *SessionAPI) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	const place = CreateSession
	traceID := s.getTraceIdFromMetadata(ctx, place)
	defer s.logproducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	s.logproducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "New request has been received")
	ctx = context.WithValue(ctx, "traceID", traceID)
	resp := s.sessionService.CreateSession(ctx, req.UserID)
	if resp.Errors == nil {
		s.logproducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, fmt.Sprintf("For person with id %s has successfully created session", req.UserID))
		return &pb.CreateSessionResponse{Success: true, SessionID: resp.Data.SessionID, ExpiryTime: resp.Data.ExpirationTime}, nil
	}
	if resp.Errors.Type == erro.ServerErrorType {
		return nil, status.Error(codes.Internal, resp.Errors.Message)
	}
	return nil, status.Error(codes.InvalidArgument, resp.Errors.Message)
}

func (s *SessionAPI) ValidateSession(ctx context.Context, req *pb.ValidateSessionRequest) (*pb.ValidateSessionResponse, error) {
	const place = ValidateSession
	traceID := s.getTraceIdFromMetadata(ctx, place)
	defer s.logproducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	flag := s.getFlagValidate(ctx, place, traceID)
	if flag == "" {
		return &pb.ValidateSessionResponse{Success: false}, status.Errorf(codes.Internal, erro.SessionServiceUnavalaible)
	}
	s.logproducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "New request has been received")
	ctx = context.WithValue(ctx, "traceID", traceID)
	resp := s.sessionService.ValidateSession(ctx, req.SessionID, flag)
	if resp.Errors == nil {
		s.logproducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, fmt.Sprintf("For person with id %s has successfully validated session", resp.Data.UserID))
		return &pb.ValidateSessionResponse{Success: true, UserID: resp.Data.UserID}, nil
	}
	if resp.Errors.Type == erro.ServerErrorType {
		return nil, status.Error(codes.Internal, resp.Errors.Message)
	}
	return nil, status.Error(codes.InvalidArgument, resp.Errors.Message)
}
func (s *SessionAPI) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	const place = DeleteSession
	traceID := s.getTraceIdFromMetadata(ctx, place)
	defer s.logproducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	s.logproducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "New request has been received")
	ctx = context.WithValue(ctx, "traceID", traceID)
	resp := s.sessionService.DeleteSession(ctx, req.SessionID)
	if resp.Errors == nil {
		s.logproducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, fmt.Sprintf("Session with id %v has successfully deleted", req.SessionID))
		return &pb.DeleteSessionResponse{Success: true}, nil
	}
	if resp.Errors.Type == erro.ServerErrorType {
		return nil, status.Error(codes.Internal, resp.Errors.Message)
	}
	return nil, status.Error(codes.InvalidArgument, resp.Errors.Message)
}
func (s *SessionAPI) getTraceIdFromMetadata(ctx context.Context, place string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.logproducer.NewSessionLog(kafka.LogLevelWarn, place, "", "Metadata not found in context")
		newtrace := uuid.New()
		return newtrace.String()
	}
	traceIDs := md.Get("traceID")
	if len(traceIDs) == 0 || traceIDs[0] == "" {
		s.logproducer.NewSessionLog(kafka.LogLevelWarn, place, "", "Trace ID not found in context")
		newtrace := uuid.New()
		return newtrace.String()
	}
	return traceIDs[0]
}
func (s *SessionAPI) getFlagValidate(ctx context.Context, place string, traceID string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.logproducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Metadata not found in context")
		return ""
	}
	flagvalidates := md.Get("flagvalidate")
	if len(flagvalidates) == 0 || flagvalidates[0] == "" {
		s.logproducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "FlagValidate not found in context")
		return ""
	}
	return flagvalidates[0]
}
