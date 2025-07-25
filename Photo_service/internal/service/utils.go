package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository"
)

func (use *PhotoServiceImplement) requestToRepository(response *repository.RepositoryResponse, traceid string) (*repository.RepositoryResponse, *ServiceResponse) {
	if !response.Success && response.Errors != nil {
		var customError *erro.CustomError
		if errors.As(response.Errors, &customError) {
			switch customError.Type {
			case erro.ServerErrorType:
				use.Logproducer.NewPhotoLog(kafka.LogLevelError, response.Place, traceid, customError.Message)
				customError.Message = erro.PhotoServiceUnavalaible
				return response, &ServiceResponse{Success: false, Errors: customError}

			case erro.ClientErrorType:
				use.Logproducer.NewPhotoLog(kafka.LogLevelWarn, response.Place, traceid, customError.Message)
				return response, &ServiceResponse{Success: false, Errors: response.Errors}
			}
		}
	}
	use.Logproducer.NewPhotoLog(kafka.LogLevelInfo, response.Place, traceid, response.SuccessMessage)
	return response, nil
}
func (use *PhotoServiceImplement) unloadPhotoCloud(ctx context.Context, file []byte, photoid string, ext string, userid string, traceid string) {
	const place = UnloadPhotoCloud
	filename := photoid + ext
	tempFile := filepath.Join(os.TempDir(), filename)
	err := os.WriteFile(tempFile, file, 0644)
	if err != nil {
		use.Logproducer.NewPhotoLog(kafka.LogLevelError, place, traceid, fmt.Sprintf("Failed to create temp file: %v", err))
		return
	}
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			use.Logproducer.NewPhotoLog(kafka.LogLevelError, place, traceid, fmt.Sprintf("Failed to remove temp file: %v", err))
		}
	}()
	cloudresponse := use.Cloud.UploadFile(ctx, tempFile, photoid, ext)
	if !cloudresponse.Success && cloudresponse.Errors != nil {
		var customError *erro.CustomError
		if errors.As(cloudresponse.Errors, &customError) {
			use.Logproducer.NewPhotoLog(kafka.LogLevelError, cloudresponse.Place, traceid, customError.Message)
		}
	}
	use.Logproducer.NewPhotoLog(kafka.LogLevelInfo, cloudresponse.Place, traceid, cloudresponse.SuccessMessage)
	photo := cloudresponse.Data.Photo
	photo.UserID = userid
	bdresponse := use.Photorepo.LoadPhoto(ctx, photo)
	if !bdresponse.Success && bdresponse.Errors != nil {
		var customError *erro.CustomError
		if errors.As(bdresponse.Errors, &customError) {
			use.Logproducer.NewPhotoLog(kafka.LogLevelError, bdresponse.Place, traceid, customError.Message)
		}
		return
	}
	use.Logproducer.NewPhotoLog(kafka.LogLevelInfo, bdresponse.Place, traceid, bdresponse.SuccessMessage)
	cacheresponse := use.Cache.DeletePhotosCache(ctx, userid)
	if !cacheresponse.Success && cacheresponse.Errors != nil {
		var customError *erro.CustomError
		if errors.As(cacheresponse.Errors, &customError) {
			use.Logproducer.NewPhotoLog(kafka.LogLevelError, cacheresponse.Place, traceid, customError.Message)
		}
		return
	}
	use.Logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceid, cacheresponse.SuccessMessage)
	use.Logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceid, fmt.Sprintf("The photo(id = %s) has been successfully uploaded to cloud and database", photoid))

}
func (use *PhotoServiceImplement) deletePhotoCloud(ctx context.Context, photoid string, ext string, traceid string) {
	const place = DeletePhotoCloud
	cloudresponse := use.Cloud.DeleteFile(ctx, photoid, ext)
	if !cloudresponse.Success && cloudresponse.Errors != nil {
		var customError *erro.CustomError
		if errors.As(cloudresponse.Errors, &customError) {
			use.Logproducer.NewPhotoLog(kafka.LogLevelError, cloudresponse.Place, traceid, customError.Message)
		}
		return
	}
	use.Logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceid, fmt.Sprintf("The photo(id = %s) has been successfully deleted from cloud and database", photoid))
}
func (use *PhotoServiceImplement) parsingIDs(id string, traceid string, place string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse Error: %v", err)
		use.Logproducer.NewPhotoLog(kafka.LogLevelWarn, place, traceid, fmterr)
		return err
	}
	return nil
}
func (use *PhotoServiceImplement) enqueueTask(ctx context.Context, task func(context.Context), timeout time.Duration, place string, traceid string) *ServiceResponse {
	select {
	case <-use.Worker.ctx.Done():
		use.Logproducer.NewPhotoLog(kafka.LogLevelError, place, traceid, erro.ContextCanceled)
		return &ServiceResponse{Success: false, Errors: erro.ServerError(erro.PhotoServiceUnavalaible)}
	case use.Worker.task_queue <- func() {
		taskCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		task(taskCtx)
	}:
		return &ServiceResponse{Success: true}
	case <-ctx.Done():
		use.Logproducer.NewPhotoLog(kafka.LogLevelError, place, traceid, erro.ContextCanceled)
		return &ServiceResponse{Success: false, Errors: erro.ServerError(erro.PhotoServiceUnavalaible)}
	default:
		use.Logproducer.NewPhotoLog(kafka.LogLevelError, place, traceid, erro.ErrorOverflowTaskQ)
		return &ServiceResponse{Success: false, Errors: erro.ServerError(erro.PhotoServiceUnavalaible)}
	}
}
