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

func TestDeleteAccount_Success(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-426614171000"
	fixedUserId := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
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
	req := &model.DeletionRequest{
		Password: "password123",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
	).Return(tx, nil)
	mockUserRepo.EXPECT().DeleteUser(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx,
		fixedUserId,
		req.Password).Return(&repository.RepositoryResponse{
		Success:        true,
		Place:          repository.DeleteUser,
		SuccessMessage: "Successful delete user in database",
	})
	mockCacheRepo.EXPECT().DeleteProfileCache(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}),
		fixedUserId.String(),
	).Return(&repository.RepositoryResponse{Success: true, Place: repository.DeleteProfileCache, SuccessMessage: "Successful delete profile from cache"})
	mockSessionClient.EXPECT().DeleteSession(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		fixedSessionId,
	).Return(&proto.DeleteSessionResponse{
		Success: true,
	}, nil)
	mockEventProducer.EXPECT().NewUserEvent(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		rabbitmq.UserDeleteKey,
		fixedUserId.String(),
		gomock.Any(),
		fixedTraceID,
	).Return(nil)
	mockTransactionRepo.EXPECT().CommitTx(ctx, tx).Return(nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.DeleteAccount(ctx, req, fixedSessionId, fixedUserId.String())
	require.True(t, response.Success)
}
func TestDeleteAccount_ParsingUserID_Error(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-426614171000"
	fixedUserId := "123e4567-e89b-12d3-a456-42661417400"
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
	req := &model.DeletionRequest{
		Password: "password123",
	}
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.DeleteAccount(ctx, req, fixedSessionId, fixedUserId)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ClientErrorType, Message: erro.ErrorInvalidUserIDFormat})
}
func TestDeleteAccount_ValidationError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-426614171000"
	fixedUserId := "123e4567-e89b-12d3-a456-426614174000"
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
	req := &model.DeletionRequest{
		Password: "pas",
	}
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.DeleteAccount(ctx, req, fixedSessionId, fixedUserId)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ClientErrorType, Message: "Password is too short"})
}
func TestDeleteAccount_BeginTxError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-426614171000"
	fixedUserId := "123e4567-e89b-12d3-a456-426614174000"
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
	req := &model.DeletionRequest{
		Password: "password123",
	}
	mockTransactionRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, fmt.Errorf("database connection error"))
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.DeleteAccount(ctx, req, fixedSessionId, fixedUserId)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}
func TestDeleteAccount_DataBaseError_ClientError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-426614171000"
	fixedUserId := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
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
	req := &model.DeletionRequest{
		Password: "password123",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		})).Return(tx, nil)
	mockUserRepo.EXPECT().DeleteUser(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx,
		fixedUserId,
		req.Password).Return(&repository.RepositoryResponse{
		Success: false,
		Place:   repository.DeleteUser,
		Errors:  &erro.CustomError{Type: erro.ClientErrorType, Message: erro.ErrorIncorrectPassword},
	})
	mockTransactionRepo.EXPECT().RollbackTx(ctx, tx).Return(nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.DeleteAccount(ctx, req, fixedSessionId, fixedUserId.String())
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ClientErrorType, Message: erro.ErrorIncorrectPassword})
}
func TestDeleteAccount_DataBaseError_InternalServerError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-426614171000"
	fixedUserId := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
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
	req := &model.DeletionRequest{
		Password: "password123",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		})).Return(tx, nil)
	mockUserRepo.EXPECT().DeleteUser(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx,
		fixedUserId,
		req.Password).Return(&repository.RepositoryResponse{
		Success: false,
		Place:   repository.DeleteUser,
		Errors:  &erro.CustomError{Type: erro.ServerErrorType, Message: erro.ErrorAfterReqUsers},
	})
	mockTransactionRepo.EXPECT().RollbackTx(ctx, tx).Return(nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.DeleteAccount(ctx, req, fixedSessionId, fixedUserId.String())
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}
func TestDeleteAccount_RetryGrpc_InternalServerError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-426614171000"
	fixedUserId := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
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
	req := &model.DeletionRequest{
		Password: "password123",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		})).Return(tx, nil)
	mockUserRepo.EXPECT().DeleteUser(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx,
		fixedUserId,
		req.Password).Return(&repository.RepositoryResponse{
		Success:        true,
		Place:          repository.DeleteUser,
		SuccessMessage: "Successful delete user in database",
	})
	mockCacheRepo.EXPECT().DeleteProfileCache(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}),
		fixedUserId.String(),
	).Return(&repository.RepositoryResponse{Success: true, Place: repository.DeleteProfileCache, SuccessMessage: "Successful delete profile from cache"})
	mockSessionClient.EXPECT().
		DeleteSession(mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
			fixedSessionId).
		Return(&pb.DeleteSessionResponse{
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
			NewUserLog(kafka.LogLevelError, gomock.Any(), fixedTraceID, "All retry attempts failed"))
	mockTransactionRepo.EXPECT().RollbackTx(ctx, tx).Return(nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.DeleteAccount(ctx, req, fixedSessionId, fixedUserId.String())
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.SessionServiceUnavalaible})
}
func TestDeleteAccount_DeleteCacheError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUserId := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
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
	req := &model.DeletionRequest{
		Password: "password123",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
	).Return(tx, nil)
	mockUserRepo.EXPECT().DeleteUser(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx,
		fixedUserId,
		req.Password).Return(&repository.RepositoryResponse{
		Success:        true,
		Place:          repository.DeleteUser,
		SuccessMessage: "Successful delete user in database",
	})
	mockCacheRepo.EXPECT().DeleteProfileCache(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}),
		fixedUserId.String(),
	).Return(&repository.RepositoryResponse{Success: false, Place: repository.DeleteProfileCache, Errors: &erro.CustomError{Type: erro.ServerErrorType, Message: erro.ErrorDelProfiles}})
	mockTransactionRepo.EXPECT().RollbackTx(ctx, tx).Return(nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.DeleteAccount(ctx, req, fixedSessionId, fixedUserId.String())
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}
func TestDeleteAccount_EventProducerError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-426614171000"
	fixedUserId := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
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
	req := &model.DeletionRequest{
		Password: "password123",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
	).Return(tx, nil)
	mockUserRepo.EXPECT().DeleteUser(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx,
		fixedUserId,
		req.Password).Return(&repository.RepositoryResponse{
		Success:        true,
		Place:          repository.DeleteUser,
		SuccessMessage: "Successful delete user in database",
	})
	mockCacheRepo.EXPECT().DeleteProfileCache(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}),
		fixedUserId.String(),
	).Return(&repository.RepositoryResponse{Success: true, Place: repository.DeleteProfileCache, SuccessMessage: "Successful delete profile from cache"})
	mockSessionClient.EXPECT().DeleteSession(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		fixedSessionId,
	).Return(&proto.DeleteSessionResponse{
		Success: true,
	}, nil)
	mockEventProducer.EXPECT().NewUserEvent(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		rabbitmq.UserDeleteKey,
		fixedUserId.String(),
		gomock.Any(),
		fixedTraceID,
	).Return(fmt.Errorf("rabbitMQ error"))
	mockTransactionRepo.EXPECT().RollbackTx(ctx, tx).Return(nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.DeleteAccount(ctx, req, fixedSessionId, fixedUserId.String())
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}
func TestDeleteAccount_CommitTransactionError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedSessionId := "123e4567-e89b-12d3-a456-426614171000"
	fixedUserId := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
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
	req := &model.DeletionRequest{
		Password: "password123",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
	).Return(tx, nil)
	mockUserRepo.EXPECT().DeleteUser(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx,
		fixedUserId,
		req.Password).Return(&repository.RepositoryResponse{
		Success:        true,
		Place:          repository.DeleteUser,
		SuccessMessage: "Successful delete user in database",
	})
	mockCacheRepo.EXPECT().DeleteProfileCache(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}),
		fixedUserId.String(),
	).Return(&repository.RepositoryResponse{Success: true, Place: repository.DeleteProfileCache, SuccessMessage: "Successful delete profile from cache"})
	mockSessionClient.EXPECT().DeleteSession(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		fixedSessionId,
	).Return(&proto.DeleteSessionResponse{
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
	mockTransactionRepo.EXPECT().CommitTx(ctx, tx).Return(fmt.Errorf("Failed to commit transaction after all attempts")).Times(3)
	mockTransactionRepo.EXPECT().RollbackTx(ctx, tx).Return(nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.DeleteAccount(ctx, req, fixedSessionId, fixedUserId.String())
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}
