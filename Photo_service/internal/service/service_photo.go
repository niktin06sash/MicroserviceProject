package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	pb "github.com/niktin06sash/MicroserviceProject/Photo_service/proto"
)

type PhotoService struct {
	repo        DBPhotoRepos
	cloud       CloudPhotoStorage
	logProducer LogProducer
	cache       CachePhotoRepos
	wg          *sync.WaitGroup
}

func NewPhotoService(repo DBPhotoRepos, cloud CloudPhotoStorage, cache CachePhotoRepos, logproducer LogProducer) *PhotoService {
	return &PhotoService{repo: repo, logProducer: logproducer, cloud: cloud, cache: cache, wg: &sync.WaitGroup{}}
}

func (use *PhotoService) DeletePhoto(ctx context.Context, userid string, photoid string) *ServiceResponse {
	traceid := ctx.Value("traceID").(string)
	bdresponse, serviceresponse := use.requestToDB(use.repo.GetPhoto(ctx, userid, photoid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	bdresponse, serviceresponse = use.requestToDB(use.repo.DeletePhoto(ctx, userid, bdresponse.Data.Photo.ID), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	_, serviceresponse = use.requestToDB(use.cache.DeletePhotoCache(ctx, userid, photoid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	ext := bdresponse.Data.ContentType
	use.wg.Add(1)
	go func() {
		defer use.wg.Done()
		use.deletePhotoCloud(photoid, ext, traceid)
	}()
	return &ServiceResponse{Success: true}
}
func (use *PhotoService) LoadPhoto(ctx context.Context, userid string, filedata []byte) *ServiceResponse {
	const place = UseCase_LoadPhoto
	traceid := ctx.Value("traceID").(string)
	_, err := uuid.Parse(userid)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse Error: %v", err)
		use.logProducer.NewPhotoLog(kafka.LogLevelWarn, place, traceid, fmterr)
		return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.InvalidUserIDFormat)}
	}
	if len(filedata) > MaxFileSize {
		use.logProducer.NewPhotoLog(kafka.LogLevelWarn, place, traceid, fmt.Sprintf("File too large: %v bytes", len(filedata)))
		return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.LargeFile)}
	}
	contentType := http.DetectContentType(filedata)
	if contentType != "image/jpeg" && contentType != "image/png" {
		use.logProducer.NewPhotoLog(kafka.LogLevelWarn, place, traceid, fmt.Sprintf("Invalid file format: %s", contentType))
		return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.InvalidFileFormat)}
	}
	photoid := uuid.New().String()
	var ext string
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	}
	use.wg.Add(1)
	go func() {
		defer use.wg.Done()
		use.unloadPhotoCloud(filedata, photoid, ext, userid, traceid)
	}()
	return &ServiceResponse{Success: true, Data: Data{PhotoID: photoid}}
}
func (use *PhotoService) GetPhoto(ctx context.Context, photoid string, userid string) *ServiceResponse {
	traceid := ctx.Value("traceID").(string)
	const place = UseCase_GetPhoto
	err := use.parsingIDs(userid, traceid, place)
	if err != nil {
		return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.InvalidUserIDFormat)}
	}
	err = use.parsingIDs(photoid, traceid, place)
	if err != nil {
		return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.InvalidPhotoIDFormat)}
	}
	cacheresponse, serviceresponse := use.requestToDB(use.cache.GetPhotoCache(ctx, userid, photoid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	if cacheresponse.Success {
		return &ServiceResponse{Success: cacheresponse.Success, Data: Data{Photo: cacheresponse.Data.GrpcPhoto}}
	}
	bdresponse, serviceresponse := use.requestToDB(use.repo.GetPhoto(ctx, userid, photoid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	photo_bd := bdresponse.Data.Photo
	grpcphoto := &pb.Photo{PhotoId: photo_bd.ID, Url: photo_bd.URL, CreatedAt: photo_bd.CreatedAt.String()}
	_, serviceresponse = use.requestToDB(use.cache.AddPhotoCache(ctx, userid, grpcphoto), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: true, Data: Data{Photo: grpcphoto}}
}
func (use *PhotoService) GetPhotos(ctx context.Context, userid string) *ServiceResponse {
	traceid := ctx.Value("traceID").(string)
	const place = UseCase_GetPhotos
	err := use.parsingIDs(userid, traceid, place)
	if err != nil {
		return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.InvalidUserIDFormat)}
	}
	cacheresponse, serviceresponse := use.requestToDB(use.cache.GetPhotosCache(ctx, userid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	if cacheresponse.Success {
		return &ServiceResponse{Success: cacheresponse.Success, Data: Data{Photos: cacheresponse.Data.GrpcPhotos}}
	}
	bdresponse, serviceresponse := use.requestToDB(use.repo.GetPhotos(ctx, userid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	photos := bdresponse.Data.Photos
	grpcphotos := []*pb.Photo{}
	for _, p := range photos {
		grpcphoto := &pb.Photo{PhotoId: p.ID, Url: p.URL, CreatedAt: p.CreatedAt.String()}
		grpcphotos = append(grpcphotos, grpcphoto)
	}
	_, serviceresponse = use.requestToDB(use.cache.AddPhotosCache(ctx, userid, grpcphotos), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: true, Data: Data{Photos: grpcphotos}}
}
