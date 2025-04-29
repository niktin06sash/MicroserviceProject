package service_test

import (
	"context"
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

func TestAuthenticateAndLogin_Success(t *testing.T) {
	fixedTraceUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid.String())
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
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
		mock.MatchedBy(func(ctx context.Context) bool {
			traceID := ctx.Value("traceID")
			return traceID != nil && traceID.(string) == fixedTraceUuid.String()
		}),
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
	mockGrpc.EXPECT().CreateSession(mock.MatchedBy(func(ctx context.Context) bool {
		traceID := ctx.Value("traceID")
		return traceID != nil && traceID.(string) == fixedTraceUuid.String()
	}), mock.MatchedBy(func(userid string) bool {
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
}
func TestAuthenticateAndLogin_ValidError(t *testing.T) {
	tests := []struct {
		name              string
		user              *model.Person
		expectedError     map[string]error
		expectedTypeError erro.ErrorType
	}{
		{
			name: "Invalid email format",
			user: &model.Person{
				Email:    "john.doeexample.com",
				Password: "password123",
			},
			expectedError: map[string]error{
				"Email": fmt.Errorf("This email format is not supported"),
			},
			expectedTypeError: erro.ClientErrorType,
		},
		{
			name: "Password too short",
			user: &model.Person{
				Email:    "john.doe@example.com",
				Password: "pass",
			},
			expectedError: map[string]error{
				"Password": fmt.Errorf("Password is too short"),
			},
			expectedTypeError: erro.ClientErrorType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fixedTraceUuid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
			ctx := context.WithValue(context.Background(), "traceID", fixedTraceUuid.String())
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repository.NewMockDBAuthenticateRepos(ctrl)
			mockGrpc := mock_client.NewMockGrpcClientService(ctrl)

			as := &service.AuthService{
				Dbrepo: mockRepo,

				GrpcClient: mockGrpc,
				Validator:  validator.New(),
			}

			response := as.AuthenticateAndLogin(ctx, tt.user)
			log.Printf("Response: %+v", response)

			require.False(t, response.Success)
			require.Equal(t, tt.expectedTypeError, response.Type)
			for field, expectedErr := range tt.expectedError {
				require.Contains(t, response.Errors, field)
				require.EqualError(t, response.Errors[field], expectedErr.Error())
			}
		})
	}
}
