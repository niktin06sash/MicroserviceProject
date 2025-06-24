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

func TestDeleteSession_Success(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-426614174000"
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
	mockSessionRepo.EXPECT().DeleteSession(ctx, fixedSessionId).Return(&repository.RepositoryResponse{Success: true, Place: repository.DeleteSession, SuccessMessage: "Successfull delete session"})
	mockLogProducer.EXPECT().NewSessionLog(kafka.LogLevelInfo, gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.DeleteSession(ctx, fixedSessionId)
	require.True(t, response.Success)
}
func TestDeleteSession_ParsingSessionID_Error(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-42661417000"
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
	response := as.DeleteSession(ctx, fixedSessionId)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ClientErrorType, Message: erro.SessionIdInvalid})
}
func TestDeleteSession_DataBaseError_InternalServerError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-426614174000"
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
	mockSessionRepo.EXPECT().DeleteSession(ctx, fixedSessionId).Return(&repository.RepositoryResponse{Success: false, Place: repository.DeleteSession, Errors: &erro.CustomError{Type: erro.ServerErrorType, Message: erro.ErrorDelSession}})
	mockLogProducer.EXPECT().NewSessionLog(gomock.Any(), gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.DeleteSession(ctx, fixedSessionId)
	require.False(t, response.Success)
	require.Equal(t, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.SessionServiceUnavalaible}, response.Errors)
}
func TestDeleteSession_DataBaseError_ClientError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-426614174000"
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
	mockSessionRepo.EXPECT().DeleteSession(ctx, fixedSessionId).Return(&repository.RepositoryResponse{Success: false, Place: repository.DeleteSession, Errors: &erro.CustomError{Type: erro.ClientErrorType, Message: erro.InvalidSession}})
	mockLogProducer.EXPECT().NewSessionLog(gomock.Any(), gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.DeleteSession(ctx, fixedSessionId)
	require.False(t, response.Success)
	require.Equal(t, &erro.CustomError{Type: erro.ClientErrorType, Message: erro.InvalidSession}, response.Errors)
}
