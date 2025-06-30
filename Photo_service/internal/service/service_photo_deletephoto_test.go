package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/service"
	mock_service "github.com/niktin06sash/MicroserviceProject/Photo_service/internal/service/mocks"
	"github.com/stretchr/testify/require"
)

func TestDeletePhoto_Success(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUserId := "123e4567-e89b-12d3-a456-426614174000"
	fixedPhotoID := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockphotorepo := mock_service.NewMockDBPhotoRepos(ctrl)
	mockphotocloud := mock_service.NewMockCloudPhotoStorage(ctrl)
	mocklogproducer := mock_service.NewMockLogProducer(ctrl)
	mockcache := mock_service.NewMockCachePhotoRepos(ctrl)
	as := service.PhotoServiceImplement{
		Photorepo:   mockphotorepo,
		Cache:       mockcache,
		Cloud:       mockphotocloud,
		Logproducer: mocklogproducer,
		Task_queue:  make(chan func(), 1000),
	}
	mockphotorepo.EXPECT().GetPhoto(ctx, fixedUserId, fixedPhotoID).Return(&repository.RepositoryResponse{
		Success:        true,
		SuccessMessage: "Successful get photo metadata to database",
		Place:          repository.GetPhoto,
		Data: repository.Data{
			Photo: &model.Photo{
				ID:          fixedPhotoID,
				UserID:      fixedUserId,
				URL:         "some_url_photo",
				ContentType: ".jpg",
				CreatedAt:   time.Now(),
			},
		},
	})
	mockphotorepo.EXPECT().DeletePhoto(ctx, fixedUserId, fixedPhotoID).Return(&repository.RepositoryResponse{
		Success:        true,
		SuccessMessage: "Successful delete photo metadata from database",
		Place:          repository.DeletePhoto,
	})
	mockcache.EXPECT().DeletePhotoCache(ctx, fixedUserId, fixedPhotoID).Return(&repository.RepositoryResponse{
		Success:        true,
		SuccessMessage: "Successful delete photo metadata from cache",
		Place:          repository.DeletePhotoCache,
	})
	mocklogproducer.EXPECT().NewPhotoLog(kafka.LogLevelInfo, gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.DeletePhoto(ctx, fixedUserId, fixedTraceID)
	require.True(t, response.Success)
	require.Nil(t, response.Errors)
}
func TestDeletePhoto_Success_WithEmptyCache(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUserId := "123e4567-e89b-12d3-a456-426614174000"
	fixedPhotoID := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockphotorepo := mock_service.NewMockDBPhotoRepos(ctrl)
	mockphotocloud := mock_service.NewMockCloudPhotoStorage(ctrl)
	mocklogproducer := mock_service.NewMockLogProducer(ctrl)
	mockcache := mock_service.NewMockCachePhotoRepos(ctrl)
	as := service.PhotoServiceImplement{
		Photorepo:   mockphotorepo,
		Cache:       mockcache,
		Cloud:       mockphotocloud,
		Logproducer: mocklogproducer,
		Task_queue:  make(chan func(), 1000),
	}
	mockphotorepo.EXPECT().GetPhoto(ctx, fixedUserId, fixedPhotoID).Return(&repository.RepositoryResponse{
		Success:        true,
		SuccessMessage: "Successful get photo metadata to database",
		Place:          repository.GetPhoto,
		Data: repository.Data{
			Photo: &model.Photo{
				ID:          fixedPhotoID,
				UserID:      fixedUserId,
				URL:         "some_url_photo",
				ContentType: ".jpg",
				CreatedAt:   time.Now(),
			},
		},
	})
	mockphotorepo.EXPECT().DeletePhoto(ctx, fixedUserId, fixedPhotoID).Return(&repository.RepositoryResponse{
		Success:        true,
		SuccessMessage: "Successful delete photo metadata from database",
		Place:          repository.DeletePhoto,
	})
	mockcache.EXPECT().DeletePhotoCache(ctx, fixedUserId, fixedPhotoID).Return(&repository.RepositoryResponse{
		Success:        false,
		SuccessMessage: "Photos metadata was not found in the cache",
		Place:          repository.DeletePhotoCache,
	})
	mocklogproducer.EXPECT().NewPhotoLog(kafka.LogLevelInfo, gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.DeletePhoto(ctx, fixedUserId, fixedTraceID)
	require.True(t, response.Success)
	require.Nil(t, response.Errors)
}
func TestDeletePhoto_DatabaseClientError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUserId := "123e4567-e89b-12d3-a456-426614174000"
	fixedPhotoID := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockphotorepo := mock_service.NewMockDBPhotoRepos(ctrl)
	mockphotocloud := mock_service.NewMockCloudPhotoStorage(ctrl)
	mocklogproducer := mock_service.NewMockLogProducer(ctrl)
	mockcache := mock_service.NewMockCachePhotoRepos(ctrl)
	as := service.PhotoServiceImplement{
		Photorepo:   mockphotorepo,
		Cache:       mockcache,
		Cloud:       mockphotocloud,
		Logproducer: mocklogproducer,
		Task_queue:  make(chan func(), 1000),
	}
	mockphotorepo.EXPECT().GetPhoto(ctx, fixedUserId, fixedPhotoID).Return(&repository.RepositoryResponse{
		Success: false,
		Place:   repository.GetPhoto,
		Errors:  erro.ClientError(erro.NonExistentData),
	},
	)
	mocklogproducer.EXPECT().NewPhotoLog(kafka.LogLevelWarn, gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.DeletePhoto(ctx, fixedUserId, fixedTraceID)
	require.False(t, response.Success)
	require.Equal(t, erro.ClientError(erro.NonExistentData), response.Errors)
}
func TestDeletePhoto_DatabaseInternalServerError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUserId := "123e4567-e89b-12d3-a456-426614174000"
	fixedPhotoID := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockphotorepo := mock_service.NewMockDBPhotoRepos(ctrl)
	mockphotocloud := mock_service.NewMockCloudPhotoStorage(ctrl)
	mocklogproducer := mock_service.NewMockLogProducer(ctrl)
	mockcache := mock_service.NewMockCachePhotoRepos(ctrl)
	as := service.PhotoServiceImplement{
		Photorepo:   mockphotorepo,
		Cache:       mockcache,
		Cloud:       mockphotocloud,
		Logproducer: mocklogproducer,
		Task_queue:  make(chan func(), 1000),
	}
	mockphotorepo.EXPECT().GetPhoto(ctx, fixedUserId, fixedPhotoID).Return(&repository.RepositoryResponse{
		Success:        true,
		SuccessMessage: "Successful get photo metadata to database",
		Place:          repository.GetPhoto,
		Data: repository.Data{
			Photo: &model.Photo{
				ID:          fixedPhotoID,
				UserID:      fixedUserId,
				URL:         "some_url_photo",
				ContentType: ".jpg",
				CreatedAt:   time.Now(),
			},
		},
	})
	mockphotorepo.EXPECT().DeletePhoto(ctx, fixedUserId, fixedPhotoID).Return(&repository.RepositoryResponse{
		Success: false,
		Place:   repository.DeletePhoto,
		Errors:  erro.ServerError(erro.ErrorAfterReqPhotos),
	})
	mocklogproducer.EXPECT().NewPhotoLog(gomock.Any(), gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.DeletePhoto(ctx, fixedUserId, fixedTraceID)
	require.False(t, response.Success)
	require.Equal(t, erro.ServerError(erro.PhotoServiceUnavalaible), response.Errors)
}
func TestDeletePhoto_DatabaseInternalServerError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUserId := "123e4567-e89b-12d3-a456-426614174000"
	fixedPhotoID := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockphotorepo := mock_service.NewMockDBPhotoRepos(ctrl)
	mockphotocloud := mock_service.NewMockCloudPhotoStorage(ctrl)
	mocklogproducer := mock_service.NewMockLogProducer(ctrl)
	mockcache := mock_service.NewMockCachePhotoRepos(ctrl)
	as := service.PhotoServiceImplement{
		Photorepo:   mockphotorepo,
		Cache:       mockcache,
		Cloud:       mockphotocloud,
		Logproducer: mocklogproducer,
		Task_queue:  make(chan func(), 1000),
	}
	mockphotorepo.EXPECT().GetPhoto(ctx, fixedUserId, fixedPhotoID).Return(&repository.RepositoryResponse{
		Success:        true,
		SuccessMessage: "Successful get photo metadata to database",
		Place:          repository.GetPhoto,
		Data: repository.Data{
			Photo: &model.Photo{
				ID:          fixedPhotoID,
				UserID:      fixedUserId,
				URL:         "some_url_photo",
				ContentType: ".jpg",
				CreatedAt:   time.Now(),
			},
		},
	})
	mockphotorepo.EXPECT().DeletePhoto(ctx, fixedUserId, fixedPhotoID).Return(&repository.RepositoryResponse{
		Success: false,
		Place:   repository.DeletePhoto,
		Errors:  erro.ServerError(erro.ErrorAfterReqPhotos),
	})
	mocklogproducer.EXPECT().NewPhotoLog(gomock.Any(), gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.DeletePhoto(ctx, fixedUserId, fixedTraceID)
	require.False(t, response.Success)
	require.Equal(t, erro.ServerError(erro.PhotoServiceUnavalaible), response.Errors)
}
