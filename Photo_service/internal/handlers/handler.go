package handlers

import (
	"context"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/service"
)

type LogProducer interface {
	NewPhotoLog(level, place, traceid, msg string)
}

const API_LoadPhoto = "API-LoadPhoto"
const API_DeletePhoto = "API-DeletePhoto"
const API_GetPhoto = "API-GetPhoto"
const API_GetPhotos = "API-GetPhotos"

type PhotoService interface {
	DeletePhoto(ctx context.Context, userid string, photoid string) *service.ServiceResponse
	LoadPhoto(ctx context.Context, userid string, filedata []byte) *service.ServiceResponse
	GetPhoto(ctx context.Context, photoid string, userid string) *service.ServiceResponse
	GetPhotos(ctx context.Context, userid string) *service.ServiceResponse
}
