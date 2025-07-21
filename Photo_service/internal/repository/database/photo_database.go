package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository"
)

type PhotoDatabase struct {
	databaseclient *DBObject
}

func NewPhotoDatabase(db *DBObject) *PhotoDatabase {
	return &PhotoDatabase{databaseclient: db}
}

const (
	insertPhotoQuery       = `INSERT INTO photos (photoid, userid, url, size, content_type, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	deletePhotoQuery       = `DELETE FROM photos WHERE photoid = $1 AND userid = $2`
	selectPhotosQuery      = `SELECT photoid, url, content_type, created_at FROM photos WHERE userid = $1`
	selectPhotoQuery       = `SELECT photoid, url, content_type, created_at FROM photos WHERE userid = $1 AND photoid = $2`
	selectContentTypeQuery = `SELECT content_type FROM photos WHERE photoid = $1`
	insertUserQuery        = "INSERT INTO usersid (userid) VALUES ($1) ON CONFLICT (userid) DO NOTHING"
	deleteUserQuery        = "DELETE FROM usersid WHERE userid = $1"
)

func (ph *PhotoDatabase) LoadPhoto(ctx context.Context, photo *model.Photo) *repository.RepositoryResponse {
	const place = LoadPhoto
	_, err := ph.databaseclient.mapstmt[insertPhotoQuery].ExecContext(ctx, photo.ID, photo.UserID, photo.URL, photo.Size, photo.ContentType, photo.CreatedAt)
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqPhotos, err)), place)
	}
	return &repository.RepositoryResponse{Success: true, Place: place, SuccessMessage: "Successful load photo metadata to database"}
}
func (ph *PhotoDatabase) DeletePhoto(ctx context.Context, userid string, photoid string) *repository.RepositoryResponse {
	const place = DeletePhoto
	var content_type string
	err := ph.databaseclient.mapstmt[selectContentTypeQuery].QueryRowContext(ctx, photoid).Scan(&content_type)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.BadResponse(erro.ClientError(erro.NonExistentData), place)
		}
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqPhotos, err)), place)
	}
	result, err := ph.databaseclient.mapstmt[deletePhotoQuery].ExecContext(ctx, photoid, userid)
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqPhotos, err)), place)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqPhotos, err)), place)
	}
	if rowsAffected == 0 {
		return repository.BadResponse(erro.ClientError(erro.DeleteSomeonePhoto), place)
	}
	return &repository.RepositoryResponse{Success: true, Place: place, Data: repository.Data{ContentType: content_type},
		SuccessMessage: "Successful delete photo metadata from database"}
}
func (ph *PhotoDatabase) GetPhotos(ctx context.Context, userid string) *repository.RepositoryResponse {
	const place = GetPhotos
	photoslice := make([]*model.Photo, 0)
	rows, err := ph.databaseclient.mapstmt[selectPhotosQuery].QueryContext(ctx, userid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.BadResponse(erro.ClientError(erro.UnregisteredUserID), place)
		}
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqPhotos, err)), place)
	}
	defer rows.Close()
	for rows.Next() {
		var photo model.Photo
		err := rows.Scan(&photo.ID, &photo.URL, &photo.ContentType, &photo.CreatedAt)
		if err != nil {
			return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqPhotos, err)), place)
		}
		photoslice = append(photoslice, &photo)
	}
	return &repository.RepositoryResponse{Success: true, Data: repository.Data{Photos: photoslice}, Place: place, SuccessMessage: "Successful get photos metadata from database"}
}
func (ph *PhotoDatabase) GetPhoto(ctx context.Context, userid string, photoid string) *repository.RepositoryResponse {
	const place = GetPhoto
	var photo model.Photo
	err := ph.databaseclient.mapstmt[selectPhotoQuery].QueryRowContext(ctx, userid, photoid).Scan(&photo.ID, &photo.URL, &photo.ContentType, &photo.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.BadResponse(erro.ClientError(erro.NonExistentData), place)
		}
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqPhotos, err)), place)
	}
	return &repository.RepositoryResponse{Success: true, Data: repository.Data{Photo: &photo}, Place: place, SuccessMessage: "Successful get photo metadata from database"}
}
func (ph *PhotoDatabase) AddUserId(ctx context.Context, userid string) *repository.RepositoryResponse {
	const place = AddUserId
	result, err := ph.databaseclient.mapstmt[insertUserQuery].ExecContext(ctx, userid)
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqPhotos, err)), place)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqPhotos, err)), place)
	}
	var message string
	if rowsAffected > 0 {
		message = "Successful add userID to Photo-Service's database after registration"
	} else {
		message = "UserID already exists (no changes made)"
	}
	return &repository.RepositoryResponse{Success: true, Place: place, SuccessMessage: message}
}

func (ph *PhotoDatabase) DeleteUserData(ctx context.Context, userid string) *repository.RepositoryResponse {
	const place = DeleteUserData
	result, err := ph.databaseclient.mapstmt[deleteUserQuery].ExecContext(ctx, userid)
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqPhotos, err)), place)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqPhotos, err)), place)
	}
	var message string
	if rowsAffected > 0 {
		message = "Successful delete userdata from Photo-Service's database after delete account"
	} else {
		message = "No user data found to delete"
	}
	return &repository.RepositoryResponse{Success: true, Place: place, SuccessMessage: message}
}
