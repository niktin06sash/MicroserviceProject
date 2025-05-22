package service_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	mock_client "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/client/mocks"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	mock_kafka "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka/mocks"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	mock_repository "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository/mocks"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDeleteAccount_Success(t *testing.T) {
	var place = "UseCase-DeleteAccount"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	password := "password123"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	mockGrpc := mock_client.NewMockGrpcClientService(ctrl)
	mockKafka := mock_kafka.NewMockKafkaProducerService(ctrl)
	as := &service.AuthService{
		Dbrepo:        mockRepo,
		Dbtxmanager:   mockTxManager,
		GrpcClient:    mockGrpc,
		KafkaProducer: mockKafka,
		Validator:     validator.New(),
	}
	tx := &sql.Tx{}
	mockTxManager.EXPECT().BeginTx(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	})).Return(tx, nil)
	mockRepo.EXPECT().DeleteUser(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx,
		mock.MatchedBy(func(userid uuid.UUID) bool {
			return userid == fixedUUID
		}),
		mock.MatchedBy(func(password string) bool {
			return password == "password123"
		}),
	).Return(&repository.DBRepositoryResponse{
		Success: true,
	})
	mockGrpc.EXPECT().DeleteSession(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}), mock.MatchedBy(func(sessionID string) bool {
		return sessionID == fixedSessId
	})).Return(&pb.DeleteSessionResponse{
		Success: true,
	}, nil)
	mockTxManager.EXPECT().CommitTx(tx).Return(nil)
	mockKafka.EXPECT().NewUserLog(kafka.LogLevelInfo, place, fixedTraceID, "Successful commit on attempt 1")
	mockKafka.EXPECT().NewUserLog(kafka.LogLevelInfo, place, fixedTraceID, "Transaction was successfully committed and user has successfully deleted his account with all data")
	response := as.DeleteAccount(ctx, fixedSessId, fixedUUID.String(), password)
	require.True(t, response.Success)
}
func TestDeleteAccount_InvalidUserID(t *testing.T) {
	var place = "UseCase-DeleteAccount"
	fixedTraceUuid := "123e4567-e89b-12d3-a456-426614174000"
	invalidUserID := "invalid-uuid"
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	password := "password123"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockKafka := mock_kafka.NewMockKafkaProducerService(ctrl)
	as := &service.AuthService{
		KafkaProducer: mockKafka,
		Validator:     validator.New(),
	}
	mockKafka.EXPECT().NewUserLog(kafka.LogLevelError, place, fixedTraceUuid, gomock.Any())
	response := as.DeleteAccount(ctx, fixedSessId, invalidUserID, password)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], erro.UserServiceUnavalaible)
	require.Equal(t, erro.ServerErrorType, response.Type)
}
func TestDeleteAccount_BeginTxError(t *testing.T) {
	var place = "UseCase-DeleteAccount"
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000").String()
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	fixedTraceUuid := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	password := "password123"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)
	mockKafka := mock_kafka.NewMockKafkaProducerService(ctrl)
	as := &service.AuthService{
		Dbtxmanager:   mockTxManager,
		Validator:     validator.New(),
		KafkaProducer: mockKafka,
	}
	mockTxManager.EXPECT().BeginTx(gomock.Any()).Return(nil, fmt.Errorf("database connection error"))
	mockKafka.EXPECT().NewUserLog(kafka.LogLevelError, place, fixedTraceUuid, gomock.Any())
	response := as.DeleteAccount(ctx, fixedSessId, fixedUUID, password)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], erro.UserServiceUnavalaible)
	require.Equal(t, erro.ServerErrorType, response.Type)
}
func TestDeleteAccount_DataBaseError_InternalServerError(t *testing.T) {
	var place = "UseCase-DeleteAccount"
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	fixedTraceUuid := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	password := "password123"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)
	mockKafka := mock_kafka.NewMockKafkaProducerService(ctrl)
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	as := &service.AuthService{
		Dbtxmanager:   mockTxManager,
		Validator:     validator.New(),
		KafkaProducer: mockKafka,
		Dbrepo:        mockRepo,
	}
	tx := &sql.Tx{}
	mockTxManager.EXPECT().BeginTx(ctx).Return(tx, nil)
	mockRepo.EXPECT().DeleteUser(ctx, tx, fixedUUID, password).Return(&repository.DBRepositoryResponse{Success: false, Errors: fmt.Errorf(erro.UserServiceUnavalaible), Type: erro.ServerErrorType})
	mockTxManager.EXPECT().RollbackTx(tx).Return(nil)
	mockKafka.EXPECT().NewUserLog(kafka.LogLevelInfo, place, fixedTraceUuid, "Successful rollback on attempt 1")
	response := as.DeleteAccount(ctx, fixedSessId, fixedUUID.String(), password)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], erro.UserServiceUnavalaible)
	require.Equal(t, erro.ServerErrorType, response.Type)
}
func TestDeleteAccount_DataBaseError_ClientError(t *testing.T) {
	var place = "UseCase-DeleteAccount"
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	fixedTraceUuid := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	password := "password123"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)
	mockKafka := mock_kafka.NewMockKafkaProducerService(ctrl)
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	as := &service.AuthService{
		Dbtxmanager:   mockTxManager,
		Validator:     validator.New(),
		KafkaProducer: mockKafka,
		Dbrepo:        mockRepo,
	}
	tx := &sql.Tx{}
	mockTxManager.EXPECT().BeginTx(ctx).Return(tx, nil)
	mockRepo.EXPECT().DeleteUser(ctx, tx, fixedUUID, password).Return(&repository.DBRepositoryResponse{Success: false, Errors: erro.ErrorInvalidPassword, Type: erro.ClientErrorType})
	mockTxManager.EXPECT().RollbackTx(tx).Return(nil)
	mockKafka.EXPECT().NewUserLog(kafka.LogLevelInfo, place, fixedTraceUuid, "Successful rollback on attempt 1")
	response := as.DeleteAccount(ctx, fixedSessId, fixedUUID.String(), password)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "ClientError")
	require.EqualError(t, response.Errors["ClientError"], erro.ErrorInvalidPassword.Error())
	require.Equal(t, erro.ClientErrorType, response.Type)
}
func TestDeleteAccount_RetryGrpc_InternalServerError(t *testing.T) {
	var place = "UseCase-DeleteAccount"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	password := "password123"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	mockGrpcClient := mock_client.NewMockGrpcClientService(ctrl)
	mockKafka := mock_kafka.NewMockKafkaProducerService(ctrl)
	as := &service.AuthService{
		Dbtxmanager:   mockTxManager,
		Dbrepo:        mockRepo,
		GrpcClient:    mockGrpcClient,
		KafkaProducer: mockKafka,
		Validator:     validator.New(),
	}
	tx := &sql.Tx{}
	mockTxManager.EXPECT().BeginTx(ctx).Return(tx, nil)
	mockRepo.EXPECT().DeleteUser(ctx, tx, fixedUUID, password).Return(&repository.DBRepositoryResponse{Success: true})
	mockGrpcClient.EXPECT().
		DeleteSession(mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
			fixedSessId).
		Return(&pb.DeleteSessionResponse{
			Success: false}, status.Error(codes.Internal, erro.SessionServiceUnavalaible)).
		Times(3)
	gomock.InOrder(
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, "Operation attempt 1 failed"),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, "Session-Service is unavailable, retrying..."),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, "Operation attempt 2 failed"),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, "Session-Service is unavailable, retrying..."),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, "Operation attempt 3 failed"),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, "Session-Service is unavailable, retrying..."),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelError, place, fixedTraceID, "All retry attempts failed"),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelInfo, place, fixedTraceID, "Successful rollback on attempt 1"),
	)
	mockTxManager.EXPECT().RollbackTx(tx).Return(nil)
	response := as.DeleteAccount(ctx, fixedSessId, fixedUUID.String(), password)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], erro.SessionServiceUnavalaible)
	require.Equal(t, erro.ServerErrorType, response.Type)
}
func TestDeleteAccount_RetryGrpc_ContextCanceled(t *testing.T) {
	var place = "UseCase-DeleteAccount"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cancel()
	password := "password123"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	mockGrpcClient := mock_client.NewMockGrpcClientService(ctrl)
	mockKafka := mock_kafka.NewMockKafkaProducerService(ctrl)
	as := &service.AuthService{
		Dbtxmanager:   mockTxManager,
		Dbrepo:        mockRepo,
		GrpcClient:    mockGrpcClient,
		KafkaProducer: mockKafka,
		Validator:     validator.New(),
	}

	tx := &sql.Tx{}
	mockTxManager.EXPECT().BeginTx(ctx).Return(tx, nil)
	mockRepo.EXPECT().DeleteUser(ctx, tx, fixedUUID, password).Return(&repository.DBRepositoryResponse{Success: true})
	gomock.InOrder(
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelError, place, fixedTraceID, gomock.Any()),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelInfo, place, fixedTraceID, "Successful rollback on attempt 1"),
	)
	mockTxManager.EXPECT().RollbackTx(tx).Return(nil)
	response := as.DeleteAccount(ctx, fixedSessId, fixedUUID.String(), password)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], "Request timed out")
	require.Equal(t, erro.ServerErrorType, response.Type)
}
func TestDeleteAccount_RetryGrpc_ClientError(t *testing.T) {
	var place = "UseCase-DeleteAccount"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	password := "password123"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	mockGrpcClient := mock_client.NewMockGrpcClientService(ctrl)
	mockKafka := mock_kafka.NewMockKafkaProducerService(ctrl)
	as := &service.AuthService{
		Dbtxmanager:   mockTxManager,
		Dbrepo:        mockRepo,
		GrpcClient:    mockGrpcClient,
		KafkaProducer: mockKafka,
		Validator:     validator.New(),
	}
	tx := &sql.Tx{}
	mockTxManager.EXPECT().BeginTx(ctx).Return(tx, nil)
	mockRepo.EXPECT().DeleteUser(ctx, tx, fixedUUID, password).Return(&repository.DBRepositoryResponse{Success: true})
	mockGrpcClient.EXPECT().
		DeleteSession(mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
			fixedSessId).
		Return(&pb.DeleteSessionResponse{
			Success: false}, status.Error(codes.InvalidArgument, "Session not found")).
		Times(1)
	gomock.InOrder(
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, "Operation attempt 1 failed"),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, gomock.Any()),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelInfo, place, fixedTraceID, "Successful rollback on attempt 1"),
	)
	mockTxManager.EXPECT().RollbackTx(tx).Return(nil)
	response := as.DeleteAccount(ctx, fixedSessId, fixedUUID.String(), password)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "ClientError")
	require.EqualError(t, response.Errors["ClientError"], "Session not found")
	require.Equal(t, erro.ClientErrorType, response.Type)
}
func TestDeleteAccount_CommitError(t *testing.T) {
	var place = "UseCase-DeleteAccount"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	password := "password123"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	mockGrpcClient := mock_client.NewMockGrpcClientService(ctrl)
	mockKafka := mock_kafka.NewMockKafkaProducerService(ctrl)
	as := &service.AuthService{
		Dbtxmanager:   mockTxManager,
		Dbrepo:        mockRepo,
		GrpcClient:    mockGrpcClient,
		KafkaProducer: mockKafka,
		Validator:     validator.New(),
	}
	tx := &sql.Tx{}
	mockTxManager.EXPECT().BeginTx(ctx).Return(tx, nil)
	mockRepo.EXPECT().DeleteUser(ctx, tx, fixedUUID, password).Return(&repository.DBRepositoryResponse{Success: true})

	mockGrpcClient.EXPECT().DeleteSession(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}), mock.MatchedBy(func(sessionID string) bool {
		return sessionID == fixedSessId
	})).Return(&pb.DeleteSessionResponse{
		Success: true,
	}, nil)
	mockTxManager.EXPECT().CommitTx(tx).Return(fmt.Errorf("Failed to commit transaction after all attempts")).Times(3)
	mockTxManager.EXPECT().RollbackTx(tx).Return(nil)
	gomock.InOrder(
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, gomock.Any()),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, gomock.Any()),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, gomock.Any()),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelError, place, fixedTraceID, "Failed to commit transaction after all attempts"),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelInfo, place, fixedTraceID, "Successful rollback on attempt 1"),
	)
	response := as.DeleteAccount(ctx, fixedSessId, fixedUUID.String(), password)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], erro.UserServiceUnavalaible)
	require.Equal(t, erro.ServerErrorType, response.Type)
}
