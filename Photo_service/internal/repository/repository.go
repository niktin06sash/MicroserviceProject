package repository

import (
	"context"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
)

type DBPhotoRepos interface {
	LoadUserID(ctx context.Context, userid string)
	DeleteUserData(ctx context.Context, userid string)
	LoadPhoto(ctx context.Context, photo *model.Photo) *RepositoryResponse
	DeletePhoto(ctx context.Context, userid string, photoid string) *RepositoryResponse
	GetPhotos(ctx context.Context, userid string) *RepositoryResponse
	GetPhoto(ctx context.Context, userid string, photoid string) *RepositoryResponse
}
type RepositoryResponse struct {
	Success        bool
	SuccessMessage string
	Data           any
	Error          error
	Place          string
}
type Repository struct {
	DBPhotoRepos
}

func NewRepository(db *DBObject) *Repository {
	return &Repository{
		DBPhotoRepos: NewPhotoPostgresRepo(db),
	}
}
