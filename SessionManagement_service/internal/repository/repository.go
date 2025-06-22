package repository

import "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"

//go:generate mockgen -source=repository.go -destination=mocks/mock.go
type RepositoryResponse struct {
	Success        bool
	SuccessMessage string
	Data           Data
	Errors         *erro.CustomError
	Place          string
}
type Data struct {
	UserID string
}

const KeySessionId = "sessionID"
const KeyUserId = "userID"
const KeyExpiryTime = "expirytime"
const SetSession = "Repository-SetSession"
const GetSession = "Repository-GetSession"
const DeleteSession = "Repository-DeleteSession"
