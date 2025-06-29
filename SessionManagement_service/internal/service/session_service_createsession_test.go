package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/service"
	mock_service "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/service/mocks"
	"github.com/stretchr/testify/require"
)

func TestCreateSession_Success(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUserId := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSessionRepo := mock_service.NewMockSessionRepos(ctrl)
	mockLogProducer := mock_service.NewMockLogProducer(ctrl)
	as := &service.SessionService{
		Repo:        mockSessionRepo,
		Logproducer: mockLogProducer,
	}
	mockSessionRepo.EXPECT().SetSession(ctx, gomock.Any()).Return(&repository.RepositoryResponse{Success: true, Place: repository.SetSession, SuccessMessage: "Successfull session installation"})
	mockLogProducer.EXPECT().NewSessionLog(kafka.LogLevelInfo, gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.CreateSession(ctx, fixedUserId)
	require.True(t, response.Success)
	require.NotNil(t, response.Data)
}
func TestCreateSession_RequiredUserID(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUserId := ""
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSessionRepo := mock_service.NewMockSessionRepos(ctrl)
	mockLogProducer := mock_service.NewMockLogProducer(ctrl)
	as := &service.SessionService{
		Repo:        mockSessionRepo,
		Logproducer: mockLogProducer,
	}
	mockLogProducer.EXPECT().NewSessionLog(gomock.Any(), gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.CreateSession(ctx, fixedUserId)
	require.False(t, response.Success)
	require.NotNil(t, response.Data)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.SessionServiceUnavalaible})
}
func TestCreateSession_DataBaseError_InternalServerError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUserId := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSessionRepo := mock_service.NewMockSessionRepos(ctrl)
	mockLogProducer := mock_service.NewMockLogProducer(ctrl)
	as := &service.SessionService{
		Repo:        mockSessionRepo,
		Logproducer: mockLogProducer,
	}
	mockSessionRepo.EXPECT().SetSession(ctx, gomock.Any()).Return(&repository.RepositoryResponse{Success: false, Place: repository.SetSession, Errors: &erro.CustomError{Type: erro.ServerErrorType, Message: erro.ErrorSet}})
	mockLogProducer.EXPECT().NewSessionLog(gomock.Any(), gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.CreateSession(ctx, fixedUserId)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.SessionServiceUnavalaible})
}
