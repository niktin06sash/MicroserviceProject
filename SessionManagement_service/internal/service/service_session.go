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
	return s.requestToDB(ctx, func(ctx context.Context) *repository.RepositoryResponse { return s.repo.SetSession(ctx, newsession) }, traceID, place)
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
	return s.requestToDB(ctx, func(ctx context.Context) *repository.RepositoryResponse { return s.repo.GetSession(ctx, sessionid) }, traceID, place)
}
func (s *SessionService) DeleteSession(ctx context.Context, sessionid string) *ServiceResponse {
	const place = DeleteSession
	traceID := ctx.Value("traceID").(string)
	if _, err := uuid.Parse(sessionid); err != nil {
		fmterr := fmt.Sprintf("UUID-Parse sessionID error: %v", err)
		s.logproducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, fmterr)
		return &ServiceResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.SessionIdInvalid}}
	}
	return s.requestToDB(ctx, func(ctx context.Context) *repository.RepositoryResponse { return s.repo.GetSession(ctx, sessionid) }, traceID, place)
}
func (s *SessionService) requestToDB(ctx context.Context, operation func(context.Context) *repository.RepositoryResponse, traceid string, place string) *ServiceResponse {
	response := operation(ctx)
	if response.Errors != nil {
		typ := response.Errors[erro.ErrorType]
		switch typ {
		case erro.ServerErrorType:
			s.logproducer.NewSessionLog(kafka.LogLevelError, response.Place, traceid, response.Errors[erro.ErrorMessage])
			response.Errors[erro.ErrorMessage] = erro.SessionServiceUnavalaible
			return &ServiceResponse{Success: false, Errors: response.Errors}
		case erro.ClientErrorType:
			s.logproducer.NewSessionLog(kafka.LogLevelWarn, response.Place, traceid, response.Errors[erro.ErrorMessage])
			return &ServiceResponse{Success: false, Errors: response.Errors}
		}
	}
	s.logproducer.NewSessionLog(kafka.LogLevelInfo, response.Place, traceid, response.SuccessMessage)
	switch place {
	case CreateSession:
		exp64 := response.Data[repository.KeyExpiryTime].(time.Time).Unix()
		return &ServiceResponse{Success: true, Data: Data{SessionID: response.Data[repository.KeySessionId].(string), ExpirationTime: exp64}}
	case DeleteSession:
		return &ServiceResponse{Success: true}
	case ValidateSession:
		return &ServiceResponse{Success: true, Data: Data{UserID: response.Data[repository.KeyUserId].(string)}}
	}
	return &ServiceResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.SessionServiceUnavalaible}}
}
