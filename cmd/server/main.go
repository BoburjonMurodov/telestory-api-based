package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bbr/telestory-api-based/internal/controllers"
	"github.com/bbr/telestory-api-based/internal/datasources"
	"github.com/bbr/telestory-api-based/internal/repositories"
	"github.com/bbr/telestory-api-based/internal/services"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Parse command line flags
	envFlag := flag.String("env", "local", "environment to run in (local, prod)")
	flag.Parse()

	// Load environment variables
	loadEnv(*envFlag)

	// Initialize Datasources
	// Initialize database
	db, err := datasources.NewPostgresConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Database connection established")

	// Run migrations automatically
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	bot, err := datasources.NewTelegramBot()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Telegram Bot initialized")

	// Initialize Repositories
	userRepo := repositories.NewUserRepository(db)
	downloadRepo := repositories.NewDownloadRepository(db)

	// Initialize Services
	userService := services.NewUserService(userRepo, downloadRepo)
	downloadService := services.NewDownloadService(downloadRepo)

	// Initialize Controllers
	httpCtrl := controllers.NewHTTPController()
	teleCtrl := controllers.NewTelegramController(bot, userService, downloadService)

	// Setup Handlers
	httpCtrl.SetupRoutes()
	teleCtrl.SetupHandlers()

	// Start Bot in Goroutine
	go bot.Start()
	log.Println("Telegram Bot started")

	// Start HTTP Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// runMigrations automatically runs all SQL migrations in the migrations folder
func runMigrations(db *sql.DB) error {
	log.Println("Running database migrations...")

	// Get migrations directory
	migrationsDir := "migrations"

	// Read all migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migrations: %v", err)
	}

	if len(files) == 0 {
		log.Println("No migration files found")
		return nil
	}

	// Execute each migration file
	for _, file := range files {
		log.Printf("Applying migration: %s", filepath.Base(file))

		// Read migration file
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %v", file, err)
		}

		// Execute migration
		if _, err := db.Exec(string(content)); err != nil {
			// Ignore "already exists" errors
			if !isAlreadyExistsError(err) {
				return fmt.Errorf("failed to apply migration %s: %v", file, err)
			}
			log.Printf("Migration %s already applied, skipping", filepath.Base(file))
		} else {
			log.Printf("Migration %s applied successfully", filepath.Base(file))
		}
	}

	log.Println("All migrations completed successfully")
	return nil
}

// isAlreadyExistsError checks if the error is due to table/column already existing
func isAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return contains(errMsg, "already exists") ||
		contains(errMsg, "duplicate") ||
		contains(errMsg, "42P07") // PostgreSQL duplicate table error code
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func loadEnv(envName string) {
	envFile := ".env." + envName
	err := godotenv.Load(envFile)
	if err != nil {
		log.Printf("Warning: Could not load %s, falling back to .env or system vars: %v", envFile, err)
		godotenv.Load()
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	log.Printf("Starting server in %s mode (loaded %s)", env, envFile)
}
