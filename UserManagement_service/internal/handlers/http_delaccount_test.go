package handlers

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
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
	mock_service "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service/mocks"
	"github.com/stretchr/testify/require"
)

func TestDeleteAccount(t *testing.T) {
	tests := []struct {
		testname             string
		method               string
		reqbody              string
		sessionID            string // Кука session_id
		userID               string // UserID из контекста
		mockservice          func(r *mock_service.MockUserAuthentication)
		expectedStatuscode   int
		expectedResponseData HTTPResponse
	}{
		{
			testname:  "Success",
			method:    http.MethodDelete,
			reqbody:   `"qwerty1234"`,                         // Пароль в теле запроса
			sessionID: "123e4567-e89b-12d3-a456-426614174000", // Кука session_id
			userID:    "123e4567-e89b-12d3-a456-426614174000", // UserID из контекста
			mockservice: func(r *mock_service.MockUserAuthentication) {
				r.EXPECT().
					DeleteAccount(gomock.Any(), "123e4567-e89b-12d3-a456-426614174000", uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), "qwerty1234").
					Return(&service.ServiceResponse{
						Success: true,
					})
			},
			expectedStatuscode: http.StatusOK,
			expectedResponseData: HTTPResponse{
				Success: true,
			},
		},
		{
			testname:           "InvalidMethod",
			method:             http.MethodGet,
			userID:             "123e4567-e89b-12d3-a456-426614174000",
			reqbody:            "",
			expectedStatuscode: http.StatusMethodNotAllowed,
			expectedResponseData: HTTPResponse{
				Success: false,
				Errors: map[string]string{
					"Method": erro.ErrorNotDelete.Error(),
				},
			},
		},
		{
			testname:           "MissingSessionCookie",
			method:             http.MethodDelete,
			userID:             "123e4567-e89b-12d3-a456-426614174000",
			reqbody:            `"qwerty1234"`,
			expectedStatuscode: http.StatusBadRequest,
			expectedResponseData: HTTPResponse{
				Success: false,
				Errors: map[string]string{
					"Cookie": erro.ErrorMissingCookie.Error(),
				},
			},
		},
		{
			testname:           "InvalidRequestBody",
			method:             http.MethodDelete,
			reqbody:            "",
			sessionID:          "123e4567-e89b-12d3-a456-426614174000",
			userID:             "123e4567-e89b-12d3-a456-426614174000",
			expectedStatuscode: http.StatusBadRequest,
			expectedResponseData: HTTPResponse{
				Success: false,
				Errors: map[string]string{
					"ReadAll": erro.ErrorReadAll.Error(),
				},
			},
		},
		{
			testname:  "InvalidPassword",
			method:    http.MethodDelete,
			reqbody:   `"wrongpassword"`,
			sessionID: "123e4567-e89b-12d3-a456-426614174000",
			userID:    "123e4567-e89b-12d3-a456-426614174000",
			mockservice: func(r *mock_service.MockUserAuthentication) {
				r.EXPECT().
					DeleteAccount(gomock.Any(), "123e4567-e89b-12d3-a456-426614174000", uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), "wrongpassword").
					Return(&service.ServiceResponse{
						Success: false,
						Errors:  map[string]error{"Password": erro.ErrorInvalidPassword},
					})
			},
			expectedStatuscode: http.StatusInternalServerError,
			expectedResponseData: HTTPResponse{
				Success: false,
				Errors: map[string]string{
					"Password": erro.ErrorInvalidPassword.Error(),
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
			handler := Handler{services}

			reqBody := bytes.NewBufferString(test.reqbody)
			req := httptest.NewRequest(test.method, "/delete", reqBody)
			req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), "userID", uuid.MustParse(test.userID))
			req = req.WithContext(ctx)

			if test.sessionID != "" {
				req.AddCookie(&http.Cookie{Name: "session_id", Value: test.sessionID})
			}

			recorder := httptest.NewRecorder()

			handler.DeleteAccount(recorder, req)

			assert.Equal(t, test.expectedStatuscode, recorder.Code)

			var response HTTPResponse
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
