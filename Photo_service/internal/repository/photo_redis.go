package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	pb "github.com/niktin06sash/MicroserviceProject/Photo_service/proto"
)

type PhotoRedisRepo struct {
	Client *RedisObject
}

func NewPhotoRedisRepo(red *RedisObject) *PhotoRedisRepo {
	return &PhotoRedisRepo{Client: red}
}

const KeyPhoto = "photo:%s:%s"
const KeyPhotos = "photos:%s"

func (ph *PhotoRedisRepo) AddPhotoCache(ctx context.Context, userid string, photo *pb.Photo) *RepositoryResponse {
	const place = AddPhotoCache
	jsondata, err := json.Marshal(photo)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorMarshal, err)), Place: place}
	}
	err = ph.Client.RedisClient.Set(ctx, fmt.Sprintf(KeyPhoto, userid, photo.PhotoId), jsondata, 1*time.Hour).Err()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorSetPhotos, err)), Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful add photo metadata in cache", Place: place}
}
func (ph *PhotoRedisRepo) AddPhotosCache(ctx context.Context, userid string, photosslice []*pb.Photo) *RepositoryResponse {
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
		pipe.Set(ctx, fmt.Sprintf(KeyPhoto, userid, photo.PhotoId), jsondata_photo, 1*time.Hour)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorPipeExec, err)), Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful add photos metadata in cache", Place: place}
}
func (ph *PhotoRedisRepo) GetPhotoCache(ctx context.Context, userid string, photoid string) *RepositoryResponse {
	const place = GetPhotoCache
	result, err := ph.Client.RedisClient.Get(ctx, fmt.Sprintf(KeyPhoto, userid, photoid)).Result()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorGetPhotos, err)), Place: place}
	}
	if len(result) == 0 {
		return &RepositoryResponse{Success: false, SuccessMessage: "Photo metadata was not found in the cache", Place: place}
	}
	var photo pb.Photo
	err = json.Unmarshal([]byte(result), &photo)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorUnmarshal, err)), Place: place}
	}
	return &RepositoryResponse{Success: true, Data: Data{GrpcPhoto: &photo}, SuccessMessage: "Successful get photo metadata from cache", Place: place}
}
func (ph *PhotoRedisRepo) GetPhotosCache(ctx context.Context, userid string) *RepositoryResponse {
	const place = GetPhotosCache
	result, err := ph.Client.RedisClient.Get(ctx, fmt.Sprintf(KeyPhotos, userid)).Result()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorGetPhotos, err)), Place: place}
	}
	if len(result) == 0 {
		return &RepositoryResponse{Success: false, SuccessMessage: "Photos metadata was not found in the cache", Place: place}
	}
	var photoslice []*pb.Photo
	err = json.Unmarshal([]byte(result), &photoslice)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorUnmarshal, err)), Place: place}
	}
	if len(photoslice) == 0 {
		return &RepositoryResponse{Success: true, Data: Data{GrpcPhotos: photoslice}, SuccessMessage: "Successful get empty photos metadata in cache", Place: place}
	}
	return &RepositoryResponse{Success: true, Data: Data{GrpcPhotos: photoslice}, SuccessMessage: "Successful get photos metadata from cache", Place: place}
}

func (ph *PhotoRedisRepo) DeletePhotosCache(ctx context.Context, userid string) *RepositoryResponse {
	const place = DeletePhotosCache
	iter := ph.Client.RedisClient.Scan(ctx, 0, fmt.Sprintf(KeyPhoto, userid, "*"), 0).Iterator()
	var count int64
	num, err := ph.Client.RedisClient.Del(ctx, fmt.Sprintf(KeyPhotos, userid)).Result()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorDelPhotos, err)), Place: place}
	}
	count += num
	for iter.Next(ctx) {
		num, err := ph.Client.RedisClient.Del(ctx, iter.Val()).Result()
		if err != nil {
			return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorDelPhotos, err)), Place: place}
		}
		count += num
	}
	if err := iter.Err(); err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorScan, err)), Place: place}
	}
	if count == 0 {
		return &RepositoryResponse{Success: false, SuccessMessage: "Photos metadata was not found in the cache", Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful delete photos metadata from cache", Place: place}
}
func (ph *PhotoRedisRepo) DeletePhotoCache(ctx context.Context, userid string, photoid string) *RepositoryResponse {
	const place = DeletePhotoCache
	num, err := ph.Client.RedisClient.Del(ctx, fmt.Sprintf(KeyPhoto, userid, photoid)).Result()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorDelPhotos, err)), Place: place}
	}
	if num == 0 {
		return &RepositoryResponse{Success: false, SuccessMessage: "Photo metadata was not found in the cache", Place: place}
	}
	num, err = ph.Client.RedisClient.Del(ctx, fmt.Sprintf(KeyPhotos, userid)).Result()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorDelPhotos, err)), Place: place}
	}
	if num == 0 {
		return &RepositoryResponse{Success: false, SuccessMessage: "Photos metadata was not found in the cache", Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful delete photo from cache", Place: place}
}
