package repository

import (
	"context"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/logger"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AuthRedis struct {
	Client *redis.Client
	logger *logger.SessionLogger
}

func validateContext(ctx context.Context, logger *logger.SessionLogger) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Error("Metadata not found in context", zap.Error(erro.ErrorMissingMetadata))
		return "", erro.ErrorMissingMetadata
	}
	requestIDs := md.Get("requestID")
	if len(requestIDs) == 0 || requestIDs[0] == "" {
		logger.Error("Request ID not found in metadata", zap.Error(erro.ErrorRequiredRequestID))
		return "", erro.ErrorRequiredRequestID
	}
	requestID := requestIDs[0]

	if ctx.Err() != nil {
		logger.Error("Context cancelled",
			zap.String("requestID", requestID),
			zap.Error(ctx.Err()),
		)
		return "", erro.ErrorContextTimeOut
	}

	return requestID, nil
}
func (redisrepo *AuthRedis) SetSession(ctx context.Context, session model.Session) *RepositoryResponse {
	requestID, err := validateContext(ctx, redisrepo.logger)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: err}
	}
	err = redisrepo.Client.HSet(ctx, session.SessionID, map[string]interface{}{
		"UserID":         session.UserID,
		"ExpirationTime": session.ExpirationTime.Format(time.RFC3339),
	}).Err()

	if err != nil {
		redisrepo.logger.Error("Hset session Error",
			zap.String("requestID", requestID),
			zap.Error(err),
		)
		return &RepositoryResponse{Success: false, Errors: erro.ErrorSetSession}
	}
	expiration := time.Until(session.ExpirationTime)
	err = redisrepo.Client.Expire(ctx, session.SessionID, expiration).Err()
	if err != nil {
		redisrepo.logger.Error("Expire session error",
			zap.String("requestID", requestID),
			zap.Error(err),
		)
		return &RepositoryResponse{Success: false, Errors: erro.ErrorSetSession}
	}
	redisrepo.logger.Info("Successful session installation")
	return &RepositoryResponse{Success: true, SessionId: session.SessionID, ExpirationTime: session.ExpirationTime}
}

func (redisrepo *AuthRedis) GetSession(ctx context.Context, sessionID string) *RepositoryResponse {
	requestID, err := validateContext(ctx, redisrepo.logger)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: err}
	}
	result, err := redisrepo.Client.HGetAll(ctx, sessionID).Result()
	if err != nil {
		redisrepo.logger.Error("HGetAll session Error",
			zap.String("requestID", requestID),
			zap.Error(err),
		)
		return &RepositoryResponse{Success: false, Errors: erro.ErrorGetSession}
	}

	if len(result) == 0 {
		redisrepo.logger.Error("HGetAll session Error",
			zap.String("requestID", requestID),
			zap.Error(erro.ErrorInvalidSessionID),
		)
		return &RepositoryResponse{Success: false, Errors: erro.ErrorInvalidSessionID}
	}

	userIDString, ok := result["UserID"]
	if !ok {
		redisrepo.logger.Error("Get UserID from session Error",
			zap.String("requestID", requestID),
			zap.Error(erro.ErrorGetUserIdSession),
		)
		return &RepositoryResponse{Success: false, Errors: erro.ErrorGetUserIdSession}
	}

	expirationTimeString, ok := result["ExpirationTime"]
	if !ok {
		redisrepo.logger.Error("Get ExpirationTime from session Error",
			zap.String("requestID", requestID),
			zap.Error(erro.ErrorGetExpirationTimeSession),
		)
		return &RepositoryResponse{Success: false, Errors: erro.ErrorGetExpirationTimeSession}
	}

	expirationTime, err := time.Parse(time.RFC3339, expirationTimeString)
	if err != nil {
		redisrepo.logger.Error("Time-parse Error",
			zap.String("requestID", requestID),
			zap.Error(err),
		)
		return &RepositoryResponse{Success: false, Errors: erro.ErrorSessionParse}
	}

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		redisrepo.logger.Error("UUID-parse Error",
			zap.String("requestID", requestID),
			zap.Error(err),
		)
		return &RepositoryResponse{Success: false, Errors: erro.ErrorSessionParse}
	}
	redisrepo.logger.Info("Successful session receiving")
	return &RepositoryResponse{Success: true, UserID: userID.String(), ExpirationTime: expirationTime}
}
func (redisrepo *AuthRedis) DeleteSession(ctx context.Context, sessionID string) *RepositoryResponse {
	requestID, err := validateContext(ctx, redisrepo.logger)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: err}
	}
	err = redisrepo.Client.Del(ctx, sessionID).Err()
	if err != nil {
		redisrepo.logger.Error("Del session Error",
			zap.String("requestID", requestID),
			zap.Error(err),
		)
		return &RepositoryResponse{Success: false, Errors: erro.ErrorInternalServer}
	}
	redisrepo.logger.Info("Successful session deleted")
	return &RepositoryResponse{Success: true}
}
func NewAuthRedis(client *redis.Client, log *logger.SessionLogger) *AuthRedis {
	return &AuthRedis{Client: client, logger: log}
}
