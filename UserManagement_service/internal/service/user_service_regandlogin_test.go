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

func TestRegistrateAndLogin_Success(t *testing.T) {
	var place = "UseCase-RegistrateAndLogin"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	user := &model.Person{
		Name:     "testname",
		Email:    "test@example.com",
		Password: "password123"}
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
	mockRepo.EXPECT().CreateUser(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx, user,
	).
		Return(&repository.DBRepositoryResponse{
			Success: true,
			UserId:  fixedUUID,
		})
	mockGrpc.EXPECT().
		CreateSession(
			mock.MatchedBy(func(ctx context.Context) bool {
				traceID := ctx.Value("traceID")
				return traceID != nil && traceID.(string) == fixedTraceID
			}),
			fixedSessId,
		).
		Return(&proto.CreateSessionResponse{
			SessionID:  fixedSessId,
			ExpiryTime: time.Now().Add(1 * time.Hour).Unix(),
			Success:    true,
		}, nil)
	mockTxManager.EXPECT().CommitTx(tx).Return(nil)
	mockKafka.EXPECT().NewUserLog(kafka.LogLevelInfo, place, fixedTraceID, "Successful commit on attempt 1")
	mockKafka.EXPECT().NewUserLog(kafka.LogLevelInfo, place, fixedTraceID, "Transaction was successfully committed and session received")
	response := as.RegistrateAndLogin(ctx, user)
	require.True(t, response.Success)
	require.Equal(t, fixedUUID, response.UserId)
	require.Equal(t, fixedSessId, response.SessionId)
	require.NotNil(t, response.ExpireSession)
}
func TestRegistrateAndLogin_ValidationErrors(t *testing.T) {
	var place = "UseCase-RegistrateAndLogin"
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
				Name:     "testname",
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
				Name:     "testname",
				Password: "pas3",
			},
			expectedError: map[string]error{
				"Password": fmt.Errorf("Password is too short"),
			},
			exprectedKafka: mockKafka.EXPECT().NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, gomock.Any()),
			responseType:   erro.ClientErrorType,
		},
		{
			name: "Too Short Name",
			user: &model.Person{
				Email:    "valid@example.com",
				Name:     "te",
				Password: "password3",
			},
			expectedError: map[string]error{
				"Name": fmt.Errorf("Name is too short"),
			},
			exprectedKafka: mockKafka.EXPECT().NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, gomock.Any()),
			responseType:   erro.ClientErrorType,
		},
		{
			name: "Too Short Name and Password",
			user: &model.Person{
				Email:    "valid@example.com",
				Name:     "te",
				Password: "pas3",
			},
			expectedError: map[string]error{
				"Name":     fmt.Errorf("Name is too short"),
				"Password": fmt.Errorf("Password is too short"),
			},
			exprectedKafka: mockKafka.EXPECT().NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, gomock.Any()),
			responseType:   erro.ClientErrorType,
		},
		{
			name: "Missing Required Fields",
			user: &model.Person{
				Email:    "",
				Name:     "",
				Password: "",
			},
			expectedError: map[string]error{
				"Email":    fmt.Errorf("Email is Null"),
				"Password": fmt.Errorf("Password is Null"),
				"Name":     fmt.Errorf("Name is Null"),
			},
			exprectedKafka: mockKafka.EXPECT().NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, gomock.Any()).AnyTimes(),
			responseType:   erro.ClientErrorType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := as.RegistrateAndLogin(ctx, tt.user)
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
func TestRegistrateAndLogin_BeginTxError(t *testing.T) {
	var place = "UseCase-RegistrateAndLogin"
	fixedTraceUuid := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	user := &model.Person{
		Name:     "testname",
		Email:    "test@example.com",
		Password: "password123"}
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
	response := as.RegistrateAndLogin(ctx, user)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], erro.UserServiceUnavalaible)
	require.Equal(t, erro.ServerErrorType, response.Type)
}
func TestRegistrateAndLogin_DataBaseError_InternalServerError(t *testing.T) {
	var place = "UseCase-RegistrateAndLogin"
	fixedTraceUuid := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	user := &model.Person{
		Name:     "testname",
		Email:    "test@example.com",
		Password: "password123"}
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
	mockRepo.EXPECT().CreateUser(ctx, tx, user).Return(&repository.DBRepositoryResponse{Success: false, Errors: fmt.Errorf(erro.UserServiceUnavalaible), Type: erro.ServerErrorType})
	mockTxManager.EXPECT().RollbackTx(tx).Return(nil)
	mockKafka.EXPECT().NewUserLog(kafka.LogLevelInfo, place, fixedTraceUuid, "Successful rollback on attempt 1")
	response := as.RegistrateAndLogin(ctx, user)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], erro.UserServiceUnavalaible)
	require.Equal(t, erro.ServerErrorType, response.Type)
}
func TestRegistrateAndLogin_DataBaseError_ClientError(t *testing.T) {
	var place = "UseCase-RegistrateAndLogin"
	fixedTraceUuid := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	user := &model.Person{
		Name:     "testname",
		Email:    "test@example.com",
		Password: "password123"}
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
	mockRepo.EXPECT().CreateUser(ctx, tx, user).Return(&repository.DBRepositoryResponse{Success: false, Errors: erro.ErrorUniqueEmail, Type: erro.ClientErrorType})
	mockTxManager.EXPECT().RollbackTx(tx).Return(nil)
	mockKafka.EXPECT().NewUserLog(kafka.LogLevelInfo, place, fixedTraceUuid, "Successful rollback on attempt 1")
	response := as.RegistrateAndLogin(ctx, user)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "ClientError")
	require.EqualError(t, response.Errors["ClientError"], erro.ErrorUniqueEmail.Error())
	require.Equal(t, erro.ClientErrorType, response.Type)
}
func TestRegistrateAndLogin_RetryGrpc_InternalServerError(t *testing.T) {
	var place = "UseCase-RegistrateAndLogin"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	user := &model.Person{
		Name:     "testname",
		Email:    "test@example.com",
		Password: "password123"}
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
	mockRepo.EXPECT().CreateUser(ctx, tx, user).Return(&repository.DBRepositoryResponse{Success: true, UserId: fixedUUID})
	mockGrpcClient.EXPECT().
		CreateSession(mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
			fixedUUID.String()).
		Return(&pb.CreateSessionResponse{
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
	response := as.RegistrateAndLogin(ctx, user)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], erro.SessionServiceUnavalaible)
	require.Equal(t, erro.ServerErrorType, response.Type)
}
func TestRegistrateAndLogin_RetryGrpc_ContextCanceled(t *testing.T) {
	var place = "UseCase-RegistrateAndLogin"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cancel()
	user := &model.Person{
		Name:     "testname",
		Email:    "test@example.com",
		Password: "password123"}
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
	mockRepo.EXPECT().CreateUser(ctx, tx, user).Return(&repository.DBRepositoryResponse{Success: true, UserId: fixedUUID})
	gomock.InOrder(
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelError, place, fixedTraceID, gomock.Any()),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelInfo, place, fixedTraceID, "Successful rollback on attempt 1"),
	)
	mockTxManager.EXPECT().RollbackTx(tx).Return(nil)
	response := as.RegistrateAndLogin(ctx, user)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], "Request timed out")
	require.Equal(t, erro.ServerErrorType, response.Type)
}
func TestRegistrateAndLogin_RetryGrpc_ClientError(t *testing.T) {
	var place = "UseCase-RegistrateAndLogin"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	user := &model.Person{
		Name:     "testname",
		Email:    "test@example.com",
		Password: "password123"}
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
	mockRepo.EXPECT().CreateUser(ctx, tx, user).Return(&repository.DBRepositoryResponse{Success: true, UserId: fixedUUID})
	mockGrpcClient.EXPECT().
		CreateSession(mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
			fixedUUID.String()).
		Return(&pb.CreateSessionResponse{
			Success: false}, status.Error(codes.InvalidArgument, "UserID is required")).
		Times(1)
	gomock.InOrder(
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, "Operation attempt 1 failed"),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelWarn, place, fixedTraceID, "UserID is required"),
		mockKafka.EXPECT().
			NewUserLog(kafka.LogLevelInfo, place, fixedTraceID, "Successful rollback on attempt 1"),
	)
	mockTxManager.EXPECT().RollbackTx(tx).Return(nil)
	response := as.RegistrateAndLogin(ctx, user)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "ClientError")
	require.EqualError(t, response.Errors["ClientError"], "UserID is required")
	require.Equal(t, erro.ClientErrorType, response.Type)
}
func TestRegistrateAndLogin_CommitError(t *testing.T) {
	var place = "UseCase-RegistrateAndLogin"
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	user := &model.Person{
		Name:     "testname",
		Email:    "test@example.com",
		Password: "password123"}
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
	mockRepo.EXPECT().CreateUser(ctx, tx, user).Return(&repository.DBRepositoryResponse{Success: true, UserId: fixedUUID})
	mockGrpcClient.EXPECT().CreateSession(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}), mock.MatchedBy(func(userid string) bool {
		return userid == fixedUUID.String()
	})).Return(&pb.CreateSessionResponse{
		Success:   true,
		SessionID: fixedSessId,
	}, nil)
	mockTxManager.EXPECT().CommitTx(tx).Return(fmt.Errorf("Failed to commit transaction after all attempts")).Times(3)
	mockGrpcClient.EXPECT().DeleteSession(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}), mock.MatchedBy(func(sessionid string) bool {
		return sessionid == fixedSessId
	})).Return(&pb.DeleteSessionResponse{
		Success: true,
	}, nil)
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
	response := as.RegistrateAndLogin(ctx, user)
	require.False(t, response.Success)
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], erro.UserServiceUnavalaible)
	require.Equal(t, erro.ServerErrorType, response.Type)
}
