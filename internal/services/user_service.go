package services

import (
	"fmt"
	"os"
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
	// 1. Check if user already exists
	existingUser, err := s.UserRepo.GetByID(teleUser.ID)
	if err == nil {
		// User exists, return it without modifying
		return existingUser, nil
	}

	// 2. User doesn't exist, create new one
	// DO NOT set language_code - let it be NULL until user chooses
	user := &models.User{
		ID:                teleUser.ID,
		FirstName:         teleUser.FirstName,
		LastName:          teleUser.LastName,
		Username:          teleUser.Username,
		LanguageCode:      "",
		IsTelegramPremium: teleUser.IsPremium,
	}

	if err := s.UserRepo.Insert(user); err != nil {
		return nil, err
	}

	return s.UserRepo.GetByID(user.ID)
}

func (s *UserService) CanDownload(user *models.User) (bool, string, error) {
	// Get environment-based limits
	env := os.Getenv("APP_ENV")

	var cooldownDuration time.Duration
	var dailyLimit int

	if env == "production" {
		cooldownDuration = 1 * time.Minute
		dailyLimit = 3
	} else {
		// Development/local environment
		cooldownDuration = 10 * time.Second
		dailyLimit = 100
	}

	// 1. Check cooldown
	if user.LastActiveAt.Valid && time.Since(user.LastActiveAt.Time) < cooldownDuration {
		remainingTime := cooldownDuration - time.Since(user.LastActiveAt.Time)
		return false, fmt.Sprintf("Please wait %d seconds between downloads.", int(remainingTime.Seconds())), nil
	}

	// 2. Check Daily Limit (if not premium)
	if !user.IsBotPremium() {
		count, err := s.DownloadRepo.CountToday(user.ID)
		if err != nil {
			return false, "", err
		}
		if count >= dailyLimit {
			return false, fmt.Sprintf("Daily limit reached (%d/%d). Upgrade to Premium for unlimited downloads!", count, dailyLimit), nil
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
