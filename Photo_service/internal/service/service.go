package service

import (
	"context"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository"
	pb "github.com/niktin06sash/MicroserviceProject/Photo_service/proto"
)

type DBPhotoRepos interface {
	LoadPhoto(ctx context.Context, photo *model.Photo) *repository.RepositoryResponse
	DeletePhoto(ctx context.Context, userid string, photoid string) *repository.RepositoryResponse
	GetPhotos(ctx context.Context, userid string) *repository.RepositoryResponse
	GetPhoto(ctx context.Context, userid string, photoid string) *repository.RepositoryResponse
	AddUserId(ctx context.Context, userid string) *repository.RepositoryResponse
	DeleteUserData(ctx context.Context, userid string) *repository.RepositoryResponse
}
type CachePhotoRepos interface {
	AddPhotoCache(ctx context.Context, userid string, photo *pb.Photo) *repository.RepositoryResponse
	AddPhotosCache(ctx context.Context, userid string, photosslice []*pb.Photo) *repository.RepositoryResponse
	GetPhotoCache(ctx context.Context, userid string, photoid string) *repository.RepositoryResponse
	GetPhotosCache(ctx context.Context, userid string) *repository.RepositoryResponse
	DeletePhotoCache(ctx context.Context, userid string, photoid string) *repository.RepositoryResponse
	DeletePhotosCache(ctx context.Context, userid string) *repository.RepositoryResponse
}
type CloudPhotoStorage interface {
	UploadFile(ctx context.Context, localfilepath string, photoid string, ext string) *repository.RepositoryResponse
	DeleteFile(ctx context.Context, id, contenttype string) *repository.RepositoryResponse
}
type LogProducer interface {
	NewPhotoLog(level, place, traceid, msg string)
}

const UseCase_LoadPhoto = "UseCase-LoadPhoto"
const UseCase_DeletePhoto = "UseCase-DeletePhoto"
const UseCase_DeleteAllUserData = "UseCase-DeleteAllUserData"
const MaxFileSize = 10 << 20
const UnloadPhotoCloud = "UnloadPhotoCloud"
const DeletePhotoCloud = "DeletePhotoCloud"
const UseCase_GetPhotos = "UseCase-GetPhotos"
const UseCase_GetPhoto = "UseCase-GetPhoto"

type ServiceResponse struct {
	Success bool
	Data    Data
	Errors  *erro.CustomError
}
type Data struct {
	PhotoID string
	Photo   *pb.Photo
	Photos  []*pb.Photo
}
