package services

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

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
	OK      bool    `json:"ok"`
	Stories []Story `json:"stories"`
	BaseURL string  `json:"base_url"`
	Success bool    `json:"success"`
}

type Story struct {
	URL     string `json:"url"`
	Date    int64  `json:"date"`
	Caption string `json:"caption"`
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
	reqURL := fmt.Sprintf("%s/get_stories_by_username?api_key=%s&username=%s&archive=true&mark=true",
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

	if err := ctx.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdown}); err != nil {
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

// DownloadStoryMedia downloads a story from URL to temp file
func (s *DownloadService) DownloadStoryMedia(baseURL, storyURL string, index int) (string, error) {
	if baseURL == "" {
		return "", fmt.Errorf("base URL is empty")
	}

	// Build full URL - ensure proper path separator
	fullURL := baseURL
	if !strings.HasSuffix(baseURL, "/") && !strings.HasPrefix(storyURL, "/") {
		fullURL += "/"
	}
	fullURL += storyURL

	log.Printf("Attempting download from: %s", fullURL)

	// Create temp file
	ext := filepath.Ext(storyURL)
	if ext == "" {
		ext = ".mp4" // default to video
	}
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("telestory-%d-%d%s", time.Now().Unix(), index, ext))

	// Download file
	resp, err := s.HTTPClient.Get(fullURL)
	if err != nil {
		return "", fmt.Errorf("failed to download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create file
	out, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	// Write to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tempFile)
		return "", fmt.Errorf("failed to write file: %v", err)
	}

	return tempFile, nil
}

// ProcessDownloadWithEdit edits an existing message with the result and downloads/uploads stories
func (s *DownloadService) ProcessDownloadWithEdit(bot *tele.Bot, msg *tele.Message, user *models.User, input string) error {
	// Get user's language
	userLang := user.LanguageCode
	if userLang == "" {
		userLang = "en"
	}

	// Get archive channel ID
	archiveChannelID := os.Getenv("ARCHIVE_CHANNEL_ID")
	if archiveChannelID == "" {
		return fmt.Errorf("ARCHIVE_CHANNEL_ID not set")
	}

	// Fetch stories from TeleStory API
	apiResp, err := s.FetchStoriesByInput(input)
	if err != nil {
		errorMsg := fmt.Sprintf(i18n.GetMessage(userLang, "fetch_error"), err.Error())
		bot.Edit(msg, errorMsg)
		return err
	}

	log.Printf("API Response - BaseURL: '%s', Stories: %d", apiResp.BaseURL, len(apiResp.Stories))

	// Count stories
	storyCount := len(apiResp.Stories)

	if storyCount == 0 {
		message := fmt.Sprintf(i18n.GetMessage(userLang, "no_stories"), input)
		bot.Edit(msg, message, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
		return nil
	}

	// Edit message to show downloading status
	downloadingMsg := fmt.Sprintf(i18n.GetMessage(userLang, "downloading"), storyCount)
	bot.Edit(msg, downloadingMsg)

	log.Printf("Using base URL for downloads: %s", apiResp.BaseURL)

	// Download stories asynchronously
	type downloadResult struct {
		index    int
		filePath string
		story    Story
		err      error
	}

	results := make(chan downloadResult, storyCount)
	var wg sync.WaitGroup

	log.Printf("Starting async download of %d stories", storyCount)
	for i, story := range apiResp.Stories {
		wg.Add(1)
		go func(idx int, st Story) {
			defer wg.Done()
			log.Printf("Downloading story %d: %s", idx, st.URL)
			filePath, err := s.DownloadStoryMedia(apiResp.BaseURL, st.URL, idx)
			if err != nil {
				log.Printf("Failed to download story %d: %v", idx, err)
			} else {
				log.Printf("Successfully downloaded story %d to %s", idx, filePath)
			}
			results <- downloadResult{index: idx, filePath: filePath, story: st, err: err}
		}(i, story)
	}

	// Wait for all downloads to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	downloaded := make([]downloadResult, 0, storyCount)
	for result := range results {
		if result.err == nil {
			downloaded = append(downloaded, result)
		}
	}

	log.Printf("Downloaded %d/%d stories successfully", len(downloaded), storyCount)

	// Upload to archive and forward to user
	archiveChatID, _ := strconv.ParseInt(archiveChannelID, 10, 64)
	archiveChat, _ := bot.ChatByID(archiveChatID)
	userChat := &tele.User{ID: user.ID}

	log.Printf("Archive chat ID: %d, User ID: %d", archiveChatID, user.ID)

	successCount := 0
	for _, result := range downloaded {
		// Build caption for archive channel (detailed)
		storyDate := time.Unix(result.story.Date, 0).Format("2006-01-02 15:04")
		archiveCaption := fmt.Sprintf(
			"📥 Requested by: %s %s (@%s)\n📍 Target: %s\n📅 Story Date: %s\n\n%s",
			user.FirstName,
			user.LastName,
			user.Username,
			input,
			storyDate,
			result.story.Caption,
		)

		// Build caption for user (simple)
		userCaption := result.story.Caption
		if userCaption == "" {
			userCaption = fmt.Sprintf(i18n.GetMessage(userLang, "story_from"), input)
		}
		userCaption = fmt.Sprintf("%s\n\n📅 %s", userCaption, storyDate)

		// Determine media type by file extension
		var archiveMsg *tele.Message
		var uploadErr error

		log.Printf("Uploading story to archive: %s", result.filePath)
		if strings.HasSuffix(result.filePath, ".mp4") || strings.HasSuffix(result.filePath, ".mov") {
			// Upload video to archive
			video := &tele.Video{File: tele.FromDisk(result.filePath), Caption: archiveCaption}
			archiveMsg, uploadErr = bot.Send(archiveChat, video)
		} else {
			// Upload photo to archive
			photo := &tele.Photo{File: tele.FromDisk(result.filePath), Caption: archiveCaption}
			archiveMsg, uploadErr = bot.Send(archiveChat, photo)
		}

		if uploadErr != nil {
			log.Printf("Failed to upload to archive: %v", uploadErr)
			os.Remove(result.filePath)
			continue
		}

		log.Printf("Uploaded to archive successfully, message ID: %d", archiveMsg.ID)

		// Send to user using file ID from archive message (not forwarding)
		var userSendErr error
		log.Printf("Sending to user %d", user.ID)

		if strings.HasSuffix(result.filePath, ".mp4") || strings.HasSuffix(result.filePath, ".mov") {
			// Send video to user
			video := &tele.Video{
				File:    tele.File{FileID: archiveMsg.Video.FileID},
				Caption: userCaption,
			}
			_, userSendErr = bot.Send(userChat, video)
		} else {
			// Send photo to user
			photo := &tele.Photo{
				File:    tele.File{FileID: archiveMsg.Photo.FileID},
				Caption: userCaption,
			}
			_, userSendErr = bot.Send(userChat, photo)
		}

		if userSendErr != nil {
			log.Printf("Failed to send to user: %v", userSendErr)
		} else {
			log.Printf("Sent to user successfully")
			successCount++
		}

		// Cleanup temp file
		os.Remove(result.filePath)
	}

	log.Printf("Successfully sent %d/%d stories to user", successCount, len(downloaded))

	// Delete processing message
	bot.Delete(msg)

	// If some stories failed to download, notify user
	if len(downloaded) < storyCount {
		errorMsg := fmt.Sprintf(i18n.GetMessage(userLang, "download_error"), len(downloaded), storyCount)
		bot.Send(&tele.User{ID: user.ID}, errorMsg)
	}

	// Log the download
	download := &models.Download{
		UserID: user.ID,
		Input:  input,
		Status: "success",
	}
	s.DownloadRepo.Create(download)

	return nil
}
