package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
)

type PhotoRedisRepo struct {
	Client *RedisObject
}

func NewPhotoRedisRepo(red *RedisObject) *PhotoRedisRepo {
	return &PhotoRedisRepo{Client: red}
}

const KeyPhoto = "photo:%s:%s"
const KeyPhotos = "photos:%s"

func (ph *PhotoRedisRepo) AddPhotoCache(ctx context.Context, userid string, photo *model.Photo) *RepositoryResponse {
	const place = AddPhotoCache
	jsondata, err := json.Marshal(photo)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorMarshal, err)), Place: place}
	}
	err = ph.Client.RedisClient.Set(ctx, fmt.Sprintf(KeyPhoto, userid, photo.ID), jsondata, 1*time.Hour).Err()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorSetPhotos, err)), Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful add photo in cache", Place: place}
}
func (ph *PhotoRedisRepo) AddPhotosCache(ctx context.Context, userid string, photosslice []*model.Photo) *RepositoryResponse {
	const place = AddPhotosCache
	jsondata_photos, err := json.Marshal(photosslice)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorMarshal, err)), Place: place}
	}
	pipe := ph.Client.RedisClient.TxPipeline()
	defer pipe.Discard()
	pipe.Set(ctx, fmt.Sprintf(KeyPhotos, userid), jsondata_photos, 1*time.Hour)
	for _, photo := range photosslice {
		jsondata_photo, err := json.Marshal(photo)
		if err != nil {
			return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorMarshal, err)), Place: place}
		}
		pipe.Set(ctx, fmt.Sprintf(KeyPhoto, userid, photo.ID), jsondata_photo, 1*time.Hour)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorPipeExec, err)), Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful add photos in cache", Place: place}
}
func (ph *PhotoRedisRepo) GetPhotoCache(ctx context.Context, userid string, photoid string) *RepositoryResponse {
	const place = GetPhotoCache
	result, err := ph.Client.RedisClient.Get(ctx, fmt.Sprintf(KeyPhoto, userid, photoid)).Result()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorGetPhotos, err)), Place: place}
	}
	if len(result) == 0 {
		return &RepositoryResponse{Success: false, SuccessMessage: "Photo was not found in the cache", Place: place}
	}
	var photo model.Photo
	err = json.Unmarshal([]byte(result), &photo)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorUnmarshal, err)), Place: place}
	}
	return &RepositoryResponse{Success: true, Data: Data{Photo: &photo}, SuccessMessage: "Successful get photo from cache", Place: place}
}
func (ph *PhotoRedisRepo) GetPhotosCache(ctx context.Context, userid string) *RepositoryResponse {
	const place = GetPhotosCache
	result, err := ph.Client.RedisClient.Get(ctx, fmt.Sprintf(KeyPhotos, userid)).Result()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorGetPhotos, err)), Place: place}
	}
	if len(result) == 0 {
		return &RepositoryResponse{Success: false, SuccessMessage: "Photos was not found in the cache", Place: place}
	}
	var photoslice []*model.Photo
	err = json.Unmarshal([]byte(result), &photoslice)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorUnmarshal, err)), Place: place}
	}
	if len(photoslice) == 0 {
		return &RepositoryResponse{Success: true, Data: Data{Photos: photoslice}, SuccessMessage: "Get empty photos list in cache", Place: place}
	}
	return &RepositoryResponse{Success: true, Data: Data{Photos: photoslice}, SuccessMessage: "Successful get photos from cache", Place: place}
}

func (ph *PhotoRedisRepo) DeletePhotosCache(ctx context.Context, userid string) *RepositoryResponse {
	const place = DeletePhotosCache
	iter := ph.Client.RedisClient.Scan(ctx, 0, fmt.Sprintf(KeyPhoto, userid, "*"), 0).Iterator()
	pipe := ph.Client.RedisClient.TxPipeline()
	defer pipe.Discard()
	pipe.Del(ctx, fmt.Sprintf(KeyPhotos, userid))
	for iter.Next(ctx) {
		pipe.Del(ctx, iter.Val())
	}
	err := iter.Err()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorScan, err)), Place: place}
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorPipeExec, err)), Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful delete photos from cache", Place: place}
}
func (ph *PhotoRedisRepo) DeletePhotoCache(ctx context.Context, userid string, photoid string) *RepositoryResponse {
	const place = DeletePhotoCache
	err := ph.Client.RedisClient.Del(ctx, fmt.Sprintf(KeyPhoto, userid, photoid)).Err()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorDelPhotos, err)), Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful delete photo from cache", Place: place}
}
