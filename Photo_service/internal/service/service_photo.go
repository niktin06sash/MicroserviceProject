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
type LogProducer interface {
	NewPhotoLog(level, place, traceid, msg string)
}
type PhotoService struct {
	repo          DBPhotoRepos
	cloud         CloudPhotoStorage
	kafkaProducer LogProducer
}
type ServiceResponse struct {
	Success bool
	Data    map[string]any
	Errors  map[string]string
}

func NewPhotoService(repo DBPhotoRepos, cloud CloudPhotoStorage, kafka LogProducer) *PhotoService {
	return &PhotoService{repo: repo, kafkaProducer: kafka, cloud: cloud}
}

const UseCase_LoadPhoto = "UseCase_LoadPhoto"
const MaxFileSize = 2 << 20
const PhotoUnloadAndSave = "PhotoUnloadAndSave"
const DeletePhotoCloud = "DeletePhotoCloud"

func (use *PhotoService) DeletePhoto(ctx context.Context, userid string, photoid string) *ServiceResponse {
	delmapa := make(map[string]string)
	traceid := ctx.Value("traceID").(string)
	bdresponse := use.repo.DeletePhoto(ctx, userid, photoid)
	if !bdresponse.Success && bdresponse.Errors != nil {
		if bdresponse.Errors.Type == erro.ClientErrorType {
			use.kafkaProducer.NewPhotoLog(kafka.LogLevelWarn, bdresponse.Place, traceid, bdresponse.Errors.Message)
			delmapa[erro.ErrorType] = erro.ClientErrorType
			delmapa[erro.ErrorMessage] = erro.ErrorMessage
			return &ServiceResponse{Success: false, Errors: delmapa}
		}
		use.kafkaProducer.NewPhotoLog(kafka.LogLevelError, bdresponse.Place, traceid, bdresponse.Errors.Message)
		return &ServiceResponse{Success: false, Errors: delmapa}
	}
	ext := bdresponse.Data[repository.KeyContentType].(string)
	go use.deletePhotoCloud(traceid, photoid, ext)
	return &ServiceResponse{Success: true}
}
func (use *PhotoService) LoadPhoto(ctx context.Context, userid string, filedata []byte) *ServiceResponse {
	loadmapa := make(map[string]string)
	const place = UseCase_LoadPhoto
	traceid := ctx.Value("traceID").(string)
	if len(filedata) > MaxFileSize {
		use.kafkaProducer.NewPhotoLog(kafka.LogLevelWarn, place, traceid, fmt.Sprintf("File too large: %v bytes", len(filedata)))
		loadmapa[erro.ErrorType] = erro.ClientErrorType
		loadmapa[erro.ErrorMessage] = erro.ErrorMessage
		return &ServiceResponse{Success: false, Errors: loadmapa}
	}
	contentType := http.DetectContentType(filedata)
	if contentType != "image/jpeg" && contentType != "image/png" {
		use.kafkaProducer.NewPhotoLog(kafka.LogLevelWarn, place, traceid, fmt.Sprintf("Invalid file format: %s", contentType))
		loadmapa[erro.ErrorType] = erro.ClientErrorType
		loadmapa[erro.ErrorMessage] = erro.ErrorMessage
		return &ServiceResponse{Success: false, Errors: loadmapa}
	}
	photoid := uuid.New().String()
	var ext string
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	}
	go use.photoUnloadAndSave(ctx, filedata, photoid, ext, userid)
	return &ServiceResponse{Success: true, Data: map[string]any{repository.KeyPhotoID: photoid}}
}
func (use *PhotoService) GetPhoto(ctx context.Context, photoid string) *ServiceResponse {
	traceid := ctx.Value("traceID").(string)
	getphotomapa := make(map[string]string)
	bdresponse := use.repo.GetPhoto(ctx, photoid)
	if !bdresponse.Success && bdresponse.Errors != nil {
		if bdresponse.Errors.Type == erro.ClientErrorType {
			use.kafkaProducer.NewPhotoLog(kafka.LogLevelWarn, bdresponse.Place, traceid, bdresponse.Errors.Message)
			getphotomapa[erro.ErrorType] = erro.ClientErrorType
			getphotomapa[erro.ErrorMessage] = erro.ErrorMessage
			return &ServiceResponse{Success: false, Errors: getphotomapa}
		}
		use.kafkaProducer.NewPhotoLog(kafka.LogLevelError, bdresponse.Place, traceid, bdresponse.Errors.Message)
		getphotomapa[erro.ErrorType] = erro.ServerErrorType
		getphotomapa[erro.ErrorMessage] = erro.ErrorMessage
		return &ServiceResponse{Success: false, Errors: getphotomapa}
	}
	photo := bdresponse.Data[repository.KeyPhoto].(*model.Photo)
	grpcphoto := &pb.Photo{PhotoId: photo.ID, Url: photo.URL, CreatedAt: photo.CreatedAt.String()}
	return &ServiceResponse{Success: true, Data: map[string]any{repository.KeyPhoto: grpcphoto}}
}
func (use *PhotoService) GetPhotos(ctx context.Context, userid string) *ServiceResponse {
	traceid := ctx.Value("traceID").(string)
	bdresponse := use.repo.GetPhotos(ctx, userid)
	getphotosmapa := make(map[string]string)
	if !bdresponse.Success && bdresponse.Errors != nil {
		if bdresponse.Errors.Type == erro.ClientErrorType {
			getphotosmapa[erro.ErrorType] = erro.ClientErrorType
			getphotosmapa[erro.ErrorMessage] = erro.ErrorMessage
			return &ServiceResponse{Success: false, Errors: getphotosmapa}
		}
		use.kafkaProducer.NewPhotoLog(kafka.LogLevelError, bdresponse.Place, traceid, bdresponse.Errors.Message)
		getphotosmapa[erro.ErrorType] = erro.ServerErrorType
		getphotosmapa[erro.ErrorMessage] = erro.ErrorMessage
		return &ServiceResponse{Success: false, Errors: getphotosmapa}
	}
	photos := bdresponse.Data[repository.KeyPhoto].([]*model.Photo)
	grpcphotos := []*pb.Photo{}
	for _, p := range photos {
		grpcphoto := &pb.Photo{PhotoId: p.ID, Url: p.URL, CreatedAt: p.CreatedAt.String()}
		grpcphotos = append(grpcphotos, grpcphoto)
	}
	return &ServiceResponse{Success: true, Data: map[string]any{repository.KeyPhoto: grpcphotos}}
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
