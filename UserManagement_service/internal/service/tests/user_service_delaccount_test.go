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
	fixedTraceUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid.String())
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
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
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceUuid.String()
		})).Return(tx, nil),
		mockRepo.EXPECT().DeleteUser(
			mock.MatchedBy(func(ctx context.Context) bool {
				traceID := ctx.Value("traceID")
				return traceID != nil && traceID.(string) == fixedTraceUuid.String()
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
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceUuid.String()
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
func TestDeleteAccount_BeginTxError(t *testing.T) {
	expectedTypeError := erro.ServerErrorType
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedSessId := "123e4567-e89b-12d3-a456-426614174000"
	fixedTraceUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid.String())
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
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
	require.Contains(t, response.Errors, "InternalServerError")
	require.EqualError(t, response.Errors["InternalServerError"], "Transaction creation error")
	require.Equal(t, expectedTypeError, response.Type)
}
