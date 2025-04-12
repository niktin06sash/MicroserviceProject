package service_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

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

func TestRegistrateAndLogin_Success(t *testing.T) {
	ctx := context.Background()
	user := &model.Person{
		Name:     "John Doe",
		Email:    "john.doe@example.com",
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
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	mockTxManager.EXPECT().BeginTx(ctx).Return(tx, nil)
	mockTxManager.EXPECT().CommitTx(tx).Return(nil)

	mockRepo.EXPECT().CreateUser(ctx, mock.MatchedBy(func(user *model.Person) bool {
		return user.Name == "John Doe" && user.Email == "john.doe@example.com"
	})).Return(&repository.DBRepositoryResponse{
		Success: true,
		UserId:  fixedUUID,
	})

	mockGrpc.EXPECT().CreateSession(ctx, mock.MatchedBy(func(userID string) bool {
		_, err := uuid.Parse(userID)
		return err == nil
	})).Return(&pb.CreateSessionResponse{
		Success:    true,
		SessionID:  "session-id",
		ExpiryTime: time.Now().Unix(),
	}, nil)
	response := as.RegistrateAndLogin(ctx, user)
	require.True(t, response.Success)
	require.Equal(t, "session-id", response.SessionId)
}

/*func TestNewAuthService(t *testing.T) {
    // Создаем контроллер для моков
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // Создаем моки для зависимостей
	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	mockGrpc := mock_client.NewMockGrpcClientService(ctrl)
    mockKafkaProducer := kafka.NewMockKafkaProducer(ctrl)
    mockGrpcClient := client.NewMockGrpcClient(ctrl)

    // Вызываем конструктор
    authService := service.NewAuthService(mockDbRepo, mockDbTxManager, mockKafkaProducer, mockGrpcClient)

    // Проверяем, что поля структуры инициализированы правильно
    require.NotNil(t, authService)
    require.Equal(t, mockDbRepo, authService.Dbrepo)
    require.Equal(t, mockDbTxManager, authService.Dbtxmanager)
    require.Equal(t, mockKafkaProducer, authService.KafkaProducer)
    require.Equal(t, mockGrpcClient, authService.GrpcClient)

    // Проверяем, что Validator инициализирован
    require.NotNil(t, authService.Validator)
}*/
