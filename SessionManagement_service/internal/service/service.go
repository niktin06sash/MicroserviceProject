package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/logger"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type SessionAPI struct {
	pb.UnimplementedSessionServiceServer
	sessionService SessionAuthentication
	logger         *logger.SessionLogger
}
type SessionAuthentication interface {
	CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error)
	ValidateSession(ctx context.Context, req *pb.ValidateSessionRequest) (*pb.ValidateSessionResponse, error)
	DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error)
}

func getRequestIdFromMetadata(ctx context.Context, log *logger.SessionLogger, place string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Warn(fmt.Sprintf("[%s] Metadata not found in context", place), zap.Error(erro.ErrorMissingMetadata))
		newreq := uuid.New()
		return newreq.String()
	}
	requestIDs := md.Get("requestID")
	if len(requestIDs) == 0 || requestIDs[0] == "" {
		log.Warn(fmt.Sprintf("[%s] Request ID not found in context", place), zap.Error(erro.ErrorRequiredRequestID))
		newreq := uuid.New()
		return newreq.String()
	}
	return requestIDs[0]
}
func NewSessionAPI(repos *repository.Repository, log *logger.SessionLogger) *SessionAPI {
	return &SessionAPI{
		sessionService: NewSessionService(repos.RedisSessionRepos, log),
		logger:         log,
	}
}
func (s *SessionAPI) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	requestID := getRequestIdFromMetadata(ctx, s.logger, "CreateSession")
	ctx = context.WithValue(ctx, "requestID", requestID)
	resp, err := s.sessionService.CreateSession(ctx, req)
	s.logger.Info("CreateSession: Succesfull send response to client",
		zap.String("requestID", requestID),
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *SessionAPI) ValidateSession(ctx context.Context, req *pb.ValidateSessionRequest) (*pb.ValidateSessionResponse, error) {
	requestID := getRequestIdFromMetadata(ctx, s.logger, "ValidateSession")
	ctx = context.WithValue(ctx, "requestID", requestID)
	resp, err := s.sessionService.ValidateSession(ctx, req)
	s.logger.Info("ValidateSession: Succesfull send response to client",
		zap.String("requestID", requestID),
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *SessionAPI) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	requestID := getRequestIdFromMetadata(ctx, s.logger, "DeleteSession")
	ctx = context.WithValue(ctx, "requestID", requestID)
	resp, err := s.sessionService.DeleteSession(ctx, req)
	s.logger.Info("DeleteSession: Succesfull send response to client",
		zap.String("requestID", requestID),
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
