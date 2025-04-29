package handlers_test

import (
	"bytes"
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

func TestDeleteAccount_MissingUserID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/delete", nil)
	ctx := context.WithValue(req.Context(), "traceID", uuid.New().String())
	ctx = context.WithValue(ctx, "sessionID", uuid.New().String())
	req = req.WithContext(ctx)
	handler := handlers.Handler{}

	recorder := httptest.NewRecorder()

	handler.DeleteAccount(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var response response.HTTPResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, false, response.Success)
	assert.Equal(t, map[string]string{"InternalServerError": erro.ErrorMissingUserID.Error()}, response.Errors)
}
func TestDeleteAccount_MissingSessionID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/delete", nil)
	ctx := context.WithValue(req.Context(), "traceID", uuid.New().String())
	req = req.WithContext(ctx)
	handler := handlers.Handler{}

	recorder := httptest.NewRecorder()

	handler.DeleteAccount(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var response response.HTTPResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, false, response.Success)
	assert.Equal(t, map[string]string{"InternalServerError": erro.ErrorMissingSessionID.Error()}, response.Errors)
}
func TestDeleteAccount(t *testing.T) {
	tests := []struct {
		testname             string
		method               string
		reqbody              string
		sessionID            string
		userID               string
		mockservice          func(r *mock_service.MockUserAuthentication)
		expectedStatuscode   int
		expectedResponseData response.HTTPResponse
	}{
		{
			testname:  "Success",
			method:    http.MethodDelete,
			reqbody:   `{"password": "qwerty1234"}`,
			sessionID: "123e4567-e89b-12d3-a456-426614174000",
			userID:    "123e4567-e89b-12d3-a456-426614174000",
			mockservice: func(r *mock_service.MockUserAuthentication) {
				r.EXPECT().
					DeleteAccount(gomock.Any(), "123e4567-e89b-12d3-a456-426614174000", uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), "qwerty1234").
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
			reqbody:            "",
			expectedStatuscode: http.StatusBadRequest,
			expectedResponseData: response.HTTPResponse{
				Success: false,
				Errors: map[string]string{
					"ClientError": erro.ErrorNotDelete.Error(),
				},
			},
		},
		{
			testname:           "UnmarshalError",
			method:             http.MethodDelete,
			reqbody:            `{"password": 1234}`,
			sessionID:          "123e4567-e89b-12d3-a456-426614174000",
			userID:             "123e4567-e89b-12d3-a456-426614174000",
			expectedStatuscode: http.StatusBadRequest,
			expectedResponseData: response.HTTPResponse{
				Success: false,
				Errors: map[string]string{
					"ClientError": erro.ErrorUnmarshal.Error(),
				},
			},
		},
		{
			testname:           "UnmarshalErrorRequired",
			method:             http.MethodDelete,
			reqbody:            `{"password": ""}`,
			sessionID:          "123e4567-e89b-12d3-a456-426614174000",
			userID:             "123e4567-e89b-12d3-a456-426614174000",
			expectedStatuscode: http.StatusBadRequest,
			expectedResponseData: response.HTTPResponse{
				Success: false,
				Errors: map[string]string{
					"ClientError": erro.ErrorUnmarshal.Error(),
				},
			},
		},
		{
			testname:  "InvalidPassword",
			method:    http.MethodDelete,
			reqbody:   `{"password": "wrongpassword"}`,
			sessionID: "123e4567-e89b-12d3-a456-426614174000",
			userID:    "123e4567-e89b-12d3-a456-426614174000",
			mockservice: func(r *mock_service.MockUserAuthentication) {
				r.EXPECT().
					DeleteAccount(gomock.Any(), "123e4567-e89b-12d3-a456-426614174000", uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), "wrongpassword").
					Return(&service.ServiceResponse{
						Success: false,
						Errors:  map[string]error{"ClientError": erro.ErrorInvalidPassword},
						Type:    erro.ClientErrorType,
					})
			},
			expectedStatuscode: http.StatusBadRequest,
			expectedResponseData: response.HTTPResponse{
				Success: false,
				Errors: map[string]string{
					"ClientError": erro.ErrorInvalidPassword.Error(),
				},
			},
		},
		{
			testname:  "InternalServerError",
			method:    http.MethodDelete,
			reqbody:   `{"password": "qwerty1234"}`,
			sessionID: "123e4567-e89b-12d3-a456-426614174000",
			userID:    "123e4567-e89b-12d3-a456-426614174000",
			mockservice: func(r *mock_service.MockUserAuthentication) {
				r.EXPECT().
					DeleteAccount(gomock.Any(), "123e4567-e89b-12d3-a456-426614174000", uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), "qwerty1234").
					Return(&service.ServiceResponse{
						Success: false,
						Errors:  map[string]error{"InternalServerError": erro.ErrorStartTransaction},
						Type:    erro.ServerErrorType,
					})
			},
			expectedStatuscode: http.StatusInternalServerError,
			expectedResponseData: response.HTTPResponse{
				Success: true,
				Errors: map[string]string{
					"InternalServerError": erro.ErrorStartTransaction.Error(),
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

			reqBody := bytes.NewBufferString(test.reqbody)
			req := httptest.NewRequest(test.method, "/delete", reqBody)
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), "traceID", uuid.New().String())
			ctx = context.WithValue(ctx, "userID", test.userID)
			ctx = context.WithValue(ctx, "sessionID", test.sessionID)
			req = req.WithContext(ctx)

			recorder := httptest.NewRecorder()

			handler.DeleteAccount(recorder, req)

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
