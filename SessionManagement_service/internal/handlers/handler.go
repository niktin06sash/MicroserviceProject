package handlers

import (
	"context"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/service"
)

type SessionAuthentication interface {
	CreateSession(ctx context.Context, userid string) *service.ServiceResponse
	ValidateSession(ctx context.Context, sessionid string) *service.ServiceResponse
	DeleteSession(ctx context.Context, sessionid string) *service.ServiceResponse
}
type LogProducer interface {
	NewSessionLog(level, place, traceid, msg string)
}

const API_CreateSession = "API-CreateSession"
const API_ValidateSession = "API-ValidateSession"
const API_DeleteSession = "API-DeleteSession"
