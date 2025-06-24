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

type SessionService struct {
	repo        SessionRepos
	logproducer LogProducer
}

func NewSessionService(repo SessionRepos, logproducer LogProducer) *SessionService {
	return &SessionService{repo: repo, logproducer: logproducer}
}
func (s *SessionService) CreateSession(ctx context.Context, userid string) *ServiceResponse {
	const place = CreateSession
	traceID := ctx.Value("traceID").(string)
	if userid == "" {
		s.logproducer.NewSessionLog(kafka.LogLevelError, place, traceID, erro.UserIdRequired)
		return &ServiceResponse{Success: false, Errors: erro.ServerError(erro.SessionServiceUnavalaible)}
	}
	newsession := &model.Session{
		SessionID:      uuid.New().String(),
		UserID:         userid,
		ExpirationTime: time.Now().Add(24 * time.Hour),
	}
	bdresponse, serviceresponse := s.requestToDB(s.repo.SetSession(ctx, newsession), traceID)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: bdresponse.Success, Data: Data{SessionID: newsession.SessionID, ExpirationTime: newsession.ExpirationTime.Unix()}}
}
func (s *SessionService) ValidateSession(ctx context.Context, sessionid string) *ServiceResponse {
	const place = ValidateSession
	traceID := ctx.Value("traceID").(string)
	flag := ctx.Value("flagvalidate").(string)
	if _, err := uuid.Parse(sessionid); err != nil {
		fmterr := fmt.Sprintf("UUID-Parse sessionID error: %v", err)
		s.logproducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, fmterr)
		return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.SessionIdInvalid)}
	}
	bdresponse, serviceresponse := s.requestToDB(s.repo.GetSession(ctx, sessionid), traceID)
	if serviceresponse != nil {
		return serviceresponse
	}
	if flag == "false" {
		return &ServiceResponse{Success: bdresponse.Success}
	}
	return &ServiceResponse{Success: bdresponse.Success, Data: Data{UserID: bdresponse.Data.UserID}}
}
func (s *SessionService) DeleteSession(ctx context.Context, sessionid string) *ServiceResponse {
	const place = DeleteSession
	traceID := ctx.Value("traceID").(string)
	if _, err := uuid.Parse(sessionid); err != nil {
		fmterr := fmt.Sprintf("UUID-Parse sessionID error: %v", err)
		s.logproducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, fmterr)
		return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.SessionIdInvalid)}
	}
	bdresponse, serviceresponse := s.requestToDB(s.repo.DeleteSession(ctx, sessionid), traceID)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: bdresponse.Success}
}
func (use *SessionService) requestToDB(response *repository.RepositoryResponse, traceid string) (*repository.RepositoryResponse, *ServiceResponse) {
	if !response.Success && response.Errors != nil {
		switch response.Errors.Type {
		case erro.ServerErrorType:
			use.logproducer.NewSessionLog(kafka.LogLevelError, response.Place, traceid, response.Errors.Message)
			response.Errors.Message = erro.SessionServiceUnavalaible
			return response, &ServiceResponse{Success: false, Errors: response.Errors}

		case erro.ClientErrorType:
			use.logproducer.NewSessionLog(kafka.LogLevelWarn, response.Place, traceid, response.Errors.Message)
			return response, &ServiceResponse{Success: false, Errors: response.Errors}
		}
	}
	use.logproducer.NewSessionLog(kafka.LogLevelInfo, response.Place, traceid, response.SuccessMessage)
	return response, nil
}
