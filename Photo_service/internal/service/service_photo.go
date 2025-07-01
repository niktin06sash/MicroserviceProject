package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	pb "github.com/niktin06sash/MicroserviceProject/Photo_service/proto"
)

type PhotoServiceImplement struct {
	Photorepo   DBPhotoRepos
	Cloud       CloudPhotoStorage
	Logproducer LogProducer
	Cache       CachePhotoRepos
	Worker      *Worker
	cancel      context.CancelFunc
}

func NewPhotoServiceImplement(repo DBPhotoRepos, cloud CloudPhotoStorage, cache CachePhotoRepos, logproducer LogProducer) *PhotoServiceImplement {
	ctx, cancel := context.WithCancel(context.Background())
	photoworker := NewWorker(ctx)
	photoworker.Start()
	return &PhotoServiceImplement{Photorepo: repo, Logproducer: logproducer, Cloud: cloud, Cache: cache, Worker: photoworker, cancel: cancel}
}

func (use *PhotoServiceImplement) DeletePhoto(ctx context.Context, userid string, photoid string) *ServiceResponse {
	const place = UseCase_DeletePhoto
	traceid := ctx.Value("traceID").(string)
	bdresponse, serviceresponse := use.requestToRepository(use.Photorepo.DeletePhoto(ctx, userid, photoid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	_, serviceresponse = use.requestToRepository(use.Cache.DeletePhotoCache(ctx, userid, photoid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	ext := bdresponse.Data.ContentType
	return use.enqueueTask(ctx, func(taskctx context.Context) {
		use.deletePhotoCloud(taskctx, photoid, ext, traceid)
	}, 15*time.Second, place, traceid)
}
func (use *PhotoServiceImplement) LoadPhoto(ctx context.Context, userid string, filedata []byte) *ServiceResponse {
	const place = UseCase_LoadPhoto
	traceid := ctx.Value("traceID").(string)
	_, err := uuid.Parse(userid)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse Error: %v", err)
		use.Logproducer.NewPhotoLog(kafka.LogLevelWarn, place, traceid, fmterr)
		return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.InvalidUserIDFormat)}
	}
	if len(filedata) > MaxFileSize {
		use.Logproducer.NewPhotoLog(kafka.LogLevelWarn, place, traceid, fmt.Sprintf("File too large: %v bytes", len(filedata)))
		return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.LargeFile)}
	}
	contentType := http.DetectContentType(filedata)
	if contentType != "image/jpeg" && contentType != "image/png" {
		use.Logproducer.NewPhotoLog(kafka.LogLevelWarn, place, traceid, fmt.Sprintf("Invalid file format: %s", contentType))
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
	resp := use.enqueueTask(ctx, func(taskctx context.Context) {
		use.unloadPhotoCloud(taskctx, filedata, photoid, ext, userid, traceid)
	}, 30*time.Second, place, traceid)
	if !resp.Success {
		return resp
	}
	return &ServiceResponse{Success: true, Data: Data{PhotoID: photoid}}
}
func (use *PhotoServiceImplement) GetPhoto(ctx context.Context, photoid string, userid string) *ServiceResponse {
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
	cacheresponse, serviceresponse := use.requestToRepository(use.Cache.GetPhotoCache(ctx, userid, photoid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	if cacheresponse.Success {
		return &ServiceResponse{Success: cacheresponse.Success, Data: Data{Photo: cacheresponse.Data.GrpcPhoto}}
	}
	bdresponse, serviceresponse := use.requestToRepository(use.Photorepo.GetPhoto(ctx, userid, photoid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	photo_bd := bdresponse.Data.Photo
	grpcphoto := &pb.Photo{PhotoId: photo_bd.ID, Url: photo_bd.URL, CreatedAt: photo_bd.CreatedAt.String()}
	_, serviceresponse = use.requestToRepository(use.Cache.AddPhotoCache(ctx, userid, grpcphoto), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: true, Data: Data{Photo: grpcphoto}}
}
func (use *PhotoServiceImplement) GetPhotos(ctx context.Context, userid string) *ServiceResponse {
	traceid := ctx.Value("traceID").(string)
	const place = UseCase_GetPhotos
	err := use.parsingIDs(userid, traceid, place)
	if err != nil {
		return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.InvalidUserIDFormat)}
	}
	cacheresponse, serviceresponse := use.requestToRepository(use.Cache.GetPhotosCache(ctx, userid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	if cacheresponse.Success {
		return &ServiceResponse{Success: cacheresponse.Success, Data: Data{Photos: cacheresponse.Data.GrpcPhotos}}
	}
	bdresponse, serviceresponse := use.requestToRepository(use.Photorepo.GetPhotos(ctx, userid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	photos := bdresponse.Data.Photos
	grpcphotos := []*pb.Photo{}
	for _, p := range photos {
		grpcphoto := &pb.Photo{PhotoId: p.ID, Url: p.URL, CreatedAt: p.CreatedAt.String()}
		grpcphotos = append(grpcphotos, grpcphoto)
	}
	_, serviceresponse = use.requestToRepository(use.Cache.AddPhotosCache(ctx, userid, grpcphotos), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: true, Data: Data{Photos: grpcphotos}}
}
func (use *PhotoServiceImplement) DeleteAllUserData(ctx context.Context, userid string, traceid string) *ServiceResponse {
	const place = UseCase_DeleteAllUserData
	getctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	getresp, serviceresponse := use.requestToRepository(use.Photorepo.GetPhotos(getctx, userid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	photos := getresp.Data.Photos
	for _, photo := range photos {
		resp := use.enqueueTask(ctx, func(taskctx context.Context) {
			use.deletePhotoCloud(taskctx, photo.ID, photo.ContentType, traceid)
		}, 15*time.Second, place, traceid)
		if !resp.Success {
			return resp
		}
	}
	deluserctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	_, serviceresponse = use.requestToRepository(use.Cache.DeleteAllPhotosCache(deluserctx, userid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	_, serviceresponse = use.requestToRepository(use.Photorepo.DeleteUserData(deluserctx, userid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: true}
}
func (use *PhotoServiceImplement) AddUserId(ctx context.Context, userid string, traceid string) *ServiceResponse {
	addctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, serviceresponse := use.requestToRepository(use.Photorepo.AddUserId(addctx, userid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: true}
}
func (use *PhotoServiceImplement) Stop() {
	use.cancel()
	use.Worker.Stop()
}
