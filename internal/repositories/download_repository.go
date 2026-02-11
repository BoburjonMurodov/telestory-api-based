package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/bbr/telestory-api-based/internal/models"
)

type DownloadRepository struct {
	DB *sql.DB
}

func NewDownloadRepository(db *sql.DB) *DownloadRepository {
	return &DownloadRepository{DB: db}
}

func (r *DownloadRepository) Create(download *models.Download) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO downloads (user_id, input, status, created_at) VALUES ($1, $2, $3, NOW()) RETURNING id, created_at`
	return r.DB.QueryRowContext(ctx, query, download.UserID, download.Input, download.Status).Scan(&download.ID, &download.CreatedAt)
}

func (r *DownloadRepository) CountToday(userID int64) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT COUNT(*) 
		FROM downloads 
		WHERE user_id = $1 
		  AND status = 'success'
		  AND created_at >= CURRENT_DATE
	`
	var count int
	err := r.DB.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}
