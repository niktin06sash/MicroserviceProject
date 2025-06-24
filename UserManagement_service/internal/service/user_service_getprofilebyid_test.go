package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
	mock_service "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetProfileById_SuccessCacheGet(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUserID := uuid.MustParse("123e4567-e89b-12d3-a456-426614173000")
	fixedEmail := "email@gmail.com"
	fixedName := "tester"
	fixedData := map[string]any{repository.KeyUserID: fixedUserID.String(), repository.KeyUserEmail: fixedEmail, repository.KeyUserName: fixedName}
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
	mockCacheRepo.EXPECT().GetProfileCache(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}),
		fixedUserID.String(),
	).Return(&repository.RepositoryResponse{Success: true, Place: repository.GetProfileCache, SuccessMessage: "Successful get profile from cache", Data: fixedData})
	mockLogProducer.EXPECT().NewUserLog(kafka.LogLevelInfo, gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.GetProfileById(ctx, fixedUserID.String())
	require.True(t, response.Success)
	require.Equal(t, response.Data, fixedData)
}
func TestGetProfileById_ParsingUserID_Error(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUserID := "123e4567-e89b-12d3-a456-42661417400"
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
	mockLogProducer.EXPECT().NewUserLog(kafka.LogLevelWarn, gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.GetProfileById(ctx, fixedUserID)
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ClientErrorType, Message: erro.ErrorInvalidUserIDFormat})
}
func TestGetProfileById_SuccessDatabaseGet(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUserId := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedEmail := "email@gmail.com"
	fixedName := "tester"
	fixedData := map[string]any{repository.KeyUserID: fixedUserId.String(), repository.KeyUserEmail: fixedEmail, repository.KeyUserName: fixedName}
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
	mockCacheRepo.EXPECT().GetProfileCache(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}),
		fixedUserId.String(),
	).Return(&repository.RepositoryResponse{Success: false, Place: repository.GetProfileCache, SuccessMessage: "Profile was not found in the cache"})
	mockLogProducer.EXPECT().NewUserLog(kafka.LogLevelInfo, gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	mockUserRepo.EXPECT().GetProfileById(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}), fixedUserId).Return(&repository.RepositoryResponse{Success: true, Place: repository.GetProfileById, SuccessMessage: "Successful get profile from database", Data: fixedData})
	mockCacheRepo.EXPECT().AddProfileCache(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}),
		fixedUserId.String(),
		fixedData,
	).Return(&repository.RepositoryResponse{Success: false, Place: repository.AddProfileCache, SuccessMessage: "Successful add profile in cache"})
	response := as.GetProfileById(ctx, fixedUserId.String())
	require.True(t, response.Success)
	require.Equal(t, response.Data, fixedData)
}
func TestGetProfileById_CacheGet_InternalServerError(t *testing.T) {
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
	mockCacheRepo.EXPECT().GetProfileCache(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}),
		fixedUserId.String(),
	).Return(&repository.RepositoryResponse{Success: false, Place: repository.GetProfileCache, Errors: &erro.CustomError{Type: erro.ServerErrorType, Message: erro.ErrorHgetAllProfiles}})
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.GetProfileById(ctx, fixedUserId.String())
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}
func TestGetProfileById_DataBase_InternalServerError(t *testing.T) {
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
	mockCacheRepo.EXPECT().GetProfileCache(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}),
		fixedUserId.String(),
	).Return(&repository.RepositoryResponse{Success: false, Place: repository.GetProfileCache, SuccessMessage: "Profile was not found in the cache"})
	mockUserRepo.EXPECT().GetProfileById(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}), fixedUserId).Return(&repository.RepositoryResponse{Success: false, Place: repository.GetProfileById, Errors: &erro.CustomError{Type: erro.ServerErrorType, Message: erro.ErrorAfterReqUsers}})
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.GetProfileById(ctx, fixedUserId.String())
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}
func TestGetProfileById_CacheAdd_InternalServerError(t *testing.T) {
	fixedTraceID := "123e4567-e89b-12d3-a456-426614174000"
	fixedUserId := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceID)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	fixedEmail := "email@gmail.com"
	fixedName := "tester"
	fixedData := map[string]any{repository.KeyUserID: fixedUserId.String(), repository.KeyUserEmail: fixedEmail, repository.KeyUserName: fixedName}
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
	mockCacheRepo.EXPECT().GetProfileCache(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}),
		fixedUserId.String(),
	).Return(&repository.RepositoryResponse{Success: false, Place: repository.GetProfileCache, SuccessMessage: "Profile was not found in the cache"})
	mockUserRepo.EXPECT().GetProfileById(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}), fixedUserId).Return(&repository.RepositoryResponse{Success: true, Place: repository.GetProfileById, SuccessMessage: "Successful get profile from database", Data: fixedData})
	mockCacheRepo.EXPECT().AddProfileCache(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceID
	}),
		fixedUserId.String(),
		fixedData,
	).Return(&repository.RepositoryResponse{Success: false, Place: repository.AddProfileCache, Errors: &erro.CustomError{Type: erro.ServerErrorType, Message: erro.ErrorHsetProfiles}})
	mockLogProducer.EXPECT().NewUserLog(gomock.Any(), gomock.Any(), fixedTraceID, gomock.Any()).AnyTimes()
	response := as.GetProfileById(ctx, fixedUserId.String())
	require.False(t, response.Success)
	require.Equal(t, response.Errors, &erro.CustomError{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible})
}
