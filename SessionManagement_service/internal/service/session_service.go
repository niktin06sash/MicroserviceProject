package service

import (
	"context"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"

	"github.com/google/uuid"
)

type SessionServiceImplementation struct {
	Repo        SessionRepos
	Logproducer LogProducer
}

func NewSessionService(repo SessionRepos, logproducer LogProducer) *SessionServiceImplementation {
	return &SessionServiceImplementation{Repo: repo, Logproducer: logproducer}
}
func (s *SessionServiceImplementation) CreateSession(ctx context.Context, userid string) *ServiceResponse {
	const place = CreateSession
	traceID := ctx.Value("traceID").(string)
	if userid == "" {
		s.Logproducer.NewSessionLog(kafka.LogLevelError, place, traceID, erro.UserIdRequired)
		return &ServiceResponse{Success: false, Errors: erro.ServerError(erro.SessionServiceUnavalaible)}
	}
	newsession := &model.Session{
		SessionID:      uuid.New().String(),
		UserID:         userid,
		ExpirationTime: time.Now().Add(24 * time.Hour),
	}
	bdresponse, serviceresponse := s.requestToRepository(s.Repo.SetSession(ctx, newsession), traceID)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: bdresponse.Success, Data: Data{SessionID: newsession.SessionID, ExpirationTime: newsession.ExpirationTime.Unix()}}
}
func (s *SessionServiceImplementation) ValidateSession(ctx context.Context, sessionid string, flag string) *ServiceResponse {
	const place = ValidateSession
	traceID := ctx.Value("traceID").(string)
	if _, err := uuid.Parse(sessionid); err != nil {
		fmterr := fmt.Sprintf("UUID-Parse sessionID error: %v", err)
		s.Logproducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, fmterr)
		return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.SessionIdInvalid)}
	}
	bdresponse, serviceresponse := s.requestToRepository(s.Repo.GetSession(ctx, sessionid, flag), traceID)
	if serviceresponse != nil {
		return serviceresponse
	}
	if flag == "false" {
		return &ServiceResponse{Success: bdresponse.Success}
	}
	return &ServiceResponse{Success: bdresponse.Success, Data: Data{UserID: bdresponse.Data.UserID}}
}
func (s *SessionServiceImplementation) DeleteSession(ctx context.Context, sessionid string) *ServiceResponse {
	const place = DeleteSession
	traceID := ctx.Value("traceID").(string)
	if _, err := uuid.Parse(sessionid); err != nil {
		fmterr := fmt.Sprintf("UUID-Parse sessionID error: %v", err)
		s.Logproducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, fmterr)
		return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.SessionIdInvalid)}
	}
	bdresponse, serviceresponse := s.requestToRepository(s.Repo.DeleteSession(ctx, sessionid), traceID)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: bdresponse.Success}
}
func (s *SessionServiceImplementation) requestToRepository(response *repository.RepositoryResponse, traceid string) (*repository.RepositoryResponse, *ServiceResponse) {
	if !response.Success && response.Errors != nil {
		switch response.Errors.Type {
		case erro.ServerErrorType:
			s.Logproducer.NewSessionLog(kafka.LogLevelError, response.Place, traceid, response.Errors.Message)
			response.Errors.Message = erro.SessionServiceUnavalaible
			return response, &ServiceResponse{Success: false, Errors: response.Errors}

		case erro.ClientErrorType:
			s.Logproducer.NewSessionLog(kafka.LogLevelWarn, response.Place, traceid, response.Errors.Message)
			return response, &ServiceResponse{Success: false, Errors: response.Errors}
		}
	}
	s.Logproducer.NewSessionLog(kafka.LogLevelInfo, response.Place, traceid, response.SuccessMessage)
	return response, nil
}
