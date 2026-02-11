package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/bbr/telestory-api-based/internal/controllers"
	"github.com/bbr/telestory-api-based/internal/datasources"
	"github.com/bbr/telestory-api-based/internal/repositories"
	"github.com/bbr/telestory-api-based/internal/services"
	"github.com/joho/godotenv"
)

func main() {
	// Parse command line flags
	envFlag := flag.String("env", "local", "environment to run in (local, prod)")
	flag.Parse()

	// Load environment variables
	loadEnv(*envFlag)

	// Initialize Datasources
	db, err := datasources.NewPostgresConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Database connection established")

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
		log.Fatal("ListenAndServe: ", err)
	}
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
