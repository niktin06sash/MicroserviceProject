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
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
	mock_service "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateAccount_Success(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUserId := uuid.MustParse("123e4567-e89b-12d3-a456-426614171000")
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
		name           string
		req            *model.UpdateRequest
		updateType     string
		place          string
		successmessage string
		args           []interface{}
	}{
		{
			name: "Successful update Name",
			req: &model.UpdateRequest{
				Name: "tester",
			},
			updateType:     "name",
			place:          repository.UpdateName,
			successmessage: "Successful update username in database",
			args:           []interface{}{"tester"},
		},
		{
			name: "Successful update Password",
			req: &model.UpdateRequest{
				LastPassword: "last_password",
				NewPassword:  "new_password",
			},
			updateType:     "password",
			place:          repository.UpdatePassword,
			successmessage: "Successful update userpassword in database",
			args:           []interface{}{"last_password", "new_password"},
		},
		{
			name: "Successful update Email",
			req: &model.UpdateRequest{
				Email:        "new@gmail.com",
				LastPassword: "last_password",
			},
			updateType:     "email",
			place:          repository.UpdateEmail,
			successmessage: "Successful update useremail in database",
			args:           []interface{}{"new@gmail.com", "last_password"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &sql.Tx{}
			mockTransactionRepo.EXPECT().BeginTx(
				mock.MatchedBy(func(ctx context.Context) bool {
					traceID := ctx.Value("traceID")
					return traceID != nil && traceID.(string) == fixedTraceID
				}),
			).Return(tx, nil)
			mockUserRepo.EXPECT().UpdateUserData(
				mock.MatchedBy(func(ctx context.Context) bool {
					traceID := ctx.Value("traceID")
					return traceID != nil && traceID.(string) == fixedTraceID
				}),
				tx,
				fixedUserId,
				tt.updateType,
				tt.args...,
			).Return(&repository.RepositoryResponse{
				Success:        true,
				Place:          tt.place,
				SuccessMessage: tt.successmessage,
			})
			mockCacheRepo.EXPECT().DeleteProfileCache(mock.MatchedBy(func(ctx context.Context) bool {
				traceID := ctx.Value("traceID")
				return traceID != nil && traceID.(string) == fixedTraceID
			}),
				fixedUserId.String(),
			).Return(&repository.RepositoryResponse{Success: true, Place: repository.DeleteProfileCache, SuccessMessage: "Successful delete profile from cache"})
			mockLogProducer.EXPECT().NewUserLog(kafka.LogLevelInfo, gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
			mockTransactionRepo.EXPECT().CommitTx(ctx, tx).Return(nil)
			response := as.UpdateAccount(ctx, tt.req, fixedUserId.String(), tt.updateType)
			require.True(t, response.Success)
		})
	}
}
func TestUpdateAccount_ValidationError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUserId := uuid.MustParse("123e4567-e89b-12d3-a456-426614171000")
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
		req           *model.UpdateRequest
		expectedError *erro.CustomError
		updateType    string
	}{
		{
			name: "Too Short Name",
			req: &model.UpdateRequest{
				Name: "te",
			},
			updateType:    "name",
			expectedError: &erro.CustomError{Type: erro.ClientErrorType, Message: "Name is too short"},
		},
		{
			name: "Too Short New Password",
			req: &model.UpdateRequest{
				LastPassword: "last_password",
				NewPassword:  "rd",
			},
			updateType:    "password",
			expectedError: &erro.CustomError{Type: erro.ClientErrorType, Message: "NewPassword is too short"},
		},
		{
			name: "Invalid Email Format",
			req: &model.UpdateRequest{
				Email:        "newgmail.com",
				LastPassword: "last_password",
			},
			updateType:    "email",
			expectedError: &erro.CustomError{Type: erro.ClientErrorType, Message: erro.ErrorNotEmailConst},
		},
		{
			name: "Too Short Last Password",
			req: &model.UpdateRequest{
				LastPassword: "l",
				NewPassword:  "new_password",
			},
			updateType:    "password",
			expectedError: &erro.CustomError{Type: erro.ClientErrorType, Message: "LastPassword is too short"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogProducer.EXPECT().NewUserLog(kafka.LogLevelWarn, gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
			response := as.UpdateAccount(ctx, tt.req, fixedUserId.String(), tt.updateType)
			require.False(t, response.Success)
			require.Equal(t, response.Errors, tt.expectedError)
		})
	}
}
func TestUpdateAccount_ParsingUserID_Error(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
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
	req := &model.UpdateRequest{
		Name: "tester",
	}
	mockLogProducer.EXPECT().NewUserLog(kafka.LogLevelWarn, gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.UpdateAccount(ctx, req, fixedUserId, "name")
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ClientErrorType, Message: erro.ErrorInvalidUserIDFormat})
}
func TestUpdateAccount_BeginTxError(t *testing.T) {
	fixedTraceUuid := "123e4567-e89b-12d3-a456-426614174000"
	fixedUserId := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
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
	req := &model.UpdateRequest{
		Email:        "new@gmail.com",
		LastPassword: "last_password",
	}
	mockTransactionRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, fmt.Errorf("database connection error"))
	mockLogProducer.EXPECT().NewUserLog(kafka.LogLevelError, gomock.Any(), fixedTraceUuid, gomock.Any()).AnyTimes()
	response := as.UpdateAccount(ctx, req, fixedUserId.String(), "email")
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}
func TestUpdateAccount_DataBaseError_ClientError(t *testing.T) {
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
	tx := &sql.Tx{}
	req := &model.UpdateRequest{
		NewPassword:  "new_password",
		LastPassword: "last_password",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		})).Return(tx, nil)
	mockUserRepo.EXPECT().UpdateUserData(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx,
		fixedUserId,
		"password",
		[]interface{}{"last_password", "new_password"},
	).Return(&repository.RepositoryResponse{
		Success: false,
		Place:   repository.UpdatePassword,
		Errors:  &erro.CustomError{Type: erro.ClientErrorType, Message: erro.ErrorIncorrectPassword},
	})
	mockTransactionRepo.EXPECT().RollbackTx(ctx, tx).Return(nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.UpdateAccount(ctx, req, fixedUserId.String(), "password")
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ClientErrorType, Message: erro.ErrorIncorrectPassword})
}
func TestUpdateAccount_DataBaseError_InternalServerError(t *testing.T) {
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
	tx := &sql.Tx{}
	req := &model.UpdateRequest{
		NewPassword:  "new_password",
		LastPassword: "last_password",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		})).Return(tx, nil)
	mockUserRepo.EXPECT().UpdateUserData(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx,
		fixedUserId,
		"password",
		[]interface{}{"last_password", "new_password"},
	).Return(&repository.RepositoryResponse{
		Success: false,
		Place:   repository.UpdatePassword,
		Errors:  &erro.CustomError{Type: erro.ServerErrorType, Message: erro.ErrorAfterReqUsers},
	})
	mockTransactionRepo.EXPECT().RollbackTx(ctx, tx).Return(nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.UpdateAccount(ctx, req, fixedUserId.String(), "password")
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}
func TestUpdateAccount_DeleteCacheError(t *testing.T) {
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
	tx := &sql.Tx{}
	req := &model.UpdateRequest{
		Name: "tester",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		})).Return(tx, nil)
	mockUserRepo.EXPECT().UpdateUserData(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx,
		fixedUserId,
		"name",
		[]interface{}{"tester"},
	).Return(&repository.RepositoryResponse{
		Success:        true,
		Place:          repository.UpdateName,
		SuccessMessage: "Successful update username in database",
	})
	mockCacheRepo.EXPECT().DeleteProfileCache(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}),
		fixedUserId.String(),
	).Return(&repository.RepositoryResponse{Success: false, Place: repository.DeleteProfileCache, Errors: &erro.CustomError{Type: erro.ServerErrorType, Message: erro.ErrorDelProfiles}})
	mockTransactionRepo.EXPECT().RollbackTx(ctx, tx).Return(nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.UpdateAccount(ctx, req, fixedUserId.String(), "name")
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}
func TestUpdateAccount_CommitTransactionError(t *testing.T) {
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
	tx := &sql.Tx{}
	req := &model.UpdateRequest{
		Email:        "new@gmail.com",
		LastPassword: "last_password",
	}
	mockTransactionRepo.EXPECT().BeginTx(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		})).Return(tx, nil)
	mockUserRepo.EXPECT().UpdateUserData(
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceID
		}),
		tx,
		fixedUserId,
		"email",
		[]interface{}{"new@gmail.com", "last_password"},
	).Return(&repository.RepositoryResponse{
		Success:        true,
		Place:          repository.UpdateEmail,
		SuccessMessage: "Successful update useremail in database",
	})
	mockCacheRepo.EXPECT().DeleteProfileCache(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}),
		fixedUserId.String(),
	).Return(&repository.RepositoryResponse{Success: true, Place: repository.DeleteProfileCache, SuccessMessage: "Successful delete profile from cache"})
	mockTransactionRepo.EXPECT().CommitTx(ctx, tx).Return(fmt.Errorf("Failed to commit transaction after all attempts")).Times(3)
	mockTransactionRepo.EXPECT().RollbackTx(ctx, tx).Return(nil)
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	response := as.UpdateAccount(ctx, req, fixedUserId.String(), "email")
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}
