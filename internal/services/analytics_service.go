package services

import (
	"fmt"

	"github.com/bbr/telestory-api-based/internal/i18n"
	"github.com/bbr/telestory-api-based/internal/repositories"
)

type AnalyticsService struct {
	UserRepo     *repositories.UserRepository
	DownloadRepo *repositories.DownloadRepository
}

func NewAnalyticsService(userRepo *repositories.UserRepository, downloadRepo *repositories.DownloadRepository) *AnalyticsService {
	return &AnalyticsService{
		UserRepo:     userRepo,
		DownloadRepo: downloadRepo,
	}
}

// GetOverallStats compiles all interesting analytics into a Markdown-formatted string based on language
func (s *AnalyticsService) GetOverallStats(langCode string) (string, error) {
	totalUsers, err := s.UserRepo.CountAllUsers()
	if err != nil {
		return "", fmt.Errorf("failed to count users: %v", err)
	}

	activeUsers7, err := s.UserRepo.CountActiveUsers(7)
	if err != nil {
		return "", fmt.Errorf("failed to count active 7d users: %v", err)
	}

	activeUsers30, err := s.UserRepo.CountActiveUsers(30)
	if err != nil {
		return "", fmt.Errorf("failed to count active 30d users: %v", err)
	}

	totalDownloads, err := s.DownloadRepo.CountTotalDownloads()
	if err != nil {
		return "", fmt.Errorf("failed to count total downloads: %v", err)
	}

	successDownloads, err := s.DownloadRepo.CountDownloadsByStatus("success")
	if err != nil {
		return "", fmt.Errorf("failed to count success downloads: %v", err)
	}

	failedDownloads, err := s.DownloadRepo.CountDownloadsByStatus("failed")
	if err != nil {
		return "", fmt.Errorf("failed to count failed downloads: %v", err)
	}

	totalDownloadsToday, err := s.DownloadRepo.CountDownloadsToday()
	if err != nil {
		return "", fmt.Errorf("failed to count total downloads today: %v", err)
	}

	successDownloadsToday, err := s.DownloadRepo.CountDownloadsTodayByStatus("success")
	if err != nil {
		return "", fmt.Errorf("failed to count success downloads today: %v", err)
	}

	failedDownloadsToday, err := s.DownloadRepo.CountDownloadsTodayByStatus("failed")
	if err != nil {
		return "", fmt.Errorf("failed to count failed downloads today: %v", err)
	}

	template := i18n.GetMessage(langCode, "stats_report")
	report := fmt.Sprintf(
		template,
		totalUsers, activeUsers7, activeUsers30,
		totalDownloadsToday, successDownloadsToday, failedDownloadsToday,
		totalDownloads, successDownloads, failedDownloads,
	)

	return report, nil
}
