package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	spa "github.com/roberthodgen/spa-server"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"your-project-path/handlers"
	"your-project-path/middleware"
	"your-project-path/websocket"
)

func main() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	if err := godotenv.Load("../.env"); err != nil {
		slog.Error("Error loading .env file", "error", err)
	}

	jwtKey := []byte(os.Getenv("JWT_SECRET"))
	middleware.SetJWTKey(jwtKey)

	db, err := setupDatabase()
	if err != nil {
		panic(err)
	}

	setupRoutes(db)
	startServer()
}

// ... helper functions for setup ...