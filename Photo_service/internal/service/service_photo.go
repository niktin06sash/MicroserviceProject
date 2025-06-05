package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository"
	pb "github.com/niktin06sash/MicroserviceProject/Photo_service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DBPhotoRepos interface {
	LoadPhoto(ctx context.Context, photo *model.Photo) *repository.RepositoryResponse
	DeletePhoto(ctx context.Context, userid string, photoid string) *repository.RepositoryResponse
	GetPhotos(ctx context.Context, userid string) *repository.RepositoryResponse
	GetPhoto(ctx context.Context, photoid string) *repository.RepositoryResponse
}
type CloudPhotoStorage interface {
	UploadFile(localfilepath string, photoid string, ext string) *repository.RepositoryResponse
	DeleteFile(id, contenttype string) *repository.RepositoryResponse
}
type PhotoService struct {
	repo          DBPhotoRepos
	cloud         CloudPhotoStorage
	kafkaProducer LogProducer
}

func NewPhotoService(repo DBPhotoRepos, cloud CloudPhotoStorage, kafka LogProducer) *PhotoService {
	return &PhotoService{repo: repo, kafkaProducer: kafka, cloud: cloud}
}

const UseCase_LoadPhoto = "UseCase_LoadPhoto"
const MaxFileSize = 2 << 20
const PhotoUnloadAndSave = "PhotoUnloadAndSave"
const DeletePhotoCloud = "DeletePhotoCloud"

func (use *PhotoService) DeletePhoto(ctx context.Context, req *pb.DeletePhotoRequest) (*pb.DeletePhotoResponse, error) {
	traceid := ctx.Value("traceID").(string)
	bdresponse := use.repo.DeletePhoto(ctx, req.UserId, req.PhotoId)
	if !bdresponse.Success && bdresponse.Errors != nil {
		if bdresponse.Errors.Type == erro.ClientErrorType {
			use.kafkaProducer.NewPhotoLog(kafka.LogLevelWarn, bdresponse.Place, traceid, bdresponse.Errors.Message)
			return nil, status.Errorf(codes.InvalidArgument, bdresponse.Errors.Message)
		}
		use.kafkaProducer.NewPhotoLog(kafka.LogLevelError, bdresponse.Place, traceid, bdresponse.Errors.Message)
		return nil, status.Errorf(codes.InvalidArgument, bdresponse.Errors.Message)
	}
	ext := bdresponse.Data[repository.KeyContentType].(string)
	go use.deletePhotoCloud(traceid, req.PhotoId, ext)
	return &pb.DeletePhotoResponse{Status: true, Message: "The photo was deleted successfully"}, nil
}
func (use *PhotoService) LoadPhoto(ctx context.Context, req *pb.LoadPhotoRequest) (*pb.LoadPhotoResponse, error) {
	const place = UseCase_LoadPhoto
	traceid := ctx.Value("traceID").(string)
	if len(req.FileData) > MaxFileSize {
		use.kafkaProducer.NewPhotoLog(kafka.LogLevelWarn, place, traceid, fmt.Sprintf("File too large: %v bytes", len(req.FileData)))
		return nil, status.Errorf(codes.InvalidArgument, "File size exceeds 2MB limit")
	}
	contentType := http.DetectContentType(req.FileData)
	if contentType != "image/jpeg" && contentType != "image/png" {
		use.kafkaProducer.NewPhotoLog(kafka.LogLevelWarn, place, traceid, fmt.Sprintf("Invalid file format: %s", contentType))
		return nil, status.Errorf(codes.InvalidArgument, "Only JPG/PNG formats allowed")
	}
	photoid := uuid.New().String()
	var ext string
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	}
	go use.photoUnloadAndSave(ctx, req.FileData, photoid, ext, req.UserId)
	return &pb.LoadPhotoResponse{Status: true, PhotoId: photoid, Message: "The photo was added successfully"}, nil
}
func (use *PhotoService) GetPhoto(ctx context.Context, req *pb.GetPhotoRequest) (*pb.GetPhotoResponse, error) {
	traceid := ctx.Value("traceID").(string)
	bdresponse := use.repo.GetPhoto(ctx, req.PhotoId)
	if !bdresponse.Success && bdresponse.Errors != nil {
		if bdresponse.Errors.Type == erro.ClientErrorType {
			use.kafkaProducer.NewPhotoLog(kafka.LogLevelWarn, bdresponse.Place, traceid, bdresponse.Errors.Message)
			return nil, status.Errorf(codes.InvalidArgument, bdresponse.Errors.Message)
		}
		use.kafkaProducer.NewPhotoLog(kafka.LogLevelError, bdresponse.Place, traceid, bdresponse.Errors.Message)
		return nil, status.Errorf(codes.InvalidArgument, bdresponse.Errors.Message)
	}
	photo := bdresponse.Data[repository.KeyPhoto].(*model.Photo)
	return &pb.GetPhotoResponse{Status: true, Photo: &pb.Photo{PhotoId: photo.ID, Url: photo.URL, CreatedAt: photo.CreatedAt.String()}}, nil
}
func (use *PhotoService) GetPhotos(ctx context.Context, req *pb.GetPhotosRequest) (*pb.GetPhotosResponse, error) {
	traceid := ctx.Value("traceID").(string)
	bdresponse := use.repo.GetPhotos(ctx, req.UserId)
	if !bdresponse.Success && bdresponse.Errors != nil {
		if bdresponse.Errors.Type == erro.ClientErrorType {
			use.kafkaProducer.NewPhotoLog(kafka.LogLevelWarn, bdresponse.Place, traceid, bdresponse.Errors.Message)
			return nil, status.Errorf(codes.InvalidArgument, bdresponse.Errors.Message)
		}
		use.kafkaProducer.NewPhotoLog(kafka.LogLevelError, bdresponse.Place, traceid, bdresponse.Errors.Message)
		return nil, status.Errorf(codes.InvalidArgument, bdresponse.Errors.Message)
	}
	photos := bdresponse.Data[repository.KeyPhoto].([]*model.Photo)
	grpcphotos := []*pb.Photo{}
	for _, p := range photos {
		grpcphoto := &pb.Photo{PhotoId: p.ID, Url: p.URL, CreatedAt: p.CreatedAt.String()}
		grpcphotos = append(grpcphotos, grpcphoto)
	}
	return &pb.GetPhotosResponse{Status: true, Photos: grpcphotos}, nil
}
func (use *PhotoService) photoUnloadAndSave(ctx context.Context, file []byte, photoid string, ext string, userid string) {
	const place = PhotoUnloadAndSave
	traceid := ctx.Value("traceID").(string)
	filename := photoid + ext
	tmpDir := os.TempDir()
	tempFile := filepath.Join(tmpDir, filename)
	err := os.WriteFile(tempFile, file, 0644)
	if err != nil {
		use.kafkaProducer.NewPhotoLog(kafka.LogLevelError, place, traceid, fmt.Sprintf("Failed to create temp file: %v", err))
		return
	}
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			use.kafkaProducer.NewPhotoLog(kafka.LogLevelError, place, traceid, fmt.Sprintf("Failed to remove temp file: %v", err))
		}
	}()
	cloudresponse := use.cloud.UploadFile(tempFile, photoid, ext)
	if !cloudresponse.Success && cloudresponse.Errors != nil {
		use.kafkaProducer.NewPhotoLog(kafka.LogLevelError, cloudresponse.Place, traceid, cloudresponse.Errors.Message)
	}
	photo := cloudresponse.Data[repository.KeyPhoto].(*model.Photo)
	photo.UserID = userid
	bdresponse := use.repo.LoadPhoto(ctx, photo)
	if !bdresponse.Success && bdresponse.Errors != nil {
		use.kafkaProducer.NewPhotoLog(kafka.LogLevelError, bdresponse.Place, traceid, bdresponse.Errors.Message)
	}
	use.kafkaProducer.NewPhotoLog(kafka.LogLevelInfo, place, traceid, fmt.Sprintf("The photo(id = %s) has been successfully uploaded to cloud and database", photoid))
}
func (use *PhotoService) deletePhotoCloud(traceid string, photoid string, ext string) {
	const place = DeletePhotoCloud
	cloudresponse := use.cloud.DeleteFile(photoid, ext)
	if !cloudresponse.Success && cloudresponse.Errors != nil {
		use.kafkaProducer.NewPhotoLog(kafka.LogLevelError, cloudresponse.Place, traceid, cloudresponse.Errors.Message)
	}
	use.kafkaProducer.NewPhotoLog(kafka.LogLevelInfo, place, traceid, fmt.Sprintf("The photo(id = %s) has been successfully deleted from cloud and database", photoid))
}
