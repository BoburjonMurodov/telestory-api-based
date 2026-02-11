package models

import (
	"time"
)

type Download struct {
	ID        int       `json:"id"`
	UserID    int64     `json:"user_id"`
	Input     string    `json:"input"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
