package services

import (
	"fmt"
	"log"
	"os"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

type LogService struct {
	Bot         *tele.Bot
	LogTargetID int64
}

func NewLogService(bot *tele.Bot) *LogService {
	logChannelID := os.Getenv("LOG_CHANNEL_ID")
	var targetID int64
	if logChannelID != "" {
		id, err := strconv.ParseInt(logChannelID, 10, 64)
		if err == nil {
			targetID = id
		} else {
			log.Printf("Warning: Invalid LOG_CHANNEL_ID format: %v", err)
		}
	}

	return &LogService{
		Bot:         bot,
		LogTargetID: targetID,
	}
}

func (s *LogService) SendLog(message string) {
	if s.LogTargetID == 0 {
		return // Logging not configured
	}

	target := &tele.Chat{ID: s.LogTargetID}
	_, err := s.Bot.Send(target, message, &tele.SendOptions{ParseMode: tele.ModeHTML})
	if err != nil {
		log.Printf("Failed to send log to channel: %v", err)
	}
}

func FormatUserLog(user *tele.User) string {
	premiumStr := "No"
	if user.IsPremium {
		premiumStr = "Yes"
	}

	return fmt.Sprintf(
		"<b>ID:</b> <code>%d</code>\n<b>First Name:</b> %s\n<b>Last Name:</b> %s\n<b>Username:</b> @%s\n<b>Language:</b> %s\n<b>Is Premium:</b> %s",
		user.ID,
		user.FirstName,
		user.LastName,
		user.Username,
		user.LanguageCode,
		premiumStr,
	)
}

func (s *LogService) LogNewUser(user *tele.User) {
	msg := fmt.Sprintf("👤 <b>New User Started Bot</b>\n\n%s", FormatUserLog(user))
	s.SendLog(msg)
}

func (s *LogService) LogSearchRequest(user *tele.User, input string) {
	msg := fmt.Sprintf("🔍 <b>New Search Request</b>\n\n<b>Input:</b> <code>%s</code>\n\n%s", input, FormatUserLog(user))
	s.SendLog(msg)
}
