package services

import (
	"fmt"
	"os"

	"github.com/bbr/telestory-api-based/internal/models"
	"github.com/bbr/telestory-api-based/internal/repositories"
	tele "gopkg.in/telebot.v3"
)

type DownloadService struct {
	DownloadRepo *repositories.DownloadRepository
}

func NewDownloadService(downloadRepo *repositories.DownloadRepository) *DownloadService {
	return &DownloadService{DownloadRepo: downloadRepo}
}

func (s *DownloadService) ProcessDownload(ctx tele.Context, user *models.User, input string) error {
	// TODO: Implement Real Download Logic (e.g., using yt-dlp or similar)
	// For now, we mock it by sending a placeholder text or file.

	// 1. Mock Download & Archive
	archiveChannelID := os.Getenv("ARCHIVE_CHANNEL_ID")
	if archiveChannelID == "" {
		return fmt.Errorf("ARCHIVE_CHANNEL_ID not set")
	}

	// Identify Archive Channel
	// recipient := &tele.Chat{ID: 0} // We need to parse the ID properly, usually handled by Telebot if string is passed to ChatID
	// Easier way: Use integer ID if possible, or Chat object.
	// Telebot expects int64 for ID. We might need to parse it.

	// Simulating "Forward from Channel" flow:
	// Since we don't have a real file yet, we will just echo back to the user for now
	// to verify the flow.

	if err := ctx.Send("✅ [MOCK] Download successful! (Real file download pending implementation)"); err != nil {
		return err
	}

	// 2. Log the download
	download := &models.Download{
		UserID: user.ID,
		Input:  input,
		Status: "success",
	}

	if err := s.DownloadRepo.Create(download); err != nil {
		return fmt.Errorf("failed to log download: %v", err)
	}

	return nil
}
