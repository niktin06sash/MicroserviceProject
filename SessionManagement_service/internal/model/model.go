package model

import (
	"time"
)

type Session struct {
	SessionID      string    `json:"sessionid"`
	UserID         string    `json:"userid"`
	ExpirationTime time.Time `json:"expiration_time"`
}
