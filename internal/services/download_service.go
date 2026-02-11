package services

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/bbr/telestory-api-based/internal/i18n"
	"github.com/bbr/telestory-api-based/internal/models"
	"github.com/bbr/telestory-api-based/internal/repositories"
	tele "gopkg.in/telebot.v3"
)

type DownloadService struct {
	DownloadRepo *repositories.DownloadRepository
	HTTPClient   *http.Client
}

func NewDownloadService(downloadRepo *repositories.DownloadRepository) *DownloadService {
	return &DownloadService{
		DownloadRepo: downloadRepo,
		HTTPClient:   &http.Client{},
	}
}

// TeleStoryResponse represents the API response structure
type TeleStoryResponse struct {
	Stories []Story `json:"stories"`
	Message string  `json:"message"`
	Success bool    `json:"success"`
}

type Story struct {
	ID        string `json:"id"`
	MediaURL  string `json:"media_url"`
	Timestamp int64  `json:"timestamp"`
}

func (s *DownloadService) FetchStoriesByInput(input string) (*TeleStoryResponse, error) {
	apiKey := os.Getenv("TELESTORY_API_KEY")
	apiURL := os.Getenv("TELESTORY_API_URL")

	if apiKey == "" || apiURL == "" {
		return nil, fmt.Errorf("TELESTORY_API_KEY or TELESTORY_API_URL not set")
	}

	// Clean input (remove @ for username or + for phone number)
	cleanInput := strings.TrimPrefix(input, "@")
	cleanInput = strings.TrimPrefix(cleanInput, "+")

	// Build request URL
	reqURL := fmt.Sprintf("%s/get_stories_by_username?api_key=%s&username=%s&archive=true&mark=false",
		apiURL, url.QueryEscape(apiKey), url.QueryEscape(cleanInput))

	// Create request
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "TeleStory Android Client v1.43Build: 79, Patch: 20250820")
	req.Header.Set("Accept-Encoding", "gzip")

	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Handle gzip response
	var reader io.ReadCloser
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %v", err)
		}
		defer reader.Close()
	} else {
		reader = resp.Body
	}

	// Parse JSON response
	var apiResp TeleStoryResponse
	if err := json.NewDecoder(reader).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &apiResp, nil
}

func (s *DownloadService) ProcessDownload(ctx tele.Context, user *models.User, input string) error {
	// Get user's language
	userLang := user.LanguageCode
	if userLang == "" {
		userLang = "en"
	}

	// Fetch stories from TeleStory API
	apiResp, err := s.FetchStoriesByInput(input)
	if err != nil {
		return ctx.Send(fmt.Sprintf(i18n.GetMessage(userLang, "fetch_error"), err.Error()))
	}

	// Count stories
	storyCount := len(apiResp.Stories)

	// Send result to user
	var message string
	if storyCount == 0 {
		message = fmt.Sprintf(i18n.GetMessage(userLang, "no_stories"), input)
	} else {
		message = fmt.Sprintf(i18n.GetMessage(userLang, "story_count"), storyCount, input)
	}

	if err := ctx.Send(message); err != nil {
		return err
	}

	// Log the download
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
