package service_test

import (
	"context"
	"database/sql"
	"log"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	mock_client "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/client/mocks"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	mock_repository "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository/mocks"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteAccount_Success(t *testing.T) {
	ctx := context.Background()
	user := &model.Person{
		Password: "password123",
	}

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
	fixedsessUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614171000")
	userUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	mockTxManager.EXPECT().BeginTx(ctx).Return(tx, nil)
	mockTxManager.EXPECT().CommitTx(tx).Return(nil)
	mockTxManager.EXPECT().RollbackTx(gomock.Any()).AnyTimes()
	mockRepo.EXPECT().DeleteUser(
		ctx,
		tx,
		mock.MatchedBy(func(userid uuid.UUID) bool {
			return userid == userUUID
		}),
		mock.MatchedBy(func(password string) bool {
			return password == "password123"
		}),
	).Return(&repository.DBRepositoryResponse{
		Success: true,
	})
	mockGrpc.EXPECT().DeleteSession(ctx, mock.MatchedBy(func(sessionID string) bool {
		return sessionID == "123e4567-e89b-12d3-a456-426614171000"
	})).Return(&pb.DeleteSessionResponse{
		Success: true,
	}, nil)

	response := as.DeleteAccount(ctx, fixedsessUUID.String(), userUUID, user.Password)
	log.Printf("Response: %+v", response)

	require.True(t, response.Success)
}
