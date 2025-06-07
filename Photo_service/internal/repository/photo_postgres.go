package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
)

type PhotoPostgresRepo struct {
	Db *DBObject
}

func NewPhotoPostgresRepo(db *DBObject) *PhotoPostgresRepo {
	return &PhotoPostgresRepo{Db: db}
}

func (ph *PhotoPostgresRepo) LoadPhoto(ctx context.Context, photo *model.Photo) *RepositoryResponse {
	const place = LoadPhoto
	err := ph.Db.DB.QueryRowContext(ctx,
		fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s) VALUES ($1, $2, $3, $4, $5, $6)",
			KeyPhotoTable, KeyPhotoID, KeyUserID, KeyPhotoURL, KeyPhotoSize, KeyContentType, KeyCreatedTime),
		photo.ID, photo.UserID, photo.URL, photo.Size, photo.ContentType, photo.CreatedAt).Err()
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyPhotoTable, err)
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyPhotoID: photo.ID}, Place: place, SuccessMessage: "Successful load photo"}
}
func (ph *PhotoPostgresRepo) DeletePhoto(ctx context.Context, userid string, photoid string) *RepositoryResponse {
	const place = DeletePhoto
	var content_type string
	err := ph.Db.DB.QueryRowContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = $1 AND %s = $2 RETURNING %s", KeyPhotoTable, KeyPhotoID, KeyUserID, KeyContentType), photoid, userid).Scan(&content_type)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: "Attempt to delete someone else's photo"}, Place: place}
		}
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyPhotoTable, err)
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, Place: place, Data: map[string]any{KeyContentType: content_type}, SuccessMessage: "Successful delete photo"}
}
func (ph *PhotoPostgresRepo) GetPhotos(ctx context.Context, userid string) *RepositoryResponse {
	const place = GetPhotos
	photoslice := make([]*model.Photo, 0)
	rows, err := ph.Db.DB.QueryContext(ctx, fmt.Sprintf("SELECT %s, %s, %s, %s FROM %s WHERE %s = $1", KeyPhotoID, KeyPhotoURL, KeyContentType, KeyCreatedTime, KeyPhotoTable, KeyUserID), userid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: "Unregistered userid has been entered"}, Place: place}
		}
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyPhotoTable, err)
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	defer rows.Close()
	for rows.Next() {
		var photo model.Photo
		err := rows.Scan(&photo.ID, &photo.URL, &photo.ContentType, &photo.CreatedAt)
		if err != nil {
			fmterr := fmt.Sprintf("Error after request into %s: %v", KeyPhotoTable, err)
			return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
		}
		photoslice = append(photoslice, &photo)
	}
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyPhoto: photoslice}, Place: place, SuccessMessage: "Successful get photos"}
}
func (ph *PhotoPostgresRepo) GetPhoto(ctx context.Context, photoid string) *RepositoryResponse {
	const place = GetPhoto
	var photo model.Photo
	err := ph.Db.DB.QueryRowContext(ctx, fmt.Sprintf("SELECT %s, %s, %s, %s FROM %s WHERE %s = $1", KeyPhotoID, KeyPhotoURL, KeyContentType, KeyCreatedTime, KeyPhotoTable, KeyPhotoID), photoid).Scan(&photo.ID, &photo.URL, &photo.ContentType, &photo.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: "A non-existent photoid has been entered"}, Place: place}
		}
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyPhotoTable, err)
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyPhoto: photo}, Place: place, SuccessMessage: "Successful get photo"}
}

func (ph *PhotoPostgresRepo) AddUserId(ctx context.Context, userid string) *RepositoryResponse {
	return &RepositoryResponse{}
}

func (ph *PhotoPostgresRepo) DeleteUserData(ctx context.Context, userid string) *RepositoryResponse {
	return &RepositoryResponse{}
}
