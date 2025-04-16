package service_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	mock_client "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/client/mocks"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	mock_repository "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository/mocks"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteAccount_Success(t *testing.T) {
	fixedReqUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "requestID", fixedReqUuid.String())
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	password := "password123"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	mockGrpc := mock_client.NewMockGrpcClientService(ctrl)

	as := &service.AuthService{
		Dbrepo:      mockRepo,
		Dbtxmanager: mockTxManager,
		GrpcClient:  mockGrpc,
		Validator:   validator.New(),
	}

	tx := &sql.Tx{}
	gomock.InOrder(
		mockTxManager.EXPECT().BeginTx(mock.MatchedBy(func(ctx context.Context) bool {
			requestID := ctx.Value("requestID")
			return requestID != nil && requestID.(string) == fixedReqUuid.String()
		})).Return(tx, nil),
		mockRepo.EXPECT().DeleteUser(
			mock.MatchedBy(func(ctx context.Context) bool {
				requestID := ctx.Value("requestID")
				return requestID != nil && requestID.(string) == fixedReqUuid.String()
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
		}),
		mockGrpc.EXPECT().DeleteSession(mock.MatchedBy(func(ctx context.Context) bool {
			requestID := ctx.Value("requestID")
			return requestID != nil && requestID.(string) == fixedReqUuid.String()
		}), mock.MatchedBy(func(sessionID string) bool {
			return sessionID == fixedSessId
		})).Return(&pb.DeleteSessionResponse{
			Success: true,
		}, nil),
		mockTxManager.EXPECT().CommitTx(tx).Return(nil),
		mockTxManager.EXPECT().RollbackTx(tx).AnyTimes(),
	)
	response := as.DeleteAccount(ctx, fixedSessId, fixedUUID, password)
	log.Printf("Response: %+v", response)
	require.True(t, response.Success)
}
func TestDeleteAccount_MissingReqId(t *testing.T) {
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	password := "password123"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	mockGrpc := mock_client.NewMockGrpcClientService(ctrl)

	as := &service.AuthService{
		Dbrepo:      mockRepo,
		Dbtxmanager: mockTxManager,
		GrpcClient:  mockGrpc,
		Validator:   validator.New(),
	}
	mockTxManager.EXPECT().BeginTx(gomock.Any()).Times(0)
	response := as.DeleteAccount(ctx, fixedSessId, fixedUUID, password)
	log.Printf("Response: %+v", response)

	require.False(t, response.Success)
	require.EqualError(t, response.Errors["ContextError"], "Error missing request ID")
}
func TestDeleteAccount_BeginTxError(t *testing.T) {
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	fixedReqUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctx := context.WithValue(context.Background(), "requestID", fixedReqUuid.String())
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	password := "password123"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)

	as := &service.AuthService{
		Dbtxmanager: mockTxManager,
		Validator:   validator.New(),
	}

	mockTxManager.EXPECT().BeginTx(gomock.Any()).Return(nil, fmt.Errorf("database connection error"))

	response := as.DeleteAccount(ctx, fixedSessId, fixedUUID, password)
	log.Printf("Response: %+v", response)

	require.False(t, response.Success)
	require.Contains(t, response.Errors, "TransactionError")
	require.EqualError(t, response.Errors["TransactionError"], "Transaction creation error")
}
func TestDeleteAccount_ContextBeforeDeleteUser(t *testing.T) {
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	fixedReqUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ctx = context.WithValue(ctx, "requestID", fixedReqUuid.String())
	password := "password123"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)

	as := &service.AuthService{
		Dbtxmanager: mockTxManager,
		Validator:   validator.New(),
	}

	mockTxManager.EXPECT().BeginTx(gomock.Any()).Return(nil, nil)
	response := as.DeleteAccount(ctx, fixedSessId, fixedUUID, password)
	log.Printf("Response: %+v", response)

	require.False(t, response.Success)
	require.Contains(t, response.Errors, "ContextError")
	require.EqualError(t, response.Errors["ContextError"], "The timeout context has expired")
}
func TestDeleteAccount_DeleteUserError(t *testing.T) {
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	fixedReqUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, "requestID", fixedReqUuid.String())

	password := "password123"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)

	as := &service.AuthService{
		Dbrepo:      mockRepo,
		Dbtxmanager: mockTxManager,
		Validator:   validator.New(),
	}

	tx := &sql.Tx{}
	mockTxManager.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)

	mockRepo.EXPECT().DeleteUser(ctx, tx, gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, tx *sql.Tx, userid uuid.UUID, password string) *repository.DBRepositoryResponse {
		return &repository.DBRepositoryResponse{
			Success: false,
			Errors:  erro.ErrorInvalidPassword,
		}
	})

	mockTxManager.EXPECT().RollbackTx(tx).AnyTimes()

	response := as.DeleteAccount(ctx, fixedSessId, fixedUUID, password)
	log.Printf("Response: %+v", response)

	require.False(t, response.Success)
	require.Contains(t, response.Errors, "DeleteError")
	require.EqualError(t, response.Errors["DeleteError"], "Invalid Password")
}
func TestDeleteAccount_DeleteSessionError(t *testing.T) {

	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	fixedReqUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, "requestID", fixedReqUuid.String())

	password := "password123"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	mockGrpcClient := mock_client.NewMockGrpcClientService(ctrl)

	as := &service.AuthService{
		Dbrepo:      mockRepo,
		Dbtxmanager: mockTxManager,
		Validator:   validator.New(),
		GrpcClient:  mockGrpcClient,
	}

	tx := &sql.Tx{}
	mockTxManager.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)

	mockRepo.EXPECT().DeleteUser(
		ctx,
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

	mockTxManager.EXPECT().RollbackTx(tx).AnyTimes()

	mockGrpcClient.EXPECT().DeleteSession(ctx, fixedSessId).DoAndReturn(func(ctx context.Context, session string) (*pb.CreateSessionResponse, error) {
		return nil, fmt.Errorf("grpc error")
	})

	response := as.DeleteAccount(ctx, fixedSessId, fixedUUID, password)
	log.Printf("Response: %+v", response)

	require.False(t, response.Success)
	require.Contains(t, response.Errors, "GrpcResponseError")
	require.EqualError(t, response.Errors["GrpcResponseError"], "Grpc's response error")
}
func TestDeleteAccount_CommitTxError(t *testing.T) {

	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	fixedReqUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, "requestID", fixedReqUuid.String())

	password := "password123"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	mockGrpcClient := mock_client.NewMockGrpcClientService(ctrl)

	as := &service.AuthService{
		Dbrepo:      mockRepo,
		Dbtxmanager: mockTxManager,
		Validator:   validator.New(),
		GrpcClient:  mockGrpcClient,
	}

	tx := &sql.Tx{}
	mockTxManager.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)

	mockRepo.EXPECT().DeleteUser(
		ctx,
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
	mockTxManager.EXPECT().CommitTx(tx).Return(fmt.Errorf("commit error"))
	mockTxManager.EXPECT().RollbackTx(tx).AnyTimes()

	mockGrpcClient.EXPECT().DeleteSession(ctx, fixedSessId).Return(&pb.DeleteSessionResponse{
		Success: true,
	}, nil)

	response := as.DeleteAccount(ctx, fixedSessId, fixedUUID, password)
	log.Printf("Response: %+v", response)

	require.False(t, response.Success)
	require.Contains(t, response.Errors, "TransactionError")
	require.EqualError(t, response.Errors["TransactionError"], erro.ErrorCommitTransaction.Error())
}
