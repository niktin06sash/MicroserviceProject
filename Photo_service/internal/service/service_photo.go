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

type PhotoService struct {
	repo        DBPhotoRepos
	cloud       CloudPhotoStorage
	logProducer LogProducer
}

func NewPhotoService(repo DBPhotoRepos, cloud CloudPhotoStorage, logproducer LogProducer) *PhotoService {
	return &PhotoService{repo: repo, logProducer: logproducer, cloud: cloud}
}

func (use *PhotoService) DeletePhoto(ctx context.Context, userid string, photoid string) *ServiceResponse {
	traceid := ctx.Value("traceID").(string)
	bdresponse := use.repo.DeletePhoto(ctx, userid, photoid)
	if bdresponse.Errors != nil {
		if bdresponse.Errors[erro.ErrorType] == erro.ClientErrorType {
			use.logProducer.NewPhotoLog(kafka.LogLevelWarn, bdresponse.Place, traceid, bdresponse.Errors[erro.ErrorMessage])
			return &ServiceResponse{Success: false, Errors: bdresponse.Errors}
		}
		use.logProducer.NewPhotoLog(kafka.LogLevelError, bdresponse.Place, traceid, bdresponse.Errors[erro.ErrorMessage])
		return &ServiceResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.PhotoServiceUnavalaible}}
	}
	use.logProducer.NewPhotoLog(kafka.LogLevelInfo, bdresponse.Place, traceid, bdresponse.SuccessMessage)
	ext := bdresponse.Data[repository.KeyContentType].(string)
	go use.deletePhotoCloud(ctx, photoid, ext)
	return &ServiceResponse{Success: true}
}
func (use *PhotoService) LoadPhoto(ctx context.Context, userid string, filedata []byte) *ServiceResponse {
	const place = UseCase_LoadPhoto
	traceid := ctx.Value("traceID").(string)
	_, err := uuid.Parse(userid)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse Error: %v", err)
		use.logProducer.NewPhotoLog(kafka.LogLevelWarn, place, traceid, fmterr)
		return &ServiceResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.InvalidUserIDFormat}}
	}
	if len(filedata) > MaxFileSize {
		use.logProducer.NewPhotoLog(kafka.LogLevelWarn, place, traceid, fmt.Sprintf("File too large: %v bytes", len(filedata)))
		return &ServiceResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: "File too large - max 10 MB"}}
	}
	contentType := http.DetectContentType(filedata)
	if contentType != "image/jpeg" && contentType != "image/png" {
		use.logProducer.NewPhotoLog(kafka.LogLevelWarn, place, traceid, fmt.Sprintf("Invalid file format: %s", contentType))
		return &ServiceResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: "Invalid file format"}}
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
	return &ServiceResponse{Success: true, Data: Data{PhotoID: photoid}}
}
func (use *PhotoService) GetPhoto(ctx context.Context, photoid string) *ServiceResponse {
	traceid := ctx.Value("traceID").(string)
	bdresponse := use.repo.GetPhoto(ctx, photoid)
	if !bdresponse.Success && bdresponse.Errors != nil {
		if bdresponse.Errors[erro.ErrorType] == erro.ClientErrorType {
			use.logProducer.NewPhotoLog(kafka.LogLevelWarn, bdresponse.Place, traceid, bdresponse.Errors[erro.ErrorMessage])
			return &ServiceResponse{Success: false, Errors: bdresponse.Errors}
		}
		use.logProducer.NewPhotoLog(kafka.LogLevelError, bdresponse.Place, traceid, bdresponse.Errors[erro.ErrorMessage])
		return &ServiceResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.PhotoServiceUnavalaible}}
	}
	use.logProducer.NewPhotoLog(kafka.LogLevelInfo, bdresponse.Place, traceid, bdresponse.SuccessMessage)
	photo := bdresponse.Data[repository.KeyPhoto].(*model.Photo)
	grpcphoto := &pb.Photo{PhotoId: photo.ID, Url: photo.URL, CreatedAt: photo.CreatedAt.String()}
	return &ServiceResponse{Success: true, Data: Data{Photo: grpcphoto}}
}
func (use *PhotoService) GetPhotos(ctx context.Context, userid string) *ServiceResponse {
	traceid := ctx.Value("traceID").(string)
	const place = UseCase_GetPhotos
	_, err := uuid.Parse(userid)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse Error: %v", err)
		use.logProducer.NewPhotoLog(kafka.LogLevelWarn, place, traceid, fmterr)
		return &ServiceResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.InvalidUserIDFormat}}
	}
	bdresponse := use.repo.GetPhotos(ctx, userid)
	if !bdresponse.Success && bdresponse.Errors != nil {
		if bdresponse.Errors[erro.ErrorType] == erro.ClientErrorType {
			use.logProducer.NewPhotoLog(kafka.LogLevelWarn, bdresponse.Place, traceid, bdresponse.Errors[erro.ErrorMessage])
			return &ServiceResponse{Success: false, Errors: bdresponse.Errors}
		}
		use.logProducer.NewPhotoLog(kafka.LogLevelError, bdresponse.Place, traceid, bdresponse.Errors[erro.ErrorMessage])
		return &ServiceResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.PhotoServiceUnavalaible}}
	}
	use.logProducer.NewPhotoLog(kafka.LogLevelInfo, bdresponse.Place, traceid, bdresponse.SuccessMessage)
	photos := bdresponse.Data[repository.KeyPhoto].([]*model.Photo)
	grpcphotos := []*pb.Photo{}
	for _, p := range photos {
		grpcphoto := &pb.Photo{PhotoId: p.ID, Url: p.URL, CreatedAt: p.CreatedAt.String()}
		grpcphotos = append(grpcphotos, grpcphoto)
	}
	return &ServiceResponse{Success: true, Data: Data{Photos: grpcphotos}}
}
func (use *PhotoService) photoUnloadAndSave(ctx context.Context, file []byte, photoid string, ext string, userid string) {
	const place = PhotoUnloadAndSave
	traceid := ctx.Value("traceID").(string)
	filename := photoid + ext
	tmpDir := os.TempDir()
	tempFile := filepath.Join(tmpDir, filename)
	err := os.WriteFile(tempFile, file, 0644)
	if err != nil {
		use.logProducer.NewPhotoLog(kafka.LogLevelError, place, traceid, fmt.Sprintf("Failed to create temp file: %v", err))
		return
	}
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			use.logProducer.NewPhotoLog(kafka.LogLevelError, place, traceid, fmt.Sprintf("Failed to remove temp file: %v", err))
		}
	}()
	select {
	case <-ctx.Done():
		use.logProducer.NewPhotoLog(kafka.LogLevelError, place, traceid, "Context canceled or timeout")
		return
	default:
		cloudresponse := use.cloud.UploadFile(tempFile, photoid, ext)
		if !cloudresponse.Success && cloudresponse.Errors != nil {
			use.logProducer.NewPhotoLog(kafka.LogLevelError, cloudresponse.Place, traceid, cloudresponse.Errors[erro.ErrorMessage])
			return
		}
		use.logProducer.NewPhotoLog(kafka.LogLevelInfo, cloudresponse.Place, traceid, cloudresponse.SuccessMessage)
		photo := cloudresponse.Data[repository.KeyPhoto].(*model.Photo)
		photo.UserID = userid
		bdresponse := use.repo.LoadPhoto(ctx, photo)
		if !bdresponse.Success && bdresponse.Errors != nil {
			use.logProducer.NewPhotoLog(kafka.LogLevelError, bdresponse.Place, traceid, bdresponse.Errors[erro.ErrorMessage])
			return
		}
		use.logProducer.NewPhotoLog(kafka.LogLevelInfo, bdresponse.Place, traceid, bdresponse.SuccessMessage)
		use.logProducer.NewPhotoLog(kafka.LogLevelInfo, place, traceid, fmt.Sprintf("The photo(id = %s) has been successfully uploaded to cloud and database", photoid))
	}
}
func (use *PhotoService) deletePhotoCloud(ctx context.Context, photoid string, ext string) {
	const place = DeletePhotoCloud
	traceid := ctx.Value("traceID").(string)
	select {
	case <-ctx.Done():
		use.logProducer.NewPhotoLog(kafka.LogLevelError, place, traceid, "Context canceled or timeout")
		return
	default:
		cloudresponse := use.cloud.DeleteFile(photoid, ext)
		if !cloudresponse.Success && cloudresponse.Errors != nil {
			use.logProducer.NewPhotoLog(kafka.LogLevelError, cloudresponse.Place, traceid, cloudresponse.Errors[erro.ErrorMessage])
			return
		}
		use.logProducer.NewPhotoLog(kafka.LogLevelInfo, place, traceid, fmt.Sprintf("The photo(id = %s) has been successfully deleted from cloud and database", photoid))
	}
}
