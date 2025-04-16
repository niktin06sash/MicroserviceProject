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
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	mock_repository "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository/mocks"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegistrateAndLogin_Success(t *testing.T) {
	fixedReqUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	ctx := context.WithValue(context.Background(), "requestID", fixedReqUuid.String())
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
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
	fixedsessUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614171000")
	gomock.InOrder(
		mockTxManager.EXPECT().BeginTx(mock.MatchedBy(func(ctx context.Context) bool {
			requestID := ctx.Value("requestID")
			return requestID != nil && requestID.(string) == fixedReqUuid.String()
		})).Return(tx, nil),
		mockRepo.EXPECT().CreateUser(ctx, tx, mock.MatchedBy(func(user *model.Person) bool {
			requestID := ctx.Value("requestID")
			return requestID != nil && requestID.(string) == fixedReqUuid.String() && user.Name == "John Doe" && user.Email == "john.doe@example.com"
		})).Return(&repository.DBRepositoryResponse{
			Success: true,
			UserId:  fixedUUID,
		}),
		mockGrpc.EXPECT().CreateSession(mock.MatchedBy(func(ctx context.Context) bool {
			requestID := ctx.Value("requestID")
			return requestID != nil && requestID.(string) == fixedReqUuid.String()
		}), mock.MatchedBy(func(userID string) bool {
			parseuserid, err := uuid.Parse(userID)
			return err == nil && parseuserid == fixedUUID
		})).Return(&pb.CreateSessionResponse{
			Success:    true,
			SessionID:  fixedsessUUID.String(),
			ExpiryTime: time.Now().Add(1 * time.Minute).Unix(),
		}, nil),
		mockTxManager.EXPECT().CommitTx(tx).Return(nil),
		mockTxManager.EXPECT().RollbackTx(tx).AnyTimes(),
	)
	response := as.RegistrateAndLogin(ctx, user)
	log.Printf("Response: %+v", response)

	require.True(t, response.Success)
	require.Equal(t, fixedsessUUID.String(), response.SessionId)
	require.NotNil(t, response.ExpireSession)
	require.True(t, response.ExpireSession.After(time.Now().Add(-1*time.Second)))
	require.NotEqual(t, "password123", user.Password)
}
func TestRegistrateAndLogin_MissingReqId(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
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
	mockTxManager.EXPECT().BeginTx(gomock.Any()).Times(0)
	response := as.RegistrateAndLogin(ctx, user)
	log.Printf("Response: %+v", response)

	require.False(t, response.Success)
	require.EqualError(t, response.Errors["ContextError"], "Error missing request ID")
}
func TestRegistrateAndLogin_ValidError(t *testing.T) {
	tests := []struct {
		name          string
		user          *model.Person
		expectedError map[string]error
	}{
		{
			name: "Invalid email format",
			user: &model.Person{
				Name:     "John Doe",
				Email:    "john.doeexample.com",
				Password: "password123",
			},
			expectedError: map[string]error{
				"Email": fmt.Errorf("This email format is not supported"),
			},
		},
		{
			name: "Password too short",
			user: &model.Person{
				Name:     "John Doe",
				Email:    "john.doe@example.com",
				Password: "pass",
			},
			expectedError: map[string]error{
				"Password": fmt.Errorf("Password is too short"),
			},
		},
		{
			name: "Missing name",
			user: &model.Person{
				Name:     "",
				Email:    "john.doe@example.com",
				Password: "password123",
			},
			expectedError: map[string]error{
				"Name": fmt.Errorf("Name is Null"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fixedReqUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
			ctx := context.WithValue(context.Background(), "requestID", fixedReqUuid.String())
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

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

			response := as.RegistrateAndLogin(ctx, tt.user)
			log.Printf("Response: %+v", response)

			require.False(t, response.Success)
			for field, expectedErr := range tt.expectedError {
				require.Contains(t, response.Errors, field)
				require.EqualError(t, response.Errors[field], expectedErr.Error())
			}
		})
	}
}
func TestRegistrateAndLogin_BeginTxError(t *testing.T) {

	fixedReqUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctx := context.WithValue(context.Background(), "requestID", fixedReqUuid.String())
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user := &model.Person{
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Password: "password123",
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)

	as := &service.AuthService{
		Dbtxmanager: mockTxManager,
		Validator:   validator.New(),
	}

	mockTxManager.EXPECT().BeginTx(gomock.Any()).Return(nil, fmt.Errorf("database connection error"))

	response := as.RegistrateAndLogin(ctx, user)
	log.Printf("Response: %+v", response)

	require.False(t, response.Success)
	require.Contains(t, response.Errors, "TransactionError")
	require.EqualError(t, response.Errors["TransactionError"], "Transaction creation error")
}
func TestRegistrateAndLogin_ContextBeforeCreateUser(t *testing.T) {

	fixedReqUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ctx = context.WithValue(ctx, "requestID", fixedReqUuid.String())
	user := &model.Person{
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Password: "password123",
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTxManager := mock_repository.NewMockDBTransactionManager(ctrl)

	as := &service.AuthService{
		Dbtxmanager: mockTxManager,
		Validator:   validator.New(),
	}

	mockTxManager.EXPECT().BeginTx(gomock.Any()).Return(nil, nil)
	response := as.RegistrateAndLogin(ctx, user)
	log.Printf("Response: %+v", response)

	require.False(t, response.Success)
	require.Contains(t, response.Errors, "ContextError")
	require.EqualError(t, response.Errors["ContextError"], "The timeout context has expired")
}
func TestRegistrateAndLogin_CreateUserError(t *testing.T) {

	fixedReqUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, "requestID", fixedReqUuid.String())

	user := &model.Person{
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Password: "password123",
	}

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

	mockRepo.EXPECT().CreateUser(ctx, tx, gomock.Any()).DoAndReturn(func(ctx context.Context, tx *sql.Tx, user *model.Person) *repository.DBRepositoryResponse {
		return &repository.DBRepositoryResponse{
			Success: false,
			Errors:  erro.ErrorUniqueEmail,
		}
	})

	mockTxManager.EXPECT().RollbackTx(tx).AnyTimes()

	response := as.RegistrateAndLogin(ctx, user)
	log.Printf("Response: %+v", response)

	require.False(t, response.Success)
	require.Contains(t, response.Errors, "RegistrateError")
	require.EqualError(t, response.Errors["RegistrateError"], "This email has already been registered")
}
func TestRegistrateAndLogin_CreateSessionError(t *testing.T) {

	fixedReqUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, "requestID", fixedReqUuid.String())

	user := &model.Person{
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Password: "password123",
	}

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

	mockRepo.EXPECT().CreateUser(ctx, tx, gomock.Any()).DoAndReturn(func(ctx context.Context, tx *sql.Tx, user *model.Person) *repository.DBRepositoryResponse {
		return &repository.DBRepositoryResponse{
			Success: true,
			UserId:  fixedUUID,
		}
	})

	mockTxManager.EXPECT().RollbackTx(tx).AnyTimes()

	mockGrpcClient.EXPECT().CreateSession(ctx, fixedUUID.String()).DoAndReturn(func(ctx context.Context, userID string) (*pb.CreateSessionResponse, error) {
		return nil, fmt.Errorf("grpc error")
	})

	response := as.RegistrateAndLogin(ctx, user)
	log.Printf("Response: %+v", response)

	require.False(t, response.Success)
	require.Contains(t, response.Errors, "GrpcResponseError")
	require.EqualError(t, response.Errors["GrpcResponseError"], "Grpc's response error")
}
func TestRegistrateAndLogin_CommitTxError(t *testing.T) {

	fixedReqUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedsessUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, "requestID", fixedReqUuid.String())

	user := &model.Person{
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Password: "password123",
	}

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

	mockRepo.EXPECT().CreateUser(ctx, tx, gomock.Any()).DoAndReturn(func(ctx context.Context, tx *sql.Tx, user *model.Person) *repository.DBRepositoryResponse {
		return &repository.DBRepositoryResponse{
			Success: true,
			UserId:  fixedUUID,
		}
	})

	mockTxManager.EXPECT().CommitTx(tx).Return(fmt.Errorf("commit error"))

	mockTxManager.EXPECT().RollbackTx(tx).AnyTimes()

	mockGrpcClient.EXPECT().CreateSession(mock.MatchedBy(func(ctx context.Context) bool {
		requestID := ctx.Value("requestID")
		return requestID != nil && requestID.(string) == fixedReqUuid.String()
	}), mock.MatchedBy(func(userID string) bool {
		parseuserid, err := uuid.Parse(userID)
		return err == nil && parseuserid == fixedUUID
	})).Return(&pb.CreateSessionResponse{
		Success:    true,
		SessionID:  fixedsessUUID.String(),
		ExpiryTime: time.Now().Add(1 * time.Minute).Unix(),
	}, nil)

	mockGrpcClient.EXPECT().DeleteSession(ctx, fixedsessUUID.String()).Return(&pb.DeleteSessionResponse{
		Success: true,
	}, nil)

	response := as.RegistrateAndLogin(ctx, user)
	log.Printf("Response: %+v", response)

	require.False(t, response.Success)
	require.Contains(t, response.Errors, "TransactionError")
	require.EqualError(t, response.Errors["TransactionError"], erro.ErrorCommitTransaction.Error())
}

/*func TestRegistrateAndLogin_ContextBeforeCreateSession(t *testing.T) {
	fixedReqUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, "requestID", fixedReqUuid.String())

	user := &model.Person{
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Password: "password123",
	}

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

	mockRepo.EXPECT().CreateUser(ctx, tx, gomock.Any()).DoAndReturn(func(ctx context.Context, tx *sql.Tx, user *model.Person) *repository.DBRepositoryResponse {
		return &repository.DBRepositoryResponse{
			Success: true,
			UserId:  fixedUUID,
		}
	})
	mockTxManager.EXPECT().RollbackTx(tx).AnyTimes()

	go func() {
		time.Sleep(91 * time.Millisecond)
		cancel()
	}()

	response := as.RegistrateAndLogin(ctx, user)
	log.Printf("Response: %+v", response)

	require.False(t, response.Success)
	require.Contains(t, response.Errors, "ContextError")
	require.EqualError(t, response.Errors["ContextError"], "The timeout context has expired")
}*/
