package service

import (
	"context"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"

	"github.com/google/uuid"
)

type SessionService struct {
	repo          SessionRepos
	kafkaProducer LogProducer
}

func NewSessionService(repo SessionRepos, kafka LogProducer) *SessionService {
	return &SessionService{repo: repo, kafkaProducer: kafka}
}
func (s *SessionService) CreateSession(ctx context.Context, userid string) *ServiceResponse {
	const place = UseCase_CreateSession
	traceID := ctx.Value("traceID").(string)
	createsessionmapa := make(map[string]string)
	if userid == "" {
		createsessionmapa[erro.ErrorMessage] = erro.UserIdRequired
		createsessionmapa[erro.ErrorType] = erro.ServerErrorType
		s.kafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, erro.UserIdRequired)
		return &ServiceResponse{Success: false, Errors: createsessionmapa}
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
	const place = UseCase_ValidateSession
	traceID := ctx.Value("traceID").(string)
	validatemapa := make(map[string]string)
	flag := ctx.Value("flagvalidate").(string)
	switch flag {
	case "true":
		if _, err := uuid.Parse(sessionid); err != nil {
			validatemapa[erro.ErrorMessage] = erro.SessionIdInvalid
			validatemapa[erro.ErrorType] = erro.ServerErrorType
			fmterr := fmt.Sprintf("UUID-Parse sessionID error: %v", err)
			s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, fmterr)
			return &ServiceResponse{Success: false, Errors: validatemapa}
		}
	case "false":
		if _, err := uuid.Parse(sessionid); err != nil {
			validatemapa[erro.ErrorMessage] = erro.ErrorInvalidSessionIdFormat
			validatemapa[erro.ErrorType] = erro.ServerErrorType
			s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Request for an unauthorized users with invalid sessionID format")
			return &ServiceResponse{Success: false, Errors: validatemapa}
		}
	}
	return s.requestToDB(ctx, func(ctx context.Context) *repository.RepositoryResponse { return s.repo.GetSession(ctx, sessionid) }, traceID, place)
}
func (s *SessionService) DeleteSession(ctx context.Context, sessionid string) *ServiceResponse {
	const place = UseCase_DeleteSession
	traceID := ctx.Value("traceID").(string)
	deletemapa := make(map[string]string)
	if _, err := uuid.Parse(sessionid); err != nil {
		deletemapa[erro.ErrorMessage] = erro.SessionIdInvalid
		deletemapa[erro.ErrorType] = erro.ServerErrorType
		fmterr := fmt.Sprintf("UUID-Parse sessionID error: %v", err)
		s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, fmterr)
		return &ServiceResponse{Success: false, Errors: deletemapa}
	}
	return s.requestToDB(ctx, func(ctx context.Context) *repository.RepositoryResponse { return s.repo.GetSession(ctx, sessionid) }, traceID, place)
}
func (s *SessionService) requestToDB(ctx context.Context, operation func(context.Context) *repository.RepositoryResponse, traceid string, place string) *ServiceResponse {
	response := operation(ctx)
	if response.Errors != nil {
		typ := response.Errors[erro.ErrorType]
		switch typ {
		case erro.ServerErrorType:
			s.kafkaProducer.NewSessionLog(kafka.LogLevelError, response.Place, traceid, response.Errors[erro.ErrorMessage])
			return &ServiceResponse{Success: false, Errors: response.Errors}
		case erro.ClientErrorType:
			s.kafkaProducer.NewSessionLog(kafka.LogLevelWarn, response.Place, traceid, response.Errors[erro.ErrorMessage])
			return &ServiceResponse{Success: false, Errors: response.Errors}
		}
	}
	s.kafkaProducer.NewSessionLog(kafka.LogLevelInfo, response.Place, traceid, response.SuccessMessage)
	switch place {
	case UseCase_CreateSession:
		exp64 := response.Data[repository.KeyExpiryTime].(time.Time).Unix()
		return &ServiceResponse{Success: true, SessionID: response.Data[repository.KeySessionId].(string), ExpirationTime: exp64}
	case UseCase_DeleteSession:
		return &ServiceResponse{Success: true}
	case UseCase_ValidateSession:
		return &ServiceResponse{Success: true, SessionID: response.Data[repository.KeyUserId].(string)}
	}
	return &ServiceResponse{Success: false}
}
