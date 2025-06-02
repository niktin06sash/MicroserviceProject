package repository

//go:generate mockgen -source=repository.go -destination=mocks/mock.go
const GetProfileCache = "Repository-GetProfileCache"
const DeleteProfileCache = "Repository-DeleteProfileCache"
const AddProfileCache = "Repository-AddProfileCache"
const CreateUser = "Repository-CreateUser"
const AuthenticateUser = "Repository-AuthenticateUser"
const DeleteUser = "Repository-DeleteUser"
const UpdateName = "Repository-UpdateName"
const UpdatePassword = "Repository-UpdatePassword"
const UpdateEmail = "Repository-UpdateEmail"
const GetMyProfile = "Repository-GetMyProfile"
const GetProfileById = "Repository-GetProfileById"
const UpdateUserData = "Repository-UpdateUserData"
const (
	KeyFriendID         = "friendid"
	KeyUserID           = "userid"
	KeyUserEmail        = "useremail"
	KeyUserName         = "username"
	KeyUserPassword     = "userpassword"
	KeyUserTable        = "users"
	KeyFriendshipsTable = "friendships"
	KeyFriends          = "friends"
)

type RepositoryResponse struct {
	Success        bool
	SuccessMessage string
	Place          string
	Data           map[string]any
	Errors         *ErrorResponse
}
type ErrorResponse struct {
	Message string
	Type    string
}
