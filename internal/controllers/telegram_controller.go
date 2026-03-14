package controllers

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bbr/telestory-api-based/internal/i18n"
	"github.com/bbr/telestory-api-based/internal/services"
	tele "gopkg.in/telebot.v3"
)

type TelegramController struct {
	Bot              *tele.Bot
	UserService      *services.UserService
	DownloadService  *services.DownloadService
	LogService       *services.LogService
	AnalyticsService *services.AnalyticsService
}

func NewTelegramController(bot *tele.Bot, userService *services.UserService, downloadService *services.DownloadService, logService *services.LogService, analyticsService *services.AnalyticsService) *TelegramController {
	return &TelegramController{
		Bot:              bot,
		UserService:      userService,
		DownloadService:  downloadService,
		LogService:       logService,
		AnalyticsService: analyticsService,
	}
}

func (c *TelegramController) SetupHandlers() {
	c.Bot.Handle("/start", c.StartHandler)
	c.Bot.Handle("/stats", c.StatsHandler)
	c.Bot.Handle(tele.OnText, c.TextHandler)
	c.Bot.Handle(tele.OnCallback, c.LanguageCallback)

	// Unsupported inputs
	c.Bot.Handle(tele.OnPhoto, c.UnsupportedHandler)
	c.Bot.Handle(tele.OnVideo, c.UnsupportedHandler)
	c.Bot.Handle(tele.OnDocument, c.UnsupportedHandler)
}

func (c *TelegramController) showLanguageMenu(ctx tele.Context) error {
	menu := &tele.ReplyMarkup{}
	btnEn := menu.Data("🇺🇸 English", "lang", "en")
	btnUz := menu.Data("🇺🇿 O'zbek", "lang", "uz")
	btnRu := menu.Data("🇷🇺 Русский", "lang", "ru")

	menu.Inline(
		menu.Row(btnEn),
		menu.Row(btnUz),
		menu.Row(btnRu),
	)
	return ctx.Send(i18n.GetMessage("en", "welcome"), menu)
}

func (c *TelegramController) UnsupportedHandler(ctx tele.Context) error {
	teleUser := ctx.Sender()
	user, err := c.UserService.RegisterUser(teleUser)
	if err != nil {
		return nil
	}
	if user.LanguageCode == "" {
		return c.showLanguageMenu(ctx)
	}
	return ctx.Send(i18n.GetMessage(user.LanguageCode, "invalid_input"))
}

func (c *TelegramController) StartHandler(ctx tele.Context) error {
	teleUser := ctx.Sender()

	// Log new user to admin channel
	c.LogService.LogNewUser(teleUser)

	// Register user first
	log.Println("user: ", teleUser)
	user, err := c.UserService.RegisterUser(teleUser)
	if err != nil {
		return ctx.Send("Welcome!")
	}

	if user.LanguageCode == "" {
		return c.showLanguageMenu(ctx)
	}

	// 2. If Language IS set, show instructions directly
	msg := i18n.GetMessage(user.LanguageCode, "instruction")
	return ctx.Send(msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func (c *TelegramController) LanguageCallback(ctx tele.Context) error {
	// Get the language code from callback data
	// Callback data format is "unique|data", e.g., "lang|uz"
	callbackData := ctx.Callback().Data
	parts := strings.Split(callbackData, "|")

	var langCode string
	if len(parts) == 2 {
		langCode = parts[1] // Extract "uz" from "lang|uz"
	} else {
		langCode = callbackData // Fallback if format is different
	}

	userID := ctx.Sender().ID

	log.Printf("Language selected: %s for user %d", langCode, userID)

	// 1. Update Lang in DB
	if err := c.UserService.UpdateLanguage(userID, langCode); err != nil {
		log.Printf("Error updating language: %v", err)
		return ctx.Respond(&tele.CallbackResponse{Text: "Error updating language"})
	}

	// 2. Respond to callback (removes loading state)
	ctx.Respond(&tele.CallbackResponse{})

	// 3. Delete Menu
	c.Bot.Delete(ctx.Message())

	// 4. Send Confirmation & Instructions in new Lang
	msg := fmt.Sprintf("%s\n\n%s",
		i18n.GetMessage(langCode, "registered"),
		i18n.GetMessage(langCode, "instruction"),
	)

	return ctx.Send(msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func isValidSearchInput(input string) bool {
	input = strings.TrimSpace(input)

	// If it contains a URL/slash or any spaces, it's invalid
	if strings.Contains(input, " ") || strings.Contains(input, "\n") || strings.Contains(input, "/") {
		return false
	}

	if len(input) < 3 || len(input) > 32 {
		return false
	}

	return true
}

func (c *TelegramController) TextHandler(ctx tele.Context) error {
	input := ctx.Text()
	teleUser := ctx.Sender()

	// Log the search request
	c.LogService.LogSearchRequest(teleUser, input)

	// 1. Get/Register User (Ensure we have latest data)
	user, err := c.UserService.RegisterUser(teleUser)
	if err != nil {
		log.Printf("Error registering user: %v", err)
		return ctx.Send("An error occurred. Please try again.")
	}

	if user.LanguageCode == "" {
		return c.showLanguageMenu(ctx)
	}

	if os.Getenv("MAINTENANCE") == "true" {
		return ctx.Send(i18n.GetMessage(user.LanguageCode, "maintenance"))
	}

	if !isValidSearchInput(input) {
		return ctx.Send(i18n.GetMessage(user.LanguageCode, "invalid_input"))
	}

	// 2. Check Limits
	allowed, reason, err := c.UserService.CanDownload(user)
	if err != nil {
		log.Printf("Error checking limits: %v", err)
		return ctx.Send("System error checking limits.")
	}
	if !allowed {
		return ctx.Send("🚫 " + reason)
	}

	// 3. Send Processing Message (localized)
	processingMsg := i18n.GetMessage(user.LanguageCode, "processing")
	sentMsg, err := c.Bot.Send(teleUser, processingMsg)
	if err != nil {
		log.Printf("Error sending processing message: %v", err)
		return ctx.Send("An error occurred.")
	}

	// 4. Record Activity
	if err := c.UserService.RecordActivity(user.ID); err != nil {
		log.Printf("Error recording activity: %v", err)
	}

	// 5. Process Download (will edit the sentMsg with result)
	if err := c.DownloadService.ProcessDownloadWithEdit(c.Bot, sentMsg, user, input); err != nil {
		log.Printf("Error processing download: %v", err)
		return err
	}

	return nil
}

func (c *TelegramController) StatsHandler(ctx tele.Context) error {
	teleUser := ctx.Sender()

	// 1. Get/Register User
	user, err := c.UserService.RegisterUser(teleUser)
	if err != nil {
		log.Printf("Error registering user: %v", err)
		return ctx.Send("An error occurred. Please try again.")
	}

	// 2. Check if user is an admin
	if user.Role != "admin" {
		log.Printf("User %d (role: %s) attempted to access /stats but was denied.", user.ID, user.Role)
		return nil // Ignore silently
	}

	// 3. Fetch analytics report
	report, err := c.AnalyticsService.GetOverallStats(user.LanguageCode)
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
		return ctx.Send("Failed to fetch analytics.")
	}

	return ctx.Send(report, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}
