package repository

import (
	"context"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
)

type PhotoPostgresRepo struct {
	Db *DBObject
}

func NewPhotoPostgresRepo(db *DBObject) *PhotoPostgresRepo {
	return &PhotoPostgresRepo{Db: db}
}

func (ph *PhotoPostgresRepo) LoadUserID(ctx context.Context, userid string) {

}
func (ph *PhotoPostgresRepo) DeleteUserData(ctx context.Context, userid string) {

}
func (ph *PhotoPostgresRepo) LoadPhoto(ctx context.Context, photo *model.Photo) *RepositoryResponse {

}
func (ph *PhotoPostgresRepo) DeletePhoto(ctx context.Context, userid string, photoid string) *RepositoryResponse {

}
func (ph *PhotoPostgresRepo) GetPhotos(ctx context.Context, userid string) *RepositoryResponse {

}
func (ph *PhotoPostgresRepo) GetPhoto(ctx context.Context, userid string, photoid string) *RepositoryResponse {

}
