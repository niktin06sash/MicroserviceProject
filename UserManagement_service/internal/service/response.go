package service

//go:generate mockgen -source=service.go -destination=mocks/mock.go
const RegistrateAndLogin = "UseCase-RegistrateAndLogin"
const AuthenticateAndLogin = "UseCase-AuthenticateAndLogin"
const DeleteAccount = "UseCase-DeleteAccount"
const Logout = "UseCase-Logout"
const UpdateAccount = "UseCase-UpdateAccount"
const GetMyProfile = "UseCase-GetMyProfile"
const GetProfileById = "UseCase-GetProfileById"
const GetMyFriends = "UseCase-GetMyFriends"

type ServiceResponse struct {
	Success   bool
	Data      map[string]any
	Errors    map[string]string
	ErrorType string
}
