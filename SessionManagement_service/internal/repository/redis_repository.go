package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"
	"github.com/redis/go-redis/v9"
)

type SessionDatabase struct {
	databaseclient *DatabaseObject
}

func NewSessionRepos(client *DatabaseObject) *SessionDatabase {
	return &SessionDatabase{databaseclient: client}
}

func (redisrepo *SessionDatabase) SetSession(ctx context.Context, session *model.Session) *RepositoryResponse {
	const place = SetSession
	jsondata, err := json.Marshal(session)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorMarshal, err)), Place: place}
	}
	err = redisrepo.databaseclient.connect.Set(ctx, session.SessionID, jsondata, 1*time.Hour).Err()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorSet, err)), Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successfull session installation", Place: place}
}

func (redisrepo *SessionDatabase) GetSession(ctx context.Context, sessionID string, flag string) *RepositoryResponse {
	const place = GetSession
	return redisrepo.getSessionData(ctx, sessionID, place, flag)
}
func (redisrepo *SessionDatabase) getSessionData(ctx context.Context, sessionID string, place string, flag string) *RepositoryResponse {
	result, err := redisrepo.databaseclient.connect.Get(ctx, sessionID).Result()
	if err != nil {
		if err == redis.Nil {
			if flag == "true" {
				return &RepositoryResponse{Success: false, Errors: erro.ClientError(erro.InvalidSession), Place: place}
			}
			return &RepositoryResponse{Success: true, SuccessMessage: "Request for an unauthorized page with invalid session", Place: place}
		}
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorGet, err)), Place: place}
	}
	if flag == "false" {
		return &RepositoryResponse{Success: false, Errors: erro.ClientError(erro.AlreadyAuthorized), Place: place}
	}
	var session model.Session
	err = json.Unmarshal([]byte(result), &session)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorUnmarshal, err)), Place: place}
	}
	return &RepositoryResponse{Success: true, Data: Data{UserID: session.UserID}, SuccessMessage: "Successfull get session", Place: place}
}
func (redisrepo *SessionDatabase) DeleteSession(ctx context.Context, sessionID string) *RepositoryResponse {
	const place = DeleteSession
	num, err := redisrepo.databaseclient.connect.Del(ctx, sessionID).Result()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorDelSession, err)), Place: place}
	}
	if num == 0 {
		return &RepositoryResponse{Success: false, Errors: erro.ClientError(erro.InvalidSession), Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successfull delete session", Place: place}
}
