package services

import (
	"time"

	"github.com/bbr/telestory-api-based/internal/models"
	"github.com/bbr/telestory-api-based/internal/repositories"
	tele "gopkg.in/telebot.v3"
)

type UserService struct {
	UserRepo     *repositories.UserRepository
	DownloadRepo *repositories.DownloadRepository
}

func NewUserService(userRepo *repositories.UserRepository, downloadRepo *repositories.DownloadRepository) *UserService {
	return &UserService{
		UserRepo:     userRepo,
		DownloadRepo: downloadRepo,
	}
}

func (s *UserService) RegisterUser(teleUser *tele.User) (*models.User, error) {
	user := &models.User{
		ID:                teleUser.ID,
		FirstName:         teleUser.FirstName,
		LastName:          teleUser.LastName,
		Username:          teleUser.Username,
		LanguageCode:      teleUser.LanguageCode,
		IsTelegramPremium: teleUser.IsPremium,
	}

	// if user.LanguageCode == "" {
	// 	user.LanguageCode = "en"
	// }

	if err := s.UserRepo.Upsert(user); err != nil {
		return nil, err
	}

	return s.UserRepo.GetByID(user.ID) // Return full user with db fields
}

func (s *UserService) CanDownload(user *models.User) (bool, string, error) {
	// 1. Check 1-minute cooldown
	if user.LastActiveAt.Valid && time.Since(user.LastActiveAt.Time) < time.Minute {
		return false, "Please wait 1 minute between downloads.", nil
	}

	// 2. Check Daily Limit (if not premium)
	if !user.IsBotPremium() {
		count, err := s.DownloadRepo.CountToday(user.ID)
		if err != nil {
			return false, "", err
		}
		if count >= 3 {
			return false, "Daily limit reached (3/3). Upgrade to Premium for unlimited downloads!", nil
		}
	}

	return true, "", nil
}

func (s *UserService) RecordActivity(userID int64) error {
	return s.UserRepo.UpdateActivity(userID)
}

func (s *UserService) UpdateLanguage(userID int64, langCode string) error {
	// Simple update query via repo (we need to add this to repo too)
	// For now, let's just reuse Upsert but that's heavy.
	// Better to add UpdateLanguage to Repo.
	return s.UserRepo.UpdateLanguage(userID, langCode)
}
