package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/assert"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
	mock_service "github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service/mocks"
	"github.com/stretchr/testify/require"
)

func TestRegistration(t *testing.T) {
	tests := []struct {
		testname             string
		method               string
		reqbody              string
		requser              model.Person
		mockservice          func(r *mock_service.MockUserAuthentication, user model.Person)
		expectedStatuscode   int
		expectedResponseData HTTPResponse
	}{
		{
			testname: "Success",
			method:   http.MethodPost,
			reqbody:  `{"name": "testname", "email": "testname@gmail.com", "password": "qwerty1234"}`,
			requser: model.Person{
				Name:     "testname",
				Email:    "testname@gmail.com",
				Password: "qwerty1234",
			},
			mockservice: func(r *mock_service.MockUserAuthentication, user model.Person) {
				r.EXPECT().
					RegistrateAndLogin(gomock.Any(), gomock.Any()).
					Return(&service.ServiceResponse{
						Success:       true,
						UserId:        uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
						SessionId:     "123e4567-e89b-12d3-a456-426614174000",
						ExpireSession: time.Now().Add(24 * time.Hour),
					})
			},
			expectedStatuscode: http.StatusOK,
			expectedResponseData: HTTPResponse{
				Success:       true,
				SessionId:     "123e4567-e89b-12d3-a456-426614174000",
				ExpireSession: time.Now().Add(24 * time.Hour).Truncate(time.Second),
			},
		},
		{
			testname:           "InvalidMethod",
			method:             http.MethodGet,
			reqbody:            "",
			expectedStatuscode: http.StatusMethodNotAllowed,
			expectedResponseData: HTTPResponse{
				Success: false,
				Errors: map[string]string{
					"Method": erro.ErrorNotPost.Error(),
				},
			},
		},
		{
			testname:           "UnmarshalError",
			method:             http.MethodPost,
			reqbody:            `{"name": "testname", "email": "testname@gmail.com", "password": 1234}`,
			expectedStatuscode: http.StatusInternalServerError,
			expectedResponseData: HTTPResponse{
				Success: false,
				Errors: map[string]string{
					"Unmarshal": erro.ErrorUnmarshal.Error(),
				},
			},
		},
		{
			testname: "InvalidData",
			method:   http.MethodPost,
			reqbody:  `{"name": "testname", "email": "testname@gmailcom", "password": "qwerty1234"}`,
			requser: model.Person{
				Name:     "testname",
				Email:    "testname@gmail.com",
				Password: "qwerty1234",
			},
			mockservice: func(r *mock_service.MockUserAuthentication, user model.Person) {
				r.EXPECT().
					RegistrateAndLogin(gomock.Any(), gomock.Any()).
					Return(&service.ServiceResponse{
						Success: false,
						Errors:  map[string]error{"Email": erro.ErrorNotEmail},
					})
			},
			expectedStatuscode: http.StatusBadRequest,
			expectedResponseData: HTTPResponse{
				Success: false,
				Errors: map[string]string{
					"Email": erro.ErrorNotEmail.Error(),
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
				test.mockservice(mockService, test.requser)
			}

			services := &service.Service{UserAuthentication: mockService}
			handler := Handler{services}

			reqBody := bytes.NewBufferString(test.reqbody)
			req := httptest.NewRequest(test.method, "/reg", reqBody)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			handler.Registration(recorder, req)

			assert.Equal(t, test.expectedStatuscode, recorder.Code)

			if test.expectedStatuscode == http.StatusOK {
				var response HTTPResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, test.expectedResponseData.Success, response.Success)
				assert.Equal(t, test.expectedResponseData.SessionId, response.SessionId)
				assert.Equal(t, test.expectedResponseData.ExpireSession.Truncate(time.Second), response.ExpireSession.Truncate(time.Second))
			} else {
				var response HTTPResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, test.expectedResponseData.Success, response.Success)
				assert.Equal(t, test.expectedResponseData.Errors, response.Errors)
			}
		})
	}
}
