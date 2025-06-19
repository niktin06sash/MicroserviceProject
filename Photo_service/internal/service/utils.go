package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository"
)

func (use *PhotoService) requestToDB(response *repository.RepositoryResponse, traceid string) (*repository.RepositoryResponse, *ServiceResponse) {
	if !response.Success && response.Errors != nil {
		switch response.Errors[erro.ErrorType] {
		case erro.ServerErrorType:
			use.logProducer.NewPhotoLog(kafka.LogLevelError, response.Place, traceid, response.Errors[erro.ErrorMessage])
			response.Errors[erro.ErrorMessage] = erro.PhotoServiceUnavalaible
			return response, &ServiceResponse{Success: false, Errors: response.Errors}

		case erro.ClientErrorType:
			use.logProducer.NewPhotoLog(kafka.LogLevelWarn, response.Place, traceid, response.Errors[erro.ErrorMessage])
			return response, &ServiceResponse{Success: false, Errors: response.Errors}
		}
	}
	use.logProducer.NewPhotoLog(kafka.LogLevelInfo, response.Place, traceid, response.SuccessMessage)
	return response, nil
}
func (use *PhotoService) unloadPhotoCloud(file []byte, photoid string, ext string, userid string, traceid string) {
	const place = UnloadPhotoCloud
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
	cloudctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cloudresponse := use.cloud.UploadFile(cloudctx, tempFile, photoid, ext)
	if !cloudresponse.Success && cloudresponse.Errors != nil {
		use.logProducer.NewPhotoLog(kafka.LogLevelError, cloudresponse.Place, traceid, cloudresponse.Errors[erro.ErrorMessage])
		return
	}
	use.logProducer.NewPhotoLog(kafka.LogLevelInfo, cloudresponse.Place, traceid, cloudresponse.SuccessMessage)
	photo := cloudresponse.Data[repository.KeyPhoto].(*model.Photo)
	photo.UserID = userid
	databasectx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	bdresponse := use.repo.LoadPhoto(databasectx, photo)
	if !bdresponse.Success && bdresponse.Errors != nil {
		use.logProducer.NewPhotoLog(kafka.LogLevelError, bdresponse.Place, traceid, bdresponse.Errors[erro.ErrorMessage])
		return
	}
	use.logProducer.NewPhotoLog(kafka.LogLevelInfo, bdresponse.Place, traceid, bdresponse.SuccessMessage)
	use.logProducer.NewPhotoLog(kafka.LogLevelInfo, place, traceid, fmt.Sprintf("The photo(id = %s) has been successfully uploaded to cloud and database", photoid))

}
func (use *PhotoService) deletePhotoCloud(photoid string, ext string, traceid string) {
	const place = DeletePhotoCloud
	newctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cloudresponse := use.cloud.DeleteFile(newctx, photoid, ext)
	if !cloudresponse.Success && cloudresponse.Errors != nil {
		use.logProducer.NewPhotoLog(kafka.LogLevelError, cloudresponse.Place, traceid, cloudresponse.Errors[erro.ErrorMessage])
		return
	}
	use.logProducer.NewPhotoLog(kafka.LogLevelInfo, place, traceid, fmt.Sprintf("The photo(id = %s) has been successfully deleted from cloud and database", photoid))
}
func (use *PhotoService) WaitGoroutines() {
	use.wg.Wait()
}
