package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	pb "github.com/niktin06sash/MicroserviceProject/Photo_service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type PhotoAPI struct {
	pb.UnimplementedPhotoServiceServer
	photoService PhotoService
	logproducer  LogProducer
}

func NewPhotoAPI(service PhotoService, logProducer LogProducer) *PhotoAPI {
	return &PhotoAPI{
		logproducer:  logProducer,
		photoService: service,
	}
}
func (s *PhotoAPI) LoadPhoto(ctx context.Context, req *pb.LoadPhotoRequest) (*pb.LoadPhotoResponse, error) {
	const place = API_LoadPhoto
	traceID := s.getTraceIdFromMetadata(ctx, place)
	defer s.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	s.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, "New request has been received")
	ctx = context.WithValue(ctx, "traceID", traceID)
	serviceresp := s.photoService.LoadPhoto(ctx, req.UserId, req.FileData)
	if serviceresp.Errors == nil {
		s.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, fmt.Sprintf("Person with id %s has successfully uploaded photo", req.UserId))
		return &pb.LoadPhotoResponse{Status: true, PhotoId: serviceresp.Data.PhotoID, Message: "You have successfully uploaded photo"}, nil
	}
	return nil, handleGrpcError(serviceresp.Errors)
}

func (s *PhotoAPI) DeletePhoto(ctx context.Context, req *pb.DeletePhotoRequest) (*pb.DeletePhotoResponse, error) {
	const place = API_DeletePhoto
	traceID := s.getTraceIdFromMetadata(ctx, place)
	defer s.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	s.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, "New request has been received")
	ctx = context.WithValue(ctx, "traceID", traceID)
	serviceresp := s.photoService.DeletePhoto(ctx, req.UserId, req.PhotoId)
	if serviceresp.Errors == nil {
		s.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, fmt.Sprintf("Person with id %s has successfully deleted photo with id %s", req.UserId, req.PhotoId))
		return &pb.DeletePhotoResponse{Status: true, Message: "You have successfully deleted photo"}, nil
	}
	return nil, handleGrpcError(serviceresp.Errors)
}

func (s *PhotoAPI) GetPhoto(ctx context.Context, req *pb.GetPhotoRequest) (*pb.GetPhotoResponse, error) {
	const place = API_GetPhoto
	traceID := s.getTraceIdFromMetadata(ctx, place)
	defer s.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	s.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, "New request has been received")
	ctx = context.WithValue(ctx, "traceID", traceID)
	serviceresp := s.photoService.GetPhoto(ctx, req.PhotoId, req.UserId)
	if serviceresp.Errors == nil {
		s.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, fmt.Sprintf("Person with id %s has successfully get photo with id %s", req.UserId, req.PhotoId))
		return &pb.GetPhotoResponse{Status: true, Photo: serviceresp.Data.Photo}, nil
	}
	return nil, handleGrpcError(serviceresp.Errors)
}
func (s *PhotoAPI) GetPhotos(ctx context.Context, req *pb.GetPhotosRequest) (*pb.GetPhotosResponse, error) {
	const place = API_GetPhotos
	traceID := s.getTraceIdFromMetadata(ctx, place)
	defer s.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, "Succesfull send response to client")
	s.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, "New request has been received")
	ctx = context.WithValue(ctx, "traceID", traceID)
	serviceresp := s.photoService.GetPhotos(ctx, req.UserId)
	if serviceresp.Errors == nil {
		s.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceID, fmt.Sprintf("Successfully get person's with id %s photos", req.UserId))
		return &pb.GetPhotosResponse{Status: true, Photos: serviceresp.Data.Photos}, nil
	}
	return nil, handleGrpcError(serviceresp.Errors)
}
func (s *PhotoAPI) getTraceIdFromMetadata(ctx context.Context, place string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.logproducer.NewPhotoLog(kafka.LogLevelWarn, place, "", "Metadata not found in context")
		newtrace := uuid.New()
		return newtrace.String()
	}
	traceIDs := md.Get("traceID")
	if len(traceIDs) == 0 || traceIDs[0] == "" {
		s.logproducer.NewPhotoLog(kafka.LogLevelWarn, place, "", "Trace ID not found in context")
		newtrace := uuid.New()
		return newtrace.String()
	}
	return traceIDs[0]
}
func handleGrpcError(err error) error {
	var customErr *erro.CustomError
	if errors.As(err, &customErr) {
		switch customErr.Type {
		case erro.ClientErrorType:
			return status.Error(codes.InvalidArgument, customErr.Message)
		case erro.ServerErrorType:
			return status.Error(codes.Internal, customErr.Message)
		}
	}
	return status.Error(codes.Internal, erro.PhotoServiceUnavalaible)
}
