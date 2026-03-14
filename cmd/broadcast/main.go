package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	tele "gopkg.in/telebot.v3"
)

// Maintenance messages in each supported language.
var messages = map[string]string{
	"en": "✅ *We're back\\!*\n\nThe service has been restored and everything is working again\\. Thanks for your patience\\! 🙏",
	"uz": "✅ *Qaytib keldik\\!*\n\nXizmat tiklandi va hamma narsa yana ishlayapti\\. Sabringiz uchun rahmat\\! 🙏",
	"ru": "✅ *Мы вернулись\\!*\n\nСервис восстановлен и всё снова работает\\. Спасибо за терпение\\! 🙏",
}

func main() {
	envFlag := flag.String("env", "prod", "environment to load (local, prod)")
	flag.Parse()

	loadEnv(*envFlag)

	db, err := connectDB()
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer db.Close()
	log.Println("Database connected")

	bot, err := newBot()
	if err != nil {
		log.Fatalf("Bot init failed: %v", err)
	}
	log.Println("Bot initialized")

	users, err := fetchAllUsers(db)
	if err != nil {
		log.Fatalf("Failed to fetch users: %v", err)
	}
	log.Printf("Sending broadcast to %d users...", len(users))

	total := len(users)
	var sent, skipped, failed int
	for i, u := range users {
		msg, ok := messages[u.lang]
		if !ok {
			msg = messages["en"]
		}

		recipient := &tele.User{ID: u.id}
		_, err := bot.Send(recipient, msg, &tele.SendOptions{ParseMode: tele.ModeMarkdownV2})
		if err != nil {
			if isBotBlockedError(err) {
				skipped++
			} else if retryAfter := retryAfterSeconds(err); retryAfter > 0 {
				// Telegram is rate-limiting us — wait and retry once
				log.Printf("Rate limited, waiting %ds...", retryAfter)
				time.Sleep(time.Duration(retryAfter+1) * time.Second)
				if _, err2 := bot.Send(recipient, msg, &tele.SendOptions{ParseMode: tele.ModeMarkdownV2}); err2 != nil {
					log.Printf("Retry failed for user %d: %v", u.id, err2)
					failed++
				} else {
					sent++
				}
			} else {
				log.Printf("Failed to send to user %d: %v", u.id, err)
				failed++
			}
		} else {
			sent++
		}

		// Progress every 50 users
		if (i+1)%50 == 0 || i+1 == total {
			log.Printf("Progress: %d/%d (sent=%d skipped=%d failed=%d)", i+1, total, sent, skipped, failed)
		}

		// Respect Telegram rate limit: ~20 messages/second to different users
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Printf("\nBroadcast complete: %d sent, %d skipped (blocked), %d failed\n", sent, skipped, failed)
}

type userRow struct {
	id   int64
	lang string
}

func fetchAllUsers(db *sql.DB) ([]userRow, error) {
	rows, err := db.Query(`SELECT id, COALESCE(language_code, 'en') FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []userRow
	for rows.Next() {
		var u userRow
		if err := rows.Scan(&u.id, &u.lang); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// retryAfterSeconds returns the retry-after value from a Telegram 429 error, or 0.
func retryAfterSeconds(err error) int {
	if err == nil {
		return 0
	}
	if floodErr, ok := err.(*tele.FloodError); ok {
		return floodErr.RetryAfter
	}
	return 0
}

func isBotBlockedError(err error) bool {
	if err == nil {
		return false
	}
	e := err.Error()
	return contains(e, "403") || contains(e, "blocked") || contains(e, "bot was blocked") || contains(e, "user is deactivated")
}

func contains(s, sub string) bool {
	if len(sub) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func connectDB() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func newBot() (*tele.Bot, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is not set")
	}
	// No poller — we only need Send for broadcasting
	return tele.NewBot(tele.Settings{Token: token})
}

func loadEnv(envName string) {
	envFile := ".env." + envName
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: could not load %s, falling back to .env: %v", envFile, err)
		godotenv.Load()
	}
}
