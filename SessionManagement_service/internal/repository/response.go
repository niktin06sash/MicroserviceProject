package repository

//go:generate mockgen -source=repository.go -destination=mocks/mock.go
type RepositoryResponse struct {
	Success        bool
	SuccessMessage string
	Data           map[string]any
	Errors         error
	Place          string
}

const KeySessionId = "sessionID"
const KeyUserId = "userID"
const ExpiryTime = "expirytime"
const SetSession = "Repository-SetSession"
const GetSession = "Repository-GetSession"
const DeleteSession = "Repository-DeleteSession"
