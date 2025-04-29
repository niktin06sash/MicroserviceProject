package service_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	mock_client "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/client/mocks"
	mock_repository "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository/mocks"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLogout_Success(t *testing.T) {
	fixedTraceUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid.String())
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	mockGrpc := mock_client.NewMockGrpcClientService(ctrl)
	as := &service.AuthService{
		Dbrepo:     mockRepo,
		GrpcClient: mockGrpc,
		Validator:  validator.New(),
	}
	fixedsessUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174100")
	mockGrpc.EXPECT().DeleteSession(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceUuid.String()
	}), mock.MatchedBy(func(sessionID string) bool {
		return sessionID == fixedsessUUID.String()
	})).Return(&pb.DeleteSessionResponse{
		Success: true,
	}, nil)
	response := as.Logout(ctx, fixedsessUUID.String())
	log.Printf("Response: %+v", response)
	require.True(t, response.Success)
}
