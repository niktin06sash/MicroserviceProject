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
	GetPhoto(ctx context.Context, photoid string) *repository.RepositoryResponse
}
type PhotoService struct {
	repo          DBPhotoRepos
	kafkaProducer LogProducer
}

func NewPhotoService(repo DBPhotoRepos, kafka LogProducer) *PhotoService {
	return &PhotoService{repo: repo, kafkaProducer: kafka}
}

const UseCase_LoadPhoto = "UseCase_LoadPhoto"
const UseCase_DeletePhoto = "UseCase_DeletePhoto"
const UseCase_GetPhotos = "UseCase_GetPhotos"
const UseCase_GetPhoto = "UseCase_GetPhoto"

func (use *PhotoService) DeletePhoto(ctx context.Context, req *pb.DeletePhotoRequest) (*pb.DeletePhotoResponse, error) {

}
func (use *PhotoService) LoadPhoto(ctx context.Context, req *pb.LoadPhotoRequest) (*pb.LoadPhotoResponse, error) {

}
func (user *PhotoService) GetPhoto(ctx context.Context, req *pb.GetPhotoRequest) (*pb.GetPhotoResponse, error) {

}
func (user *PhotoService) GetPhotos(ctx context.Context, req *pb.GetPhotosRequest) (*pb.GetPhotosResponse, error) {

}
