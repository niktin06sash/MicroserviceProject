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

func TestValidateSessionTrue_Success(t *testing.T) {
	flag := "true"
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
	mockSessionRepo.EXPECT().GetSession(ctx, fixedSessionId, flag).Return(&repository.RepositoryResponse{Success: true, SuccessMessage: "Successfull get session",
		Data: repository.Data{UserID: gomock.Any().String()}, Place: repository.GetSession})
	mockLogProducer.EXPECT().NewSessionLog(kafka.LogLevelInfo, gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.ValidateSession(ctx, fixedSessionId, flag)
	require.True(t, response.Success)
	require.NotNil(t, response.Data)
}
func TestValidateSessionFalse_Success(t *testing.T) {
	flag := "false"
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
	mockSessionRepo.EXPECT().GetSession(ctx, fixedSessionId, flag).Return(&repository.RepositoryResponse{Success: true, SuccessMessage: "Request for an unauthorized page with invalid session", Place: repository.GetSession})
	mockLogProducer.EXPECT().NewSessionLog(kafka.LogLevelInfo, gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.ValidateSession(ctx, fixedSessionId, flag)
	require.True(t, response.Success)
}
func TestValidateSessionTrue_InvalidSession(t *testing.T) {
	flag := "true"
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
	mockSessionRepo.EXPECT().GetSession(ctx, fixedSessionId, flag).Return(&repository.RepositoryResponse{Success: false, Errors: &erro.CustomError{Type: erro.ClientErrorType, Message: erro.InvalidSession}, Place: repository.GetSession})
	mockLogProducer.EXPECT().NewSessionLog(gomock.Any(), gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.ValidateSession(ctx, fixedSessionId, flag)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ClientErrorType, Message: erro.InvalidSession})
}
func TestValidateSessionFalse_AlreadyAuthorized(t *testing.T) {
	flag := "false"
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
	mockSessionRepo.EXPECT().GetSession(ctx, fixedSessionId, flag).Return(&repository.RepositoryResponse{Success: false, Errors: &erro.CustomError{Type: erro.ClientErrorType, Message: erro.AlreadyAuthorized}, Place: repository.GetSession})
	mockLogProducer.EXPECT().NewSessionLog(gomock.Any(), gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.ValidateSession(ctx, fixedSessionId, flag)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ClientErrorType, Message: erro.AlreadyAuthorized})
}
func TestValidateSession_DataBaseError_InternalServerError(t *testing.T) {
	flag := "true"
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
	mockSessionRepo.EXPECT().GetSession(ctx, fixedSessionId, flag).Return(&repository.RepositoryResponse{Success: false, Place: repository.GetSession, Errors: &erro.CustomError{Type: erro.ServerErrorType, Message: erro.ErrorHgetAll}})
	mockLogProducer.EXPECT().NewSessionLog(gomock.Any(), gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.ValidateSession(ctx, fixedSessionId, flag)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.SessionServiceUnavalaible})
}
func TestValidateSession_ParsingSessionID_Error(t *testing.T) {
	flag := "false"
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
	response := as.ValidateSession(ctx, fixedSessionId, flag)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ClientErrorType, Message: erro.SessionIdInvalid})
}
