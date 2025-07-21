package repository

import (
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
	pb "github.com/niktin06sash/MicroserviceProject/Photo_service/proto"
)

type RepositoryResponse struct {
	Success        bool
	SuccessMessage string
	Place          string
	Data           Data
	Errors         error
}
type Data struct {
	ContentType string
	Photo       *model.Photo
	Photos      []*model.Photo
	GrpcPhoto   *pb.Photo
	GrpcPhotos  []*pb.Photo
}

func BadResponse(err error, place string) *RepositoryResponse {
	return &RepositoryResponse{
		Success: false,
		Errors:  err,
		Place:   place,
	}
}
