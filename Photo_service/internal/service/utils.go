package service

import (
	"context"
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
		switch response.Errors.Type {
		case erro.ServerErrorType:
			use.logproducer.NewPhotoLog(kafka.LogLevelError, response.Place, traceid, response.Errors.Message)
			response.Errors.Message = erro.PhotoServiceUnavalaible
			return response, &ServiceResponse{Success: false, Errors: response.Errors}

		case erro.ClientErrorType:
			use.logproducer.NewPhotoLog(kafka.LogLevelWarn, response.Place, traceid, response.Errors.Message)
			return response, &ServiceResponse{Success: false, Errors: response.Errors}
		}
	}
	use.logproducer.NewPhotoLog(kafka.LogLevelInfo, response.Place, traceid, response.SuccessMessage)
	return response, nil
}
func (use *PhotoServiceImplement) unloadPhotoCloud(ctx context.Context, file []byte, photoid string, ext string, userid string, traceid string) {
	const place = UnloadPhotoCloud
	filename := photoid + ext
	tmpDir := os.TempDir()
	tempFile := filepath.Join(tmpDir, filename)
	err := os.WriteFile(tempFile, file, 0644)
	if err != nil {
		use.logproducer.NewPhotoLog(kafka.LogLevelError, place, traceid, fmt.Sprintf("Failed to create temp file: %v", err))
		return
	}
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			use.logproducer.NewPhotoLog(kafka.LogLevelError, place, traceid, fmt.Sprintf("Failed to remove temp file: %v", err))
		}
	}()
	cloudresponse := use.cloud.UploadFile(ctx, tempFile, photoid, ext)
	if !cloudresponse.Success && cloudresponse.Errors != nil {
		use.logproducer.NewPhotoLog(kafka.LogLevelError, cloudresponse.Place, traceid, cloudresponse.Errors.Message)
		return
	}
	use.logproducer.NewPhotoLog(kafka.LogLevelInfo, cloudresponse.Place, traceid, cloudresponse.SuccessMessage)
	photo := cloudresponse.Data.Photo
	photo.UserID = userid
	bdresponse := use.photorepo.LoadPhoto(ctx, photo)
	if !bdresponse.Success && bdresponse.Errors != nil {
		use.logproducer.NewPhotoLog(kafka.LogLevelError, bdresponse.Place, traceid, bdresponse.Errors.Message)
		return
	}
	use.logproducer.NewPhotoLog(kafka.LogLevelInfo, bdresponse.Place, traceid, bdresponse.SuccessMessage)
	use.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceid, fmt.Sprintf("The photo(id = %s) has been successfully uploaded to cloud and database", photoid))

}
func (use *PhotoServiceImplement) deletePhotoCloud(ctx context.Context, photoid string, ext string, traceid string) {
	const place = DeletePhotoCloud
	cloudresponse := use.cloud.DeleteFile(ctx, photoid, ext)
	if !cloudresponse.Success && cloudresponse.Errors != nil {
		use.logproducer.NewPhotoLog(kafka.LogLevelError, cloudresponse.Place, traceid, cloudresponse.Errors.Message)
		return
	}
	use.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, traceid, fmt.Sprintf("The photo(id = %s) has been successfully deleted from cloud and database", photoid))
}
func (use *PhotoServiceImplement) parsingIDs(id string, traceid string, place string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse Error: %v", err)
		use.logproducer.NewPhotoLog(kafka.LogLevelWarn, place, traceid, fmterr)
		return err
	}
	return nil
}
func (use *PhotoServiceImplement) enqueueTask(ctx context.Context, task func(context.Context), timeout time.Duration, place string, traceid string) *ServiceResponse {
	select {
	case use.task_queue <- func() {
		taskCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		task(taskCtx)
	}:
		return &ServiceResponse{Success: true}
	case <-ctx.Done():
		use.logproducer.NewPhotoLog(kafka.LogLevelError, place, traceid, erro.ContextCanceled)
		return &ServiceResponse{Success: false, Errors: erro.ServerError(erro.PhotoServiceUnavalaible)}

	default:
		use.logproducer.NewPhotoLog(kafka.LogLevelError, place, traceid, erro.ErrorOverflowTaskQ)
		return &ServiceResponse{Success: false, Errors: erro.ServerError(erro.PhotoServiceUnavalaible)}
	}
}
