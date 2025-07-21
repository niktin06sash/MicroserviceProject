package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository"
	pb "github.com/niktin06sash/MicroserviceProject/Photo_service/proto"
	"github.com/redis/go-redis/v9"
)

type PhotoCache struct {
	cacheclient *CacheObject
}

func NewPhotoCache(red *CacheObject) *PhotoCache {
	return &PhotoCache{cacheclient: red}
}

const KeyPhoto = "photo:%s:%s"
const KeyPhotos = "photos:%s"

func (ph *PhotoCache) AddPhotoCache(ctx context.Context, userid string, photo *pb.Photo) *repository.RepositoryResponse {
	const place = AddPhotoCache
	jsondata, err := json.Marshal(photo)
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorMarshal, err)), place)
	}
	err = ph.cacheclient.connect.Set(ctx, fmt.Sprintf(KeyPhoto, userid, photo.PhotoId), jsondata, 1*time.Hour).Err()
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorSetPhotos, err)), place)
	}
	return &repository.RepositoryResponse{Success: true, SuccessMessage: "Successful add photo metadata in cache", Place: place}
}
func (ph *PhotoCache) AddPhotosCache(ctx context.Context, userid string, photosslice []*pb.Photo) *repository.RepositoryResponse {
	const place = AddPhotosCache
	jsondata_photos, err := json.Marshal(photosslice)
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorMarshal, err)), place)
	}
	pipe := ph.cacheclient.connect.TxPipeline()
	defer pipe.Discard()
	pipe.Set(ctx, fmt.Sprintf(KeyPhotos, userid), jsondata_photos, 1*time.Hour)
	for _, photo := range photosslice {
		jsondata_photo, err := json.Marshal(photo)
		if err != nil {
			return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorMarshal, err)), place)
		}
		pipe.Set(ctx, fmt.Sprintf(KeyPhoto, userid, photo.PhotoId), jsondata_photo, 1*time.Hour)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorPipeExec, err)), place)
	}
	return &repository.RepositoryResponse{Success: true, SuccessMessage: "Successful add photos metadata in cache", Place: place}
}
func (ph *PhotoCache) GetPhotoCache(ctx context.Context, userid string, photoid string) *repository.RepositoryResponse {
	const place = GetPhotoCache
	result, err := ph.cacheclient.connect.Get(ctx, fmt.Sprintf(KeyPhoto, userid, photoid)).Result()
	if err != nil {
		if err == redis.Nil {
			return &repository.RepositoryResponse{Success: false, SuccessMessage: "Photo metadata was not found in the cache", Place: place}
		}
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorGetPhotos, err)), place)
	}
	var photo pb.Photo
	err = json.Unmarshal([]byte(result), &photo)
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorUnmarshal, err)), place)
	}
	return &repository.RepositoryResponse{Success: true, Data: repository.Data{GrpcPhoto: &photo}, SuccessMessage: "Successful get photo metadata from cache", Place: place}
}
func (ph *PhotoCache) GetPhotosCache(ctx context.Context, userid string) *repository.RepositoryResponse {
	const place = GetPhotosCache
	result, err := ph.cacheclient.connect.Get(ctx, fmt.Sprintf(KeyPhotos, userid)).Result()
	if err != nil {
		if err == redis.Nil {
			return &repository.RepositoryResponse{Success: false, SuccessMessage: "Photo metadata was not found in the cache", Place: place}
		}
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorGetPhotos, err)), place)
	}
	var photoslice []*pb.Photo
	err = json.Unmarshal([]byte(result), &photoslice)
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorUnmarshal, err)), place)
	}
	if len(photoslice) == 0 {
		return &repository.RepositoryResponse{Success: true, Data: repository.Data{GrpcPhotos: photoslice}, SuccessMessage: "Successful get empty photos metadata in cache", Place: place}
	}
	return &repository.RepositoryResponse{Success: true, Data: repository.Data{GrpcPhotos: photoslice}, SuccessMessage: "Successful get photos metadata from cache", Place: place}
}

func (ph *PhotoCache) DeleteAllPhotosCache(ctx context.Context, userid string) *repository.RepositoryResponse {
	const place = DeleteAllPhotosCache
	var count int64
	num, err := ph.cacheclient.connect.Del(ctx, fmt.Sprintf(KeyPhotos, userid)).Result()
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorDelPhotos, err)), place)
	}
	count += num
	iter := ph.cacheclient.connect.Scan(ctx, 0, fmt.Sprintf(KeyPhoto, userid, "*"), 0).Iterator()
	for iter.Next(ctx) {
		num, err := ph.cacheclient.connect.Del(ctx, iter.Val()).Result()
		if err != nil {
			return &repository.RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorDelPhotos, err)), Place: place}
		}
		count += num
	}
	if err := iter.Err(); err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorScan, err)), place)
	}
	if count == 0 {
		return &repository.RepositoryResponse{Success: false, SuccessMessage: "Photos metadata was not found in the cache", Place: place}
	}
	return &repository.RepositoryResponse{Success: true, SuccessMessage: "Successful delete all photos metadata from cache", Place: place}
}
func (ph *PhotoCache) DeletePhotosCache(ctx context.Context, userid string) *repository.RepositoryResponse {
	const place = DeletePhotosCache
	num, err := ph.cacheclient.connect.Del(ctx, fmt.Sprintf(KeyPhotos, userid)).Result()
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorDelPhotos, err)), place)
	}
	if num == 0 {
		return &repository.RepositoryResponse{Success: false, SuccessMessage: "Photos metadata was not found in the cache", Place: place}
	}
	return &repository.RepositoryResponse{Success: true, SuccessMessage: "Successful delete photos metadata from cache", Place: place}
}
func (ph *PhotoCache) DeletePhotoCache(ctx context.Context, userid string, photoid string) *repository.RepositoryResponse {
	const place = DeletePhotoCache
	_, err := ph.cacheclient.connect.Del(ctx, fmt.Sprintf(KeyPhoto, userid, photoid)).Result()
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorDelPhotos, err)), place)
	}
	_, err = ph.cacheclient.connect.Del(ctx, fmt.Sprintf(KeyPhotos, userid)).Result()
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorDelPhotos, err)), place)
	}
	return &repository.RepositoryResponse{Success: true, SuccessMessage: "Successful delete photo metadata from cache", Place: place}
}
