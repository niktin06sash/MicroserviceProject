package repository

import (
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock.go
const GetProfileCache = "Repository-GetProfileCache"
const DeleteProfileCache = "Repository-DeleteProfileCache"
const AddProfileCache = "Repository-AddProfileCache"
const CreateUser = "Repository-CreateUser"
const GetUser = "Repository-GetUser"
const DeleteUser = "Repository-DeleteUser"
const UpdateName = "Repository-UpdateName"
const UpdatePassword = "Repository-UpdatePassword"
const UpdateEmail = "Repository-UpdateEmail"
const GetMyProfile = "Repository-GetMyProfile"
const GetProfileById = "Repository-GetProfileById"
const UpdateUserData = "Repository-UpdateUserData"
const (
	KeyUserID    = "id"
	KeyUserTable = "users"
	KeyUser      = "user"
)

func BadRepositoryResponse(erro *erro.CustomError, place string) *RepositoryResponse {
	return &RepositoryResponse{Errors: erro, Success: false, Place: place}
}
func SuccessRepositoryResponse(data map[string]any, place string, succmessage string) *RepositoryResponse {
	return &RepositoryResponse{Data: data, Place: place, Success: true, SuccessMessage: succmessage}
}

type RepositoryResponse struct {
	Success        bool
	SuccessMessage string
	Place          string
	Data           map[string]any
	Errors         error
}
