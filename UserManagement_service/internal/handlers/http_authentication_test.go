package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/assert"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
	mock_service "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service/mocks"
	"github.com/stretchr/testify/require"
)

func TestAuthentication(t *testing.T) {
	tests := []struct {
		testname             string
		method               string
		reqbody              string
		mockservice          func(r *mock_service.MockUserAuthentication)
		expectedStatuscode   int
		expectedResponseData response.HTTPResponse
	}{
		{
			testname: "Success",
			method:   http.MethodPost,
			reqbody:  `{"email": "testname@gmail.com", "password": "qwerty1234"}`,
			mockservice: func(r *mock_service.MockUserAuthentication) {
				r.EXPECT().
					AuthenticateAndLogin(gomock.Any(), &model.Person{
						Email:    "testname@gmail.com",
						Password: "qwerty1234",
					}).
					Return(&service.ServiceResponse{
						Success:       true,
						UserId:        uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
						SessionId:     "123e4567-e89b-12d3-a456-426614174000",
						ExpireSession: time.Now().Add(24 * time.Hour),
					})
			},
			expectedStatuscode: http.StatusOK,
			expectedResponseData: response.HTTPResponse{
				Success: true,
				Data: map[string]any{
					"UserID": uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				},
			},
		},
		{
			testname:           "InvalidMethod",
			method:             http.MethodGet,
			reqbody:            "",
			expectedStatuscode: http.StatusBadRequest,
			expectedResponseData: response.HTTPResponse{
				Success: false,
				Errors: map[string]string{
					"ClientError": erro.ErrorNotPost.Error(),
				},
			},
		},
		{
			testname:           "UnmarshalError",
			method:             http.MethodPost,
			reqbody:            `{"email": "testname@gmail.com", "password": 1234}`,
			expectedStatuscode: http.StatusBadRequest,
			expectedResponseData: response.HTTPResponse{
				Success: false,
				Errors: map[string]string{
					"ClientError": erro.ErrorUnmarshal.Error(),
				},
			},
		},
		{
			testname: "InvalidData",
			method:   http.MethodPost,
			reqbody:  `{"email": "testname@gmailcom", "password": "qwerty1234"}`,
			mockservice: func(r *mock_service.MockUserAuthentication) {
				r.EXPECT().
					AuthenticateAndLogin(gomock.Any(), &model.Person{
						Email:    "testname@gmailcom",
						Password: "qwerty1234",
					}).
					Return(&service.ServiceResponse{
						Success: false,
						Errors: map[string]error{
							"Email": erro.ErrorNotEmail,
						},
						Type: erro.ClientErrorType,
					})
			},
			expectedStatuscode: http.StatusBadRequest,
			expectedResponseData: response.HTTPResponse{
				Success: false,
				Errors: map[string]string{
					"Email": erro.ErrorNotEmail.Error(),
				},
			},
		},
		{
			testname: "InvalidPassword",
			method:   http.MethodPost,
			reqbody:  `{"email": "testname@gmail.com", "password": "qwerty1234"}`,
			mockservice: func(r *mock_service.MockUserAuthentication) {
				r.EXPECT().
					AuthenticateAndLogin(gomock.Any(), &model.Person{
						Email:    "testname@gmail.com",
						Password: "qwerty1234",
					}).
					Return(&service.ServiceResponse{
						Success: false,
						Errors: map[string]error{
							"ClientError": erro.ErrorInvalidPassword,
						},
						Type: erro.ClientErrorType,
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
			testname: "InternalServerError",
			method:   http.MethodPost,
			reqbody:  `{"email": "testname@gmail.com", "password": "qwerty1234"}`,
			mockservice: func(r *mock_service.MockUserAuthentication) {
				r.EXPECT().
					AuthenticateAndLogin(gomock.Any(), &model.Person{
						Email:    "testname@gmail.com",
						Password: "qwerty1234",
					}).
					Return(&service.ServiceResponse{
						Success: false,
						Errors: map[string]error{
							"InternalServerError": erro.ErrorContextTimeout,
						},
						Type: erro.ServerErrorType,
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

			reqBody := bytes.NewBufferString(test.reqbody)
			req := httptest.NewRequest(test.method, "/auth", reqBody)
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), "traceID", uuid.New().String())
			req = req.WithContext(ctx)
			recorder := httptest.NewRecorder()

			handler.Authentication(recorder, req)

			assert.Equal(t, test.expectedStatuscode, recorder.Code)

			var response response.HTTPResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, test.expectedResponseData.Success, response.Success)

			if test.expectedStatuscode == http.StatusOK {
				assert.Equal(t, test.expectedResponseData.Data["UserID"].(uuid.UUID).String(), response.Data["UserID"])
			} else {
				assert.Equal(t, test.expectedResponseData.Errors, response.Errors)
			}
		})
	}
}
