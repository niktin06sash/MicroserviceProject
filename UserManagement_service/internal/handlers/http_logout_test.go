package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/assert"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
	mock_service "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service/mocks"
	"github.com/stretchr/testify/require"
)

func TestLogout_MissingUserID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/logout", nil)
	ctx := context.WithValue(req.Context(), "traceID", uuid.New().String())
	ctx = context.WithValue(ctx, "sessionID", uuid.New().String())
	req = req.WithContext(ctx)
	handler := handlers.Handler{}

	recorder := httptest.NewRecorder()

	handler.Logout(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var response response.HTTPResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, false, response.Success)
	assert.Equal(t, map[string]string{"InternalServerError": erro.ErrorMissingUserID.Error()}, response.Errors)
}
func TestLogout_MissingSessionID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/logout", nil)
	ctx := context.WithValue(req.Context(), "traceID", uuid.New().String())
	req = req.WithContext(ctx)
	handler := handlers.Handler{}

	recorder := httptest.NewRecorder()

	handler.Logout(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var response response.HTTPResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, false, response.Success)
	assert.Equal(t, map[string]string{"InternalServerError": erro.ErrorMissingSessionID.Error()}, response.Errors)
}
func TestLogout(t *testing.T) {
	tests := []struct {
		testname             string
		method               string
		sessionID            string
		userID               string
		mockservice          func(r *mock_service.MockUserAuthentication)
		expectedStatuscode   int
		expectedResponseData response.HTTPResponse
	}{
		{
			testname:  "Success",
			method:    http.MethodDelete,
			sessionID: "123e4567-e89b-12d3-a456-426614174000",
			userID:    "123e4567-e89b-12d3-a456-426614174000",
			mockservice: func(r *mock_service.MockUserAuthentication) {
				r.EXPECT().
					Logout(gomock.Any(), "123e4567-e89b-12d3-a456-426614174000").
					Return(&service.ServiceResponse{
						Success: true,
					})
			},
			expectedStatuscode: http.StatusOK,
			expectedResponseData: response.HTTPResponse{
				Success: true,
			},
		},
		{
			testname:           "InvalidMethod",
			method:             http.MethodGet,
			userID:             "123e4567-e89b-12d3-a456-426614174000",
			sessionID:          "123e4567-e89b-12d3-a456-426614174000",
			expectedStatuscode: http.StatusBadRequest,
			expectedResponseData: response.HTTPResponse{
				Success: false,
				Errors: map[string]string{
					"ClientError": erro.ErrorNotDelete.Error(),
				},
			},
		},
		{
			testname:  "InternalServerError",
			method:    http.MethodDelete,
			sessionID: "123e4567-e89b-12d3-a456-426614174000",
			userID:    "123e4567-e89b-12d3-a456-426614174000",
			mockservice: func(r *mock_service.MockUserAuthentication) {
				r.EXPECT().
					Logout(gomock.Any(), "123e4567-e89b-12d3-a456-426614174000").
					Return(&service.ServiceResponse{
						Success: false,
						Errors:  map[string]error{"InternalServerError": erro.ErrorContextTimeout},
						Type:    erro.ServerErrorType,
					})
			},
			expectedStatuscode: http.StatusInternalServerError,
			expectedResponseData: response.HTTPResponse{
				Success: false,
				Errors: map[string]string{
					"InternalServerError": erro.ErrorContextTimeout.Error(),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockUserAuthentication(ctrl)
			if test.mockservice != nil {
				test.mockservice(mockService)
			}

			services := &service.Service{UserAuthentication: mockService}
			handler := handlers.Handler{Services: services}
			req := httptest.NewRequest(test.method, "/logout", nil)
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), "traceID", uuid.New().String())
			ctx = context.WithValue(ctx, "userID", test.userID)
			ctx = context.WithValue(ctx, "sessionID", test.sessionID)
			req = req.WithContext(ctx)

			recorder := httptest.NewRecorder()

			handler.Logout(recorder, req)

			assert.Equal(t, test.expectedStatuscode, recorder.Code)

			var response response.HTTPResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)

			if test.expectedStatuscode == http.StatusOK {
				assert.Equal(t, test.expectedResponseData.Success, response.Success)
			} else {
				assert.Equal(t, test.expectedResponseData.Errors, response.Errors)
			}
		})
	}
}
