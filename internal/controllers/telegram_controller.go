package controllers

import (
	"fmt"
	"log"

	"github.com/bbr/telestory-api-based/internal/i18n"
	"github.com/bbr/telestory-api-based/internal/services"
	tele "gopkg.in/telebot.v3"
)

type TelegramController struct {
	Bot             *tele.Bot
	UserService     *services.UserService
	DownloadService *services.DownloadService
}

func NewTelegramController(bot *tele.Bot, userService *services.UserService, downloadService *services.DownloadService) *TelegramController {
	return &TelegramController{
		Bot:             bot,
		UserService:     userService,
		DownloadService: downloadService,
	}
}

func (c *TelegramController) SetupHandlers() {
	c.Bot.Handle("/start", c.StartHandler)
	c.Bot.Handle(tele.OnText, c.TextHandler)
	c.Bot.Handle(tele.OnCallback, c.LanguageCallback)
}

func (c *TelegramController) StartHandler(ctx tele.Context) error {
	// Register user first in EN default
	_, err := c.UserService.RegisterUser(ctx.Sender())
	if err != nil {
		return ctx.Send("Welcome! (Error saving profile)")
	}

	// Send Language Selection Menu
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

func (c *TelegramController) LanguageCallback(ctx tele.Context) error {
	langCode := ctx.Args()[0] // "en", "uz", or "ru"
	userID := ctx.Sender().ID

	// 1. Update Lang in DB
	if err := c.UserService.UpdateLanguage(userID, langCode); err != nil {
		return ctx.Respond(&tele.CallbackResponse{Text: "Error updating language"})
	}

	// 2. Delete Menu
	c.Bot.Delete(ctx.Message())

	// 3. Send Confirmation & Instructions in new Lang
	msg := fmt.Sprintf("%s\n\n%s",
		i18n.GetMessage(langCode, "registered"),
		i18n.GetMessage(langCode, "instruction"),
	)

	return ctx.Send(msg)
}

func (c *TelegramController) TextHandler(ctx tele.Context) error {
	input := ctx.Text()
	teleUser := ctx.Sender()

	// 1. Get/Register User (Ensure we have latest data)
	user, err := c.UserService.RegisterUser(teleUser)
	if err != nil {
		log.Printf("Error registering user: %v", err)
		return ctx.Send("An error occurred. Please try again.")
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

	// 3. Process Download
	statusMsg, _ := c.Bot.Send(teleUser, "⏳ Processing...")

	err = c.DownloadService.ProcessDownload(ctx, user, input)

	// Cleanup status message
	c.Bot.Delete(statusMsg)

	if err != nil {
		log.Printf("Download error: %v", err)
		return ctx.Send("❌ Failed to process request.")
	}

	// 4. Update Activity timestamp
	c.UserService.RecordActivity(user.ID)

	return nil
}
