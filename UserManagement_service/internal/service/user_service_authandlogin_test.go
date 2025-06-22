package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
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

func TestAuthenticateAndLogin_Success(t *testing.T) {
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
	req := &model.AuthenticationRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	mockUserRepo.EXPECT().GetUser(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		req.Email,
		req.Password,
	).
		Return(&repository.RepositoryResponse{
			Success:        true,
			Data:           map[string]any{repository.KeyUserID: fixedUserId.String()},
			Place:          repository.GetUser,
			SuccessMessage: "Successful get user from database",
		})
	mockSessionClient.EXPECT().
		CreateSession(
			mock.MatchedBy(func(ctx context.Context) bool {
				traceID := ctx.Value("traceID")
				return traceID != nil && traceID.(string) == fixedTraceID
			}),
			fixedUserId.String(),
		).
		Return(&proto.CreateSessionResponse{
			SessionID:  fixedSessionId,
			ExpiryTime: time.Now().Add(1 * time.Hour).Unix(),
			Success:    true,
		}, nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.AuthenticateAndLogin(ctx, req)
	require.True(t, response.Success)
	require.Equal(t, fixedUserId.String(), response.Data[repository.KeyUserID])
	require.Equal(t, fixedSessionId, response.Data[service.KeySessionID])
}
func TestAuthenticateAndLogin_ValidationErrors(t *testing.T) {
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
		req           *model.AuthenticationRequest
		expectedError *erro.CustomError
		logproducer   *gomock.Call
	}{
		{
			name: "Invalid Email Format",
			req: &model.AuthenticationRequest{
				Email:    "testexample.com",
				Password: "password123",
			},
			expectedError: &erro.CustomError{Type: erro.ClientErrorType, Message: erro.ErrorNotEmailConst},
			logproducer:   mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes(),
		},
		{
			name: "Too Short Password",
			req: &model.AuthenticationRequest{
				Email:    "valid@example.com",
				Password: "pas3",
			},
			expectedError: &erro.CustomError{Type: erro.ClientErrorType, Message: "Password is too short"},
			logproducer:   mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes(),
		},
		{
			name: "Required Field",
			req: &model.AuthenticationRequest{
				Email:    "valid@example.com",
				Password: "",
			},
			expectedError: &erro.CustomError{Type: erro.ClientErrorType, Message: "Password is Null"},
			logproducer:   mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := as.AuthenticateAndLogin(ctx, tt.req)
			require.False(t, response.Success)
			require.Equal(t, tt.expectedError, response.Errors)
		})
	}
}
func TestAuthenticateAndLogin_DataBaseError_ClientError(t *testing.T) {
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
	req := &model.AuthenticationRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	mockUserRepo.EXPECT().GetUser(ctx, req.Email, req.Password).Return(
		&repository.RepositoryResponse{Success: false, Errors: &erro.CustomError{Type: erro.ClientErrorType, Message: erro.ErrorInvalidPasswordConst}, Place: repository.GetUser},
	)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.AuthenticateAndLogin(ctx, req)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ClientErrorType, Message: erro.ErrorInvalidPasswordConst})
}
func TestAuthenticateAndLogin_DataBaseError_InternalServerError(t *testing.T) {
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
	req := &model.AuthenticationRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	mockUserRepo.EXPECT().GetUser(ctx, req.Email, req.Password).Return(
		&repository.RepositoryResponse{Success: false, Errors: &erro.CustomError{Type: erro.ServerErrorType, Message: erro.ErrorAfterReqUsers}, Place: repository.GetUser},
	)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.AuthenticateAndLogin(ctx, req)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}
func TestAuthenticateAndLogin_RetryGrpc_InternalServerError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
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
	req := &model.AuthenticationRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	mockUserRepo.EXPECT().GetUser(ctx, req.Email, req.Password).Return(&repository.RepositoryResponse{
		Success:        true,
		Data:           map[string]any{repository.KeyUserID: fixedUserId.String()},
		Place:          repository.GetUser,
		SuccessMessage: "Successful get user from database",
	})
	mockLogProducer.EXPECT().NewUserLog(kafka.LogLevelInfo, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockSessionClient.EXPECT().
		CreateSession(mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
			fixedUserId.String()).
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
	response := as.AuthenticateAndLogin(ctx, req)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.SessionServiceUnavalaible})
}
