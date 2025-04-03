package model

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	SessionID      string
	UserID         uuid.UUID
	ExpirationTime time.Time
}
