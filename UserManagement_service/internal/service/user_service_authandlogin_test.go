package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	mock_client "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/client/mocks"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	mock_kafka "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka/mocks"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	mock_repository "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository/mocks"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuthenticateAndLogin_Success(t *testing.T) {
	var place = "UseCase-AuthenticateAndLogin"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-426614171000"
	fixedUserId := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	mockGrpc := mock_client.NewMockGrpcClientService(ctrl)
	mockKafka := mock_kafka.NewMockKafkaProducerService(ctrl)
	as := &service.AuthService{
		Dbrepo:        mockRepo,
		GrpcClient:    mockGrpc,
		KafkaProducer: mockKafka,
		Validator:     validator.New(),
	}
	user := &model.Person{
		Email:    "test@example.com",
		Password: "password123",
	}
	mockRepo.EXPECT().GetUser(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		user.Email,
		user.Password,
	).
		Return(&repository.DBRepositoryResponse{
			Success: true,
			UserId:  fixedUserId,
		})
	mockGrpc.EXPECT().
		CreateSession(
			mock.MatchedBy(func(ctx context.Context) bool {
				traceID := ctx.Value("traceID")
				return traceID != nil && traceID.(string) == fixedTraceID
			}),
			"123e4567-e89b-12d3-a456-426614174000",
		).
		Return(&proto.CreateSessionResponse{
			SessionID:  fixedSessionId,
			ExpiryTime: time.Now().Add(1 * time.Hour).Unix(),
			Success:    true,
		}, nil)
	mockKafka.EXPECT().NewUserLog(kafka.LogLevelInfo, place, fixedTraceID, "The session was created successfully and received")
	response := as.AuthenticateAndLogin(ctx, user)
	require.True(t, response.Success)
	require.Equal(t, fixedUserId, response.UserId)
	require.Equal(t, fixedSessionId, response.SessionId)
	require.NotNil(t, response.ExpireSession)
}
func TestAuthenticateAndLogin_ValidationErrors(t *testing.T) {
	var place = "UseCase-AuthenticateAndLogin"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	mockGrpc := mock_client.NewMockGrpcClientService(ctrl)
	mockKafka := mock_kafka.NewMockKafkaProducerService(ctrl)
	as := &service.AuthService{
		Dbrepo:        mockRepo,
		GrpcClient:    mockGrpc,
		KafkaProducer: mockKafka,
		Validator:     validator.New(),
	}
	tests := []struct {
		name           string
		user           *model.Person
		expectedError  map[string]error
		exprectedKafka *gomock.Call
		responseType   erro.ErrorType
	}{
		{
			name: "Invalid Email",
			user: &model.Person{
				Email:    "testexample.com",
				Password: "password123",
			},
			expectedError: map[string]error{
				"Email": erro.ErrorNotEmail,
			},
			exprectedKafka: mockKafka.EXPECT().NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, gomock.Any()),
			responseType:   erro.ClientErrorType,
		},
		{
			name: "Too Short Password",
			user: &model.Person{
				Email:    "valid@example.com",
				Password: "pas3",
			},
			expectedError: map[string]error{
				"Password": fmt.Errorf("Password is too short"),
			},
			exprectedKafka: mockKafka.EXPECT().NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, gomock.Any()),
			responseType:   erro.ClientErrorType,
		},
		{
			name: "Missing Required Fields",
			user: &model.Person{
				Email: "",
				Name:  "",
			},
			expectedError: map[string]error{
				"Email":    fmt.Errorf("Email is Null"),
				"Password": fmt.Errorf("Password is Null"),
			},
			exprectedKafka: mockKafka.EXPECT().NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, gomock.Any()).AnyTimes(),
			responseType:   erro.ClientErrorType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := as.AuthenticateAndLogin(ctx, tt.user)
			require.False(t, response.Success)
			require.Equal(t, tt.responseType, response.Type)
			require.Len(t, response.Errors, len(tt.expectedError))
			for field, expectedErr := range tt.expectedError {
				actualErr, exists := response.Errors[field]
				require.True(t, exists)
				require.EqualError(t, actualErr, expectedErr.Error())
			}
		})
	}
}
func TestAuthenticateAndLogin_DataBaseError_ClientError(t *testing.T) {
	fixedTraceUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid.String())
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	as := &service.AuthService{
		Dbtxmanager: mockTxManager,
		Validator:   validator.New(),
		Dbrepo:      mockRepo,
	}
	user := &model.Person{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	mockRepo.EXPECT().GetUser(ctx, user.Email, user.Password).Return(&repository.DBRepositoryResponse{Success: false, Errors: erro.ErrorInvalidPassword, Type: erro.ClientErrorType})
	response := as.AuthenticateAndLogin(ctx, user)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "ClientError")
	require.EqualError(t, response.Errors["ClientError"], erro.ErrorInvalidPassword.Error())
	require.Equal(t, erro.ClientErrorType, response.Type)
}
func TestAuthenticateAndLogin_DataBaseError_InternalServerError(t *testing.T) {
	fixedTraceUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid.String())
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	as := &service.AuthService{
		Dbtxmanager: mockTxManager,
		Validator:   validator.New(),
		Dbrepo:      mockRepo,
	}
	user := &model.Person{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	mockRepo.EXPECT().GetUser(ctx, user.Email, user.Password).Return(&repository.DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError, Type: erro.ServerErrorType})
	response := as.AuthenticateAndLogin(ctx, user)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], erro.ErrorDbRepositoryError.Error())
	require.Equal(t, erro.ServerErrorType, response.Type)
}
func TestAuthenticateAndLogin_RetryGrpc_ContextCanceled(t *testing.T) {
	var place = "UseCase-AuthenticateAndLogin"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
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
	user := &model.Person{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	mockRepo.EXPECT().GetUser(ctx, user.Email, user.Password).Return(&repository.DBRepositoryResponse{Success: true, UserId: fixedUUID})
	mockKafka.EXPECT().NewUserLog(kafka.LogLevelError, place, fixedTraceID, gomock.Any())
	response := as.AuthenticateAndLogin(ctx, user)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], erro.ErrorContextTimeout.Error())
	require.Equal(t, erro.ServerErrorType, response.Type)
}
func TestAuthenticateAndLogin_RetryGrpc_ClientError(t *testing.T) {
	var place = "UseCase-AuthenticateAndLogin"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
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
	user := &model.Person{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	mockRepo.EXPECT().GetUser(ctx, user.Email, user.Password).Return(&repository.DBRepositoryResponse{Success: true, UserId: fixedUUID})
	mockGrpcClient.EXPECT().
		CreateSession(mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
			fixedUUID.String()).
		Return(&pb.CreateSessionResponse{
			Success: false}, status.Error(codes.InvalidArgument, "Session not found")).
		Times(1)
	gomock.InOrder(
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, "Operation attempt 1 failed"),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, "Session not found"),
	)
	response := as.AuthenticateAndLogin(ctx, user)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "ClientError")
	require.EqualError(t, response.Errors["ClientError"], "Session not found")
	require.Equal(t, erro.ClientErrorType, response.Type)
}
func TestAuthenticateAndLogin_RetryGrpc_InternalServerError(t *testing.T) {
	var place = "UseCase-AuthenticateAndLogin"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
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
	user := &model.Person{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	mockRepo.EXPECT().GetUser(ctx, user.Email, user.Password).Return(&repository.DBRepositoryResponse{Success: true, UserId: fixedUUID})
	mockGrpcClient.EXPECT().
		CreateSession(mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
			fixedUUID.String()).
		Return(&pb.CreateSessionResponse{
			Success: false}, status.Error(codes.Internal, "Hset session Error")).
		Times(3)
	gomock.InOrder(
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, "Operation attempt 1 failed"),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, gomock.Any()),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, "Operation attempt 2 failed"),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, gomock.Any()),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, "Operation attempt 3 failed"),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, gomock.Any()),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelError, place, fixedTraceID, "All retry attempts failed"),
	)
	response := as.AuthenticateAndLogin(ctx, user)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], erro.ErrorAllRetryFailed.Error())
	require.Equal(t, erro.ServerErrorType, response.Type)
}
