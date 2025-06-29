package service_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/rabbitmq"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
	mock_service "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRegistrateAndLogin_Success(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-426614171000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := mock_service.NewMockDBUserRepos(ctrl)
	mockSessionClient := mock_service.NewMockSessionClient(ctrl)
	mockLogProducer := mock_service.NewMockLogProducer(ctrl)
	mockEventProducer := mock_service.NewMockEventProducer(ctrl)
	mockTransactionRepo := mock_service.NewMockDBTransactionManager(ctrl)
	mockCacheRepo := mock_service.NewMockCacheUserRepos(ctrl)
	as := &service.UserService{
		Dbrepo:            mockUserRepo,
		Validator:         validator.New(),
		GrpcSessionClient: mockSessionClient,
		LogProducer:       mockLogProducer,
		EventProducer:     mockEventProducer,
		Dbtxmanager:       mockTransactionRepo,
		CacheUserRepos:    mockCacheRepo,
	}
	tx := &sql.Tx{}
	req := &model.RegistrationRequest{
		Name:     "tester",
		Email:    "test@example.com",
		Password: "password123",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
	).Return(tx, nil)
	mockUserRepo.EXPECT().CreateUser(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx,
		mock.MatchedBy(func(u *model.User) bool {
			return u.Name == "tester" && u.Email == "test@example.com"
		}),
	).Return(&repository.RepositoryResponse{
		Success:        true,
		Place:          repository.CreateUser,
		SuccessMessage: "Successful create user in database",
	})
	mockSessionClient.EXPECT().CreateSession(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		gomock.Any(),
	).Return(&proto.CreateSessionResponse{
		SessionID:  fixedSessionId,
		ExpiryTime: time.Now().Add(1 * time.Hour).Unix(),
		Success:    true,
	}, nil)

	mockEventProducer.EXPECT().NewUserEvent(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		rabbitmq.UserRegistrationKey,
		gomock.Any(),
		gomock.Any(),
		fixedTraceID,
	).Return(nil)
	mockTransactionRepo.EXPECT().CommitTx(ctx, tx).Return(nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.RegistrateAndLogin(ctx, req)
	require.True(t, response.Success)
	require.NotEmpty(t, response.Data[repository.KeyUserID])
	require.Equal(t, fixedSessionId, response.Data[service.KeySessionID])
}
func TestRegistrateAndLogin_ValidationErrors(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := mock_service.NewMockDBUserRepos(ctrl)
	mockSessionClient := mock_service.NewMockSessionClient(ctrl)
	mockLogProducer := mock_service.NewMockLogProducer(ctrl)
	mockEventProducer := mock_service.NewMockEventProducer(ctrl)
	mockTransactionRepo := mock_service.NewMockDBTransactionManager(ctrl)
	mockCacheRepo := mock_service.NewMockCacheUserRepos(ctrl)
	as := &service.UserService{
		Dbrepo:            mockUserRepo,
		Validator:         validator.New(),
		GrpcSessionClient: mockSessionClient,
		LogProducer:       mockLogProducer,
		EventProducer:     mockEventProducer,
		Dbtxmanager:       mockTransactionRepo,
		CacheUserRepos:    mockCacheRepo,
	}
	tests := []struct {
		name          string
		req           *model.RegistrationRequest
		expectedError *erro.CustomError
		logproducer   *gomock.Call
	}{
		{
			name: "Invalid Email Format",
			req: &model.RegistrationRequest{
				Email:    "testexample.com",
				Password: "password123",
				Name:     "tester",
			},
			expectedError: &erro.CustomError{Type: erro.ClientErrorType, Message: erro.ErrorNotEmailConst},
			logproducer:   mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes(),
		},
		{
			name: "Too Short Password",
			req: &model.RegistrationRequest{
				Email:    "valid@example.com",
				Password: "pas3",
				Name:     "tester",
			},
			expectedError: &erro.CustomError{Type: erro.ClientErrorType, Message: "Password is too short"},
			logproducer:   mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes(),
		},
		{
			name: "Too Long Password",
			req: &model.RegistrationRequest{
				Email:    "valid@example.com",
				Password: "pas3sdsssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssss",
				Name:     "tester",
			},
			expectedError: &erro.CustomError{Type: erro.ClientErrorType, Message: "Password is too long"},
			logproducer:   mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes(),
		},
		{
			name: "Too Short Name",
			req: &model.RegistrationRequest{
				Email:    "valid@example.com",
				Password: "pas3sdsssss",
				Name:     "te",
			},
			expectedError: &erro.CustomError{Type: erro.ClientErrorType, Message: "Name is too short"},
			logproducer:   mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes(),
		},
		{
			name: "Required Field",
			req: &model.RegistrationRequest{
				Email:    "valid@example.com",
				Password: "sdddsdscs",
			},
			expectedError: &erro.CustomError{Type: erro.ClientErrorType, Message: "Name is Null"},
			logproducer:   mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := as.RegistrateAndLogin(ctx, tt.req)
			require.False(t, response.Success)
			require.Equal(t, tt.expectedError, response.Errors)
		})
	}
}
func TestRegistrateAndLogin_BeginTxError(t *testing.T) {
	fixedTraceUuid := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := mock_service.NewMockDBUserRepos(ctrl)
	mockSessionClient := mock_service.NewMockSessionClient(ctrl)
	mockLogProducer := mock_service.NewMockLogProducer(ctrl)
	mockEventProducer := mock_service.NewMockEventProducer(ctrl)
	mockTransactionRepo := mock_service.NewMockDBTransactionManager(ctrl)
	mockCacheRepo := mock_service.NewMockCacheUserRepos(ctrl)
	as := &service.UserService{
		Dbrepo:            mockUserRepo,
		Validator:         validator.New(),
		GrpcSessionClient: mockSessionClient,
		LogProducer:       mockLogProducer,
		EventProducer:     mockEventProducer,
		Dbtxmanager:       mockTransactionRepo,
		CacheUserRepos:    mockCacheRepo,
	}
	req := &model.RegistrationRequest{
		Name:     "tester",
		Email:    "test@example.com",
		Password: "password123",
	}
	mockTransactionRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, fmt.Errorf("database connection error"))
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.RegistrateAndLogin(ctx, req)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}
func TestRegistrateAndLogin_DataBaseError_ClientError(t *testing.T) {
	fixedTraceUuid := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := mock_service.NewMockDBUserRepos(ctrl)
	mockSessionClient := mock_service.NewMockSessionClient(ctrl)
	mockLogProducer := mock_service.NewMockLogProducer(ctrl)
	mockEventProducer := mock_service.NewMockEventProducer(ctrl)
	mockTransactionRepo := mock_service.NewMockDBTransactionManager(ctrl)
	mockCacheRepo := mock_service.NewMockCacheUserRepos(ctrl)
	as := &service.UserService{
		Dbrepo:            mockUserRepo,
		Validator:         validator.New(),
		GrpcSessionClient: mockSessionClient,
		LogProducer:       mockLogProducer,
		EventProducer:     mockEventProducer,
		Dbtxmanager:       mockTransactionRepo,
		CacheUserRepos:    mockCacheRepo,
	}
	tx := &sql.Tx{}
	req := &model.RegistrationRequest{
		Name:     "tester",
		Email:    "test@example.com",
		Password: "password123",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceUuid
		})).Return(tx, nil)
	mockUserRepo.EXPECT().CreateUser(ctx, tx, mock.MatchedBy(func(u *model.User) bool {
		return u.Name == "tester" && u.Email == "test@example.com"
	})).Return(
		&repository.RepositoryResponse{Success: false, Errors: &erro.CustomError{Type: erro.ClientErrorType, Message: erro.ErrorUniqueEmailConst}, Place: repository.CreateUser},
	)
	mockTransactionRepo.EXPECT().RollbackTx(ctx, tx).Return(nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.RegistrateAndLogin(ctx, req)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ClientErrorType, Message: erro.ErrorUniqueEmailConst})
}
func TestRegistrateAndLogin_DataBaseError_InternalServerError(t *testing.T) {
	fixedTraceUuid := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := mock_service.NewMockDBUserRepos(ctrl)
	mockSessionClient := mock_service.NewMockSessionClient(ctrl)
	mockLogProducer := mock_service.NewMockLogProducer(ctrl)
	mockEventProducer := mock_service.NewMockEventProducer(ctrl)
	mockTransactionRepo := mock_service.NewMockDBTransactionManager(ctrl)
	mockCacheRepo := mock_service.NewMockCacheUserRepos(ctrl)
	as := &service.UserService{
		Dbrepo:            mockUserRepo,
		Validator:         validator.New(),
		GrpcSessionClient: mockSessionClient,
		LogProducer:       mockLogProducer,
		EventProducer:     mockEventProducer,
		Dbtxmanager:       mockTransactionRepo,
		CacheUserRepos:    mockCacheRepo,
	}
	tx := &sql.Tx{}
	req := &model.RegistrationRequest{
		Name:     "tester",
		Email:    "test@example.com",
		Password: "password123",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceUuid
		})).Return(tx, nil)
	mockUserRepo.EXPECT().CreateUser(ctx, tx, mock.MatchedBy(func(u *model.User) bool {
		return u.Name == "tester" && u.Email == "test@example.com"
	})).Return(
		&repository.RepositoryResponse{Success: false, Errors: &erro.CustomError{Type: erro.ServerErrorType, Message: erro.ErrorAfterReqUsers}, Place: repository.CreateUser},
	)
	mockTransactionRepo.EXPECT().RollbackTx(ctx, tx).Return(nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.RegistrateAndLogin(ctx, req)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}

func TestRegistrateAndLogin_RetryGrpc_InternalServerError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := mock_service.NewMockDBUserRepos(ctrl)
	mockSessionClient := mock_service.NewMockSessionClient(ctrl)
	mockLogProducer := mock_service.NewMockLogProducer(ctrl)
	mockEventProducer := mock_service.NewMockEventProducer(ctrl)
	mockTransactionRepo := mock_service.NewMockDBTransactionManager(ctrl)
	mockCacheRepo := mock_service.NewMockCacheUserRepos(ctrl)
	as := &service.UserService{
		Dbrepo:            mockUserRepo,
		Validator:         validator.New(),
		GrpcSessionClient: mockSessionClient,
		LogProducer:       mockLogProducer,
		EventProducer:     mockEventProducer,
		Dbtxmanager:       mockTransactionRepo,
		CacheUserRepos:    mockCacheRepo,
	}
	tx := &sql.Tx{}
	req := &model.RegistrationRequest{
		Name:     "tester",
		Email:    "test@example.com",
		Password: "password123",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		})).Return(tx, nil)
	mockUserRepo.EXPECT().CreateUser(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx,
		mock.MatchedBy(func(u *model.User) bool {
			return u.Name == "tester" && u.Email == "test@example.com"
		}),
	).Return(&repository.RepositoryResponse{
		Success:        true,
		Place:          repository.CreateUser,
		SuccessMessage: "Successful create user in database",
	})
	mockLogProducer.EXPECT().NewUserLog(kafka.LogLevelInfo, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockSessionClient.EXPECT().
		CreateSession(mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
			gomock.Any()).
		Return(&pb.CreateSessionResponse{
			Success: false}, status.Error(codes.Internal, erro.SessionServiceUnavalaible)).
		Times(3)
	gomock.InOrder(
		mockLogProducer.EXPECT().
			NewUserLog(kafka.LogLevelWarn, gomock.Any(), fixedTraceID, "Operation attempt 1 failed"),
		mockLogProducer.EXPECT().
			NewUserLog(kafka.LogLevelWarn, gomock.Any(), fixedTraceID, "Session-Service is unavailable, retrying..."),
		mockLogProducer.EXPECT().
			NewUserLog(kafka.LogLevelWarn, gomock.Any(), fixedTraceID, "Operation attempt 2 failed"),
		mockLogProducer.EXPECT().
			NewUserLog(kafka.LogLevelWarn, gomock.Any(), fixedTraceID, "Session-Service is unavailable, retrying..."),
		mockLogProducer.EXPECT().
			NewUserLog(kafka.LogLevelWarn, gomock.Any(), fixedTraceID, "Operation attempt 3 failed"),
		mockLogProducer.EXPECT().
			NewUserLog(kafka.LogLevelWarn, gomock.Any(), fixedTraceID, "Session-Service is unavailable, retrying..."),
		mockLogProducer.EXPECT().
			NewUserLog(kafka.LogLevelError, gomock.Any(), fixedTraceID, "All retry attempts failed"),
	)
	mockTransactionRepo.EXPECT().RollbackTx(ctx, tx).Return(nil)
	response := as.RegistrateAndLogin(ctx, req)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.SessionServiceUnavalaible})
}
func TestRegistrateAndLogin_EventProducerError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-426614171000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := mock_service.NewMockDBUserRepos(ctrl)
	mockSessionClient := mock_service.NewMockSessionClient(ctrl)
	mockLogProducer := mock_service.NewMockLogProducer(ctrl)
	mockEventProducer := mock_service.NewMockEventProducer(ctrl)
	mockTransactionRepo := mock_service.NewMockDBTransactionManager(ctrl)
	mockCacheRepo := mock_service.NewMockCacheUserRepos(ctrl)
	as := &service.UserService{
		Dbrepo:            mockUserRepo,
		Validator:         validator.New(),
		GrpcSessionClient: mockSessionClient,
		LogProducer:       mockLogProducer,
		EventProducer:     mockEventProducer,
		Dbtxmanager:       mockTransactionRepo,
		CacheUserRepos:    mockCacheRepo,
	}
	tx := &sql.Tx{}
	req := &model.RegistrationRequest{
		Name:     "tester",
		Email:    "test@example.com",
		Password: "password123",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
	).Return(tx, nil)
	mockUserRepo.EXPECT().CreateUser(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx,
		mock.MatchedBy(func(u *model.User) bool {
			return u.Name == "tester" && u.Email == "test@example.com"
		}),
	).Return(&repository.RepositoryResponse{
		Success:        true,
		Place:          repository.CreateUser,
		SuccessMessage: "Successful create user in database",
	})
	mockSessionClient.EXPECT().CreateSession(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		gomock.Any(),
	).Return(&proto.CreateSessionResponse{
		SessionID:  fixedSessionId,
		ExpiryTime: time.Now().Add(1 * time.Hour).Unix(),
		Success:    true,
	}, nil)

	mockEventProducer.EXPECT().NewUserEvent(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		rabbitmq.UserRegistrationKey,
		gomock.Any(),
		gomock.Any(),
		fixedTraceID,
	).Return(fmt.Errorf("rabbitMQ error"))
	mockTransactionRepo.EXPECT().RollbackTx(ctx, tx).Return(nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.RegistrateAndLogin(ctx, req)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}
func TestRegistrateAndLogin_CommitTransactionError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-426614171000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := mock_service.NewMockDBUserRepos(ctrl)
	mockSessionClient := mock_service.NewMockSessionClient(ctrl)
	mockLogProducer := mock_service.NewMockLogProducer(ctrl)
	mockEventProducer := mock_service.NewMockEventProducer(ctrl)
	mockTransactionRepo := mock_service.NewMockDBTransactionManager(ctrl)
	mockCacheRepo := mock_service.NewMockCacheUserRepos(ctrl)
	as := &service.UserService{
		Dbrepo:            mockUserRepo,
		Validator:         validator.New(),
		GrpcSessionClient: mockSessionClient,
		LogProducer:       mockLogProducer,
		EventProducer:     mockEventProducer,
		Dbtxmanager:       mockTransactionRepo,
		CacheUserRepos:    mockCacheRepo,
	}
	tx := &sql.Tx{}
	req := &model.RegistrationRequest{
		Name:     "tester",
		Email:    "test@example.com",
		Password: "password123",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
	).Return(tx, nil)
	mockUserRepo.EXPECT().CreateUser(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx,
		mock.MatchedBy(func(u *model.User) bool {
			return u.Name == "tester" && u.Email == "test@example.com"
		}),
	).Return(&repository.RepositoryResponse{
		Success:        true,
		Place:          repository.CreateUser,
		SuccessMessage: "Successful create user in database",
	})
	mockSessionClient.EXPECT().CreateSession(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		gomock.Any(),
	).Return(&proto.CreateSessionResponse{
		SessionID:  fixedSessionId,
		ExpiryTime: time.Now().Add(1 * time.Hour).Unix(),
		Success:    true,
	}, nil)

	mockEventProducer.EXPECT().NewUserEvent(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		rabbitmq.UserRegistrationKey,
		gomock.Any(),
		gomock.Any(),
		fixedTraceID,
	).Return(nil)
	mockTransactionRepo.EXPECT().CommitTx(ctx, tx).Return(fmt.Errorf("Failed to commit transaction after all attempts")).Times(3)
	mockSessionClient.EXPECT().DeleteSession(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}), mock.MatchedBy(func(sessionid string) bool {
		return sessionid == fixedSessionId
	})).Return(&pb.DeleteSessionResponse{
		Success: true,
	}, nil)
	mockEventProducer.EXPECT().NewUserEvent(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		rabbitmq.UserDeleteKey,
		gomock.Any(),
		gomock.Any(),
		fixedTraceID,
	).Return(nil)
	mockTransactionRepo.EXPECT().RollbackTx(ctx, tx).Return(nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.RegistrateAndLogin(ctx, req)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}
