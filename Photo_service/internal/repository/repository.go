package repository

import (
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
	pb "github.com/niktin06sash/MicroserviceProject/Photo_service/proto"
)

type RepositoryResponse struct {
	Success        bool
	SuccessMessage string
	Place          string
	Data           Data
	Errors         *erro.CustomError
}
type Data struct {
	ContentType string
	Photo       *model.Photo
	Photos      []*model.Photo
	GrpcPhoto   *pb.Photo
	GrpcPhotos  []*pb.Photo
}

const LoadPhoto = "Repository-LoadPhoto"
const DeletePhoto = "Repository-DeletePhoto"
const GetPhotos = "Repository-GetPhotos"
const GetPhoto = "Repository-GetPhoto"
const DeleteUserData = "Repository-DeleteUserData"
const AddUserId = "Repository-AddUserId"
const AddPhotoCache = "Repository-AddPhotoCache"
const AddPhotosCache = "Repository-AddPhotosCache"
const GetPhotoCache = "Repository-GetPhotoCache"
const GetPhotosCache = "Repository-GetPhotosCache"
const DeletePhotosCache = "Repository-DeletePhotosCache"
const DeletePhotoCache = "Repository-DeletePhotoCache"
const DeleteAllPhotosCache = "Repository-DeleteAllPhotosCache"
