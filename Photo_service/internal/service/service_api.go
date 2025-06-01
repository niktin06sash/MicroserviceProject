package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository"
	pb "github.com/niktin06sash/MicroserviceProject/Photo_service/proto"
	"google.golang.org/grpc/metadata"
)

type PhotoAPI struct {
	pb.UnimplementedPhotoServiceServer
	sessionService SessionAuthentication
	kafkaProducer  kafka.KafkaProducerService
}

func NewPhotoAPI(repos *repository.Repository, kafka kafka.KafkaProducerService) *PhotoAPI {
	return &PhotoAPI{
		kafkaProducer: kafka,
	}
}

const API_LoadPhoto = "API-LoadPhoto"
const API_DeletePhoto = "API-DeletePhoto"
const API_GetPhoto = "API-GetPhoto"
const API_GetPhotos = "API-GetPhotos"

type SessionAuthentication interface {
	DeletePhoto(ctx context.Context, req *pb.DeletePhotoRequest) (*pb.DeletePhotoResponse, error)
	LoadPhoto(ctx context.Context, req *pb.LoadPhotoRequest) (*pb.LoadPhotoResponse, error)
	GetPhoto(ctx context.Context, req *pb.GetPhotoRequest) (*pb.GetPhotoResponse, error)
	GetPhotos(ctx context.Context, req *pb.GetPhotosRequest) (*pb.GetPhotosResponse, error)
}

func (s *PhotoAPI) LoadPhoto(ctx context.Context, req *pb.LoadPhotoRequest) (*pb.LoadPhotoResponse, error) {
	const place = API_LoadPhoto
	traceID := s.getTraceIdFromMetadata(ctx, place)
	s.kafkaProducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, "New request has been received")
	ctx = context.WithValue(ctx, "traceID", traceID)
	//resp, err := s.sessionService.CreateSession(ctx, req)
	s.kafkaProducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	return resp, err
}

func (s *PhotoAPI) DeletePhoto(ctx context.Context, req *pb.DeletePhotoRequest) (*pb.DeletePhotoResponse, error) {
	const place = API_DeletePhoto
	traceID := s.getTraceIdFromMetadata(ctx, place)
	s.kafkaProducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, "New request has been received")
	ctx = context.WithValue(ctx, "traceID", traceID)
	//resp, err := s.sessionService.ValidateSession(ctx, req)
	s.kafkaProducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	return resp, err
}

func (s *PhotoAPI) GetPhoto(ctx context.Context, req *pb.GetPhotoRequest) (*pb.GetPhotoResponse, error) {
	const place = API_GetPhoto
	traceID := s.getTraceIdFromMetadata(ctx, place)
	s.kafkaProducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, "New request has been received")
	ctx = context.WithValue(ctx, "traceID", traceID)
	//resp, err := s.sessionService.DeleteSession(ctx, req)
	s.kafkaProducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	return resp, err
}
func (s *PhotoAPI) GetPhotos(ctx context.Context, req *pb.GetPhotosRequest) (*pb.GetPhotosResponse, error) {
	const place = API_GetPhotos
	traceID := s.getTraceIdFromMetadata(ctx, place)
	s.kafkaProducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, "New request has been received")
	ctx = context.WithValue(ctx, "traceID", traceID)
	//resp, err := s.sessionService.DeleteSession(ctx, req)
	s.kafkaProducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	return resp, err
}
func (s *PhotoAPI) getTraceIdFromMetadata(ctx context.Context, place string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.kafkaProducer.NewPhotoLog(kafka.LogLevelWarn, place, "", "Metadata not found in context")
		newtrace := uuid.New()
		return newtrace.String()
	}
	traceIDs := md.Get("traceID")
	if len(traceIDs) == 0 || traceIDs[0] == "" {
		s.kafkaProducer.NewPhotoLog(kafka.LogLevelWarn, place, "", "Trace ID not found in context")
		newtrace := uuid.New()
		return newtrace.String()
	}
	return traceIDs[0]
}
