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

const (
	insertPhotoQuery  = `INSERT INTO photos (photoid, userid, url, size, content_type, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	deletePhotoQuery  = `DELETE FROM photos WHERE photoid = $1 AND userid = $2 RETURNING content_type`
	selectPhotosQuery = `SELECT photoid, url, content_type, created_at FROM photos WHERE userid = $1`
	selectPhotoQuery  = `SELECT photoid, url, content_type, created_at FROM photos WHERE userid = $1 AND photoid = $2`
	insertUserQuery   = "INSERT INTO usersid (userid) VALUES ($1) ON CONFLICT (userid) DO NOTHING"
	deleteUserQuery   = "DELETE FROM usersid WHERE userid = $1"
)

func (ph *PhotoPostgresRepo) LoadPhoto(ctx context.Context, photo *model.Photo) *RepositoryResponse {
	const place = LoadPhoto
	err := ph.Db.mapstmt[insertPhotoQuery].QueryRowContext(ctx, photo.ID, photo.UserID, photo.URL, photo.Size, photo.ContentType, photo.CreatedAt).Err()
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyPhotoTable, err)
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyPhotoID: photo.ID}, Place: place, SuccessMessage: "Successful load photo metadata to database"}
}
func (ph *PhotoPostgresRepo) DeletePhoto(ctx context.Context, userid string, photoid string) *RepositoryResponse {
	const place = DeletePhoto
	var content_type string
	err := ph.Db.mapstmt[deletePhotoQuery].QueryRowContext(ctx, photoid, userid).Scan(&content_type)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: "Attempt to delete someone else's photo"}, Place: place}
		}
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyPhotoTable, err)
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, Place: place, Data: map[string]any{KeyContentType: content_type}, SuccessMessage: "Successful delete photo metadata from database"}
}
func (ph *PhotoPostgresRepo) GetPhotos(ctx context.Context, userid string) *RepositoryResponse {
	const place = GetPhotos
	photoslice := make([]*model.Photo, 0)
	rows, err := ph.Db.mapstmt[selectPhotosQuery].QueryContext(ctx, userid)
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
			fmterr := fmt.Sprintf("Error after scan photo data from %s: %v", KeyPhotoTable, err)
			return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
		}
		photoslice = append(photoslice, &photo)
	}
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyPhoto: photoslice}, Place: place, SuccessMessage: "Successful get photos metadata from database"}
}
func (ph *PhotoPostgresRepo) GetPhoto(ctx context.Context, photoid string, userid string) *RepositoryResponse {
	const place = GetPhoto
	var photo model.Photo
	err := ph.Db.mapstmt[selectPhotoQuery].QueryRowContext(ctx, photoid, userid).Scan(&photo.ID, &photo.URL, &photo.ContentType, &photo.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: "A non-existent data has been entered"}, Place: place}
		}
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyPhotoTable, err)
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyPhoto: photo}, Place: place, SuccessMessage: "Successful get photo metadata from database"}
}

func (ph *PhotoPostgresRepo) AddUserId(ctx context.Context, userid string) *RepositoryResponse {
	const place = AddUserId
	_, err := ph.Db.mapstmt[insertUserQuery].ExecContext(ctx, userid)
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUsersIdTable, err)
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, Place: place, SuccessMessage: "Successful add userID to Photo-Service's database after registration"}
}
func (ph *PhotoPostgresRepo) DeleteUserData(ctx context.Context, userid string) *RepositoryResponse {
	const place = DeleteUserData
	err := ph.Db.mapstmt[deleteUserQuery].QueryRowContext(ctx, userid)
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUsersIdTable, err)
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, Place: place, SuccessMessage: "Successful delete userdata from Photo-Service's database after delete account"}
}
