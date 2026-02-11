package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID                int64        `json:"id"`
	FirstName         string       `json:"first_name"`
	LastName          string       `json:"last_name"`
	Username          string       `json:"username"`
	PhoneNumber       string       `json:"phone_number"`  // New field
	LanguageCode      string       `json:"language_code"` // Treated as empty string if NULL
	IsTelegramPremium bool         `json:"is_telegram_premium"`
	PremiumExpiresAt  sql.NullTime `json:"premium_expires_at"`
	Role              string       `json:"role"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
	LastActiveAt      sql.NullTime `json:"last_active_at"`
}

func (u *User) IsBotPremium() bool {
	if !u.PremiumExpiresAt.Valid {
		return false
	}
	return u.PremiumExpiresAt.Time.After(time.Now())
}
