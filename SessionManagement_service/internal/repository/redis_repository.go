package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/logger"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AuthRedis struct {
	Client *RedisObject
	logger *logger.SessionLogger
}

func validateContext(ctx context.Context, logger *logger.SessionLogger, place string) (string, error) {
	traceID := ctx.Value("traceID").(string)
	select {
	case <-ctx.Done():
		logger.Error(fmt.Sprintf("[%s] Context time-out", place),
			zap.String("requestID", traceID),
			zap.Error(ctx.Err()),
		)
		return "", status.Errorf(codes.Internal, "request timed out")
	default:
		return traceID, nil
	}
}
func (redisrepo *AuthRedis) SetSession(ctx context.Context, session model.Session) *RepositoryResponse {
	traceID, err := validateContext(ctx, redisrepo.logger, "SetSession")
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: err}
	}
	err = redisrepo.Client.RedisClient.HSet(ctx, session.SessionID, map[string]interface{}{
		"UserID":         session.UserID,
		"ExpirationTime": session.ExpirationTime.Format(time.RFC3339),
	}).Err()
	if err != nil {
		redisrepo.logger.Error("SetSession: Hset session Error",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, "Hset session Error")}
	}
	expiration := time.Until(session.ExpirationTime)
	err = redisrepo.Client.RedisClient.Expire(ctx, session.SessionID, expiration).Err()
	if err != nil {
		redisrepo.logger.Error("SetSession: Expire session error",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, "Expire session Error")}
	}
	redisrepo.logger.Info("SetSession: Successful session installation",
		zap.String("traceID", traceID),
	)
	return &RepositoryResponse{Success: true, SessionId: session.SessionID, ExpirationTime: session.ExpirationTime}
}

func (redisrepo *AuthRedis) GetSession(ctx context.Context, sessionID string) *RepositoryResponse {
	traceID, err := validateContext(ctx, redisrepo.logger, "GetSession")
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: err}
	}
	result, err := redisrepo.Client.RedisClient.HGetAll(ctx, sessionID).Result()
	if err != nil {
		if err == redis.Nil {
			redisrepo.logger.Warn("GetSession: Session not found",
				zap.String("traceID", traceID),
				zap.String("sessionID", sessionID),
			)
			return &RepositoryResponse{
				Success: false,
				Errors:  status.Errorf(codes.InvalidArgument, "Session not found"),
			}
		}
		redisrepo.logger.Error("GetSession: HGetAll session Error",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, "HGetAll session Error")}
	}
	if len(result) == 0 {
		redisrepo.logger.Error("GetSession: Session is empty or invalid",
			zap.String("traceID", traceID),
			zap.Error(erro.ErrorInvalidSessionID),
		)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.InvalidArgument, "Session is empty or invalid")}
	}
	userIDString, ok := result["UserID"]
	if !ok {
		redisrepo.logger.Error("GetSession: Get UserID from session Error",
			zap.String("traceID", traceID),
			zap.Error(erro.ErrorGetUserIdSession),
		)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.InvalidArgument, "Get UserID from session Error")}
	}
	expirationTimeString, ok := result["ExpirationTime"]
	if !ok {
		redisrepo.logger.Error("GetSession: Get ExpirationTime from session Error",
			zap.String("traceID", traceID),
			zap.Error(erro.ErrorGetExpirationTimeSession),
		)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.InvalidArgument, "Get Expiration Time from session Error")}
	}
	expirationTime, err := time.Parse(time.RFC3339, expirationTimeString)
	if err != nil {
		redisrepo.logger.Error("GetSession: Time-parse Error",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, "Time-parse Error")}
	}
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		redisrepo.logger.Error("GetSession: UUID-parse Error",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, "UUID-parse Error")}
	}
	redisrepo.logger.Info("GetSession: Successful session receiving",
		zap.String("traceID", traceID),
	)
	return &RepositoryResponse{Success: true, UserID: userID.String(), ExpirationTime: expirationTime}
}
func (redisrepo *AuthRedis) DeleteSession(ctx context.Context, sessionID string) *RepositoryResponse {
	traceID, err := validateContext(ctx, redisrepo.logger, "DeleteSession")
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: err}
	}
	err = redisrepo.Client.RedisClient.Del(ctx, sessionID).Err()
	if err != nil {
		if err == redis.Nil {
			redisrepo.logger.Warn("DeleteSession: Session not found",
				zap.String("traceID", traceID),
				zap.String("sessionID", sessionID),
			)
			return &RepositoryResponse{
				Success: false,
				Errors:  status.Errorf(codes.InvalidArgument, "Session not found"),
			}
		}
		redisrepo.logger.Error("DeleteSession: Del session Error",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, "Del session Error")}
	}
	redisrepo.logger.Info("DeleteSession: Successful session deleted",
		zap.String("traceID", traceID),
	)
	return &RepositoryResponse{Success: true}
}
func NewAuthRedis(client *RedisObject, log *logger.SessionLogger) *AuthRedis {
	return &AuthRedis{Client: client, logger: log}
}
