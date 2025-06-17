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
		return &ServiceResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.SessionServiceUnavalaible}}
	}
	sessionID := uuid.New()
	expiryTime := time.Now().Add(24 * time.Hour)
	newsession := model.Session{
		SessionID:      sessionID.String(),
		UserID:         userid,
		ExpirationTime: expiryTime,
	}
	bdresponse, serviceresponse := s.requestToDB(s.repo.SetSession(ctx, newsession), traceID)
	if serviceresponse != nil {
		return serviceresponse
	}
	exp64 := bdresponse.Data[repository.KeyExpiryTime].(time.Time).Unix()
	return &ServiceResponse{Success: bdresponse.Success, Data: Data{SessionID: bdresponse.Data[repository.KeySessionId].(string), ExpirationTime: exp64}}
}
func (s *SessionService) ValidateSession(ctx context.Context, sessionid string) *ServiceResponse {
	const place = ValidateSession
	traceID := ctx.Value("traceID").(string)
	flag := ctx.Value("flagvalidate").(string)
	switch flag {
	case "true":
		if _, err := uuid.Parse(sessionid); err != nil {
			fmterr := fmt.Sprintf("UUID-Parse sessionID error: %v", err)
			s.logproducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, fmterr)
			return &ServiceResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.SessionIdInvalid}}
		}
	case "false":
		if _, err := uuid.Parse(sessionid); err != nil {
			s.logproducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Request for an unauthorized users with invalid sessionID format")
			return &ServiceResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.SessionIdInvalid}}
		}
	}
	bdresponse, serviceresponse := s.requestToDB(s.repo.GetSession(ctx, sessionid), traceID)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: bdresponse.Success, Data: Data{UserID: bdresponse.Data[repository.KeyUserId].(string)}}
}
func (s *SessionService) DeleteSession(ctx context.Context, sessionid string) *ServiceResponse {
	const place = DeleteSession
	traceID := ctx.Value("traceID").(string)
	if _, err := uuid.Parse(sessionid); err != nil {
		fmterr := fmt.Sprintf("UUID-Parse sessionID error: %v", err)
		s.logproducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, fmterr)
		return &ServiceResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.SessionIdInvalid}}
	}
	bdresponse, serviceresponse := s.requestToDB(s.repo.DeleteSession(ctx, sessionid), traceID)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: bdresponse.Success}
}
func (use *SessionService) requestToDB(response *repository.RepositoryResponse, traceid string) (*repository.RepositoryResponse, *ServiceResponse) {
	if !response.Success && response.Errors != nil {
		switch response.Errors[erro.ErrorType] {
		case erro.ServerErrorType:
			use.logproducer.NewSessionLog(kafka.LogLevelError, response.Place, traceid, response.Errors[erro.ErrorMessage])
			response.Errors[erro.ErrorMessage] = erro.SessionServiceUnavalaible
			return response, &ServiceResponse{Success: false, Errors: response.Errors}

		case erro.ClientErrorType:
			use.logproducer.NewSessionLog(kafka.LogLevelWarn, response.Place, traceid, response.Errors[erro.ErrorMessage])
			return response, &ServiceResponse{Success: false, Errors: response.Errors}
		}
	}
	use.logproducer.NewSessionLog(kafka.LogLevelInfo, response.Place, traceid, response.SuccessMessage)
	return response, nil
}
