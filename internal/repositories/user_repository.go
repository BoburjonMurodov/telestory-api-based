package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/bbr/telestory-api-based/internal/models"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Upsert(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO users (id, first_name, last_name, username, language_code, is_telegram_premium, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			username = EXCLUDED.username,
			language_code = EXCLUDED.language_code,
			is_telegram_premium = EXCLUDED.is_telegram_premium,
			updated_at = NOW();
	`

	_, err := r.DB.ExecContext(ctx, query,
		user.ID,
		user.FirstName,
		user.LastName,
		user.Username,
		user.LanguageCode,
		user.IsTelegramPremium,
	)
	return err
}

func (r *UserRepository) GetByID(id int64) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := &models.User{}
	query := `SELECT id, first_name, last_name, username, language_code, is_telegram_premium, premium_expires_at, role, created_at, updated_at, last_active_at FROM users WHERE id = $1`

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.LanguageCode,
		&user.IsTelegramPremium,
		&user.PremiumExpiresAt,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastActiveAt,
	)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) UpdateActivity(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE users SET last_active_at = NOW() WHERE id = $1`
	_, err := r.DB.ExecContext(ctx, query, id)
	return err
}

func (r *UserRepository) UpdateLanguage(id int64, langCode string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE users SET language_code = $1 WHERE id = $2`
	_, err := r.DB.ExecContext(ctx, query, langCode, id)
	return err
}
