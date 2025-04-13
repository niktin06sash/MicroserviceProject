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
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	mock_repository "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository/mocks"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthenticateAndLogin_Success(t *testing.T) {
	ctx := context.Background()
	user := &model.Person{
		Email:    "john.doe@example.com",
		Password: "password123",
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
	mockGrpc := mock_client.NewMockGrpcClientService(ctrl)

	as := &service.AuthService{
		Dbrepo:     mockRepo,
		GrpcClient: mockGrpc,
		Validator:  validator.New(),
	}
	fixedUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fixedsessUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174100")
	mockRepo.EXPECT().GetUser(
		ctx,
		mock.MatchedBy(func(useremail string) bool {
			return useremail == "john.doe@example.com"
		}),
		mock.MatchedBy(func(password string) bool {
			return password == "password123"
		}),
	).Return(&repository.DBRepositoryResponse{
		Success: true,
		UserId:  fixedUUID,
	})
	mockGrpc.EXPECT().CreateSession(ctx, mock.MatchedBy(func(userid string) bool {
		parseuserid, err := uuid.Parse(userid)
		return err == nil && parseuserid == fixedUUID
	})).Return(&pb.CreateSessionResponse{
		Success:    true,
		SessionID:  fixedsessUUID.String(),
		ExpiryTime: time.Now().Add(1 * time.Minute).Unix(),
	}, nil)

	response := as.AuthenticateAndLogin(ctx, user)
	log.Printf("Response: %+v", response)

	require.True(t, response.Success)
	require.Equal(t, "123e4567-e89b-12d3-a456-426614174100", response.SessionId)
	require.NotNil(t, response.ExpireSession)
	require.True(t, response.ExpireSession.After(time.Now().Add(-1*time.Second)))
	require.Equal(t, "password123", user.Password)
}
