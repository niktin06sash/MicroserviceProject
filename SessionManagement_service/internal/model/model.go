package model

import (
	"time"
)

type Session struct {
	SessionID      string
	UserID         string
	ExpirationTime time.Time
}
