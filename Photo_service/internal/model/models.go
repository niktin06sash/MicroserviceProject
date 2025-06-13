package model

import "time"

type Photo struct {
	ID          string    `json:"photo_id"`
	UserID      string    `json:"user_id"`
	URL         string    `json:"url"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
}
