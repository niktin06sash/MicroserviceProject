package service

import (
	"context"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository"
	pb "github.com/niktin06sash/MicroserviceProject/Photo_service/proto"
)

type DBPhotoRepos interface {
	LoadPhoto(ctx context.Context, photo *model.Photo) *repository.RepositoryResponse
	DeletePhoto(ctx context.Context, userid string, photoid string) *repository.RepositoryResponse
	GetPhotos(ctx context.Context, userid string) *repository.RepositoryResponse
	GetPhoto(ctx context.Context, userid string, photoid string) *repository.RepositoryResponse
}
type CloudPhotoStorage interface {
	UploadFile(ctx context.Context, localfilepath string, photoid string, ext string) *repository.RepositoryResponse
	DeleteFile(ctx context.Context, id, contenttype string) *repository.RepositoryResponse
}
type LogProducer interface {
	NewPhotoLog(level, place, traceid, msg string)
}

const UseCase_LoadPhoto = "UseCase_LoadPhoto"
const MaxFileSize = 10 << 20
const UnloadPhotoCloud = "UnloadPhotoCloud"
const DeletePhotoCloud = "DeletePhotoCloud"
const UseCase_GetPhotos = "UseCase_GetPhotos"
const UseCase_GetPhoto = "UseCase_GetPhoto"

type ServiceResponse struct {
	Success bool
	Data    Data
	Errors  map[string]string
}
type Data struct {
	PhotoID string
	Photo   *pb.Photo
	Photos  []*pb.Photo
}
