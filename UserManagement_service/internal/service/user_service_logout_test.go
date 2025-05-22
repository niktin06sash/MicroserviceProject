package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	mock_client "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/client/mocks"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	mock_kafka "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka/mocks"
	mock_repository "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository/mocks"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLogout_Success(t *testing.T) {
	var place = "UseCase-Logout"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
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

	mockGrpc.EXPECT().DeleteSession(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}), mock.MatchedBy(func(sessionID string) bool {
		return sessionID == fixedSessId
	})).Return(&pb.DeleteSessionResponse{
		Success: true,
	}, nil)
	mockKafka.EXPECT().NewUserLog(kafka.LogLevelInfo, place, fixedTraceID, "The session was deleted successfully")
	response := as.Logout(ctx, fixedSessId)
	require.True(t, response.Success)
}
func TestLogout_RetryGrpc_ContextCanceled(t *testing.T) {
	var place = "UseCase-Logout"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cancel()
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
	mockKafka.EXPECT().NewUserLog(kafka.LogLevelError, place, fixedTraceID, gomock.Any())
	response := as.Logout(ctx, fixedSessId)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], "Request timed out")
	require.Equal(t, erro.ServerErrorType, response.Type)
}
func TestLogout_RetryGrpc_ClientError(t *testing.T) {
	var place = "UseCase-Logout"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
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
	)
	response := as.Logout(ctx, fixedSessId)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "ClientError")
	require.EqualError(t, response.Errors["ClientError"], "Session not found")
	require.Equal(t, erro.ClientErrorType, response.Type)
}
func TestLogout_RetryGrpc_InternalServerError(t *testing.T) {
	var place = "UseCase-Logout"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
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
	)
	response := as.Logout(ctx, fixedSessId)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], erro.SessionServiceUnavalaible)
	require.Equal(t, erro.ServerErrorType, response.Type)
}
