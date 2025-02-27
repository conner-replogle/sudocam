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

	"server/handlers"
	"server/middleware"
	"server/models"
	"server/websocket"
)

func setupDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.User{}, &models.Camera{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func setupRoutes(db *gorm.DB) {
	jwtKey := []byte(os.Getenv("JWT_SECRET"))

	// Auth routes
	http.HandleFunc("/api/signup", handlers.HandleSignup(db))
	http.HandleFunc("/api/login", handlers.HandleLogin(db, jwtKey))
	http.HandleFunc("/api/validate", handlers.ValidateToken)

	// User routes
	http.HandleFunc("/api/users/cameras", middleware.AuthMiddleware(handlers.UsersCameras(db)))
	// Camera routes
	http.HandleFunc("/api/cameras/generate", middleware.AuthMiddleware(handlers.HandleGenerateCamera(jwtKey)))
	http.HandleFunc("/api/cameras/register", handlers.HandleRegisterCamera(db, jwtKey))
	http.HandleFunc("/api/camera/status", handlers.UpdateCameraStatus(db))
	http.HandleFunc("/api/camera/ping", handlers.PingCamera(db))

	// WebSocket route
	http.HandleFunc("/ws", websocket.HandleWebSocket(db))

	// TURN credentials route
	http.HandleFunc("/api/turn", handlers.HandleTURNCredentials())

	// Serve static files
	http.Handle("/", spa.SpaHandler("ui/dist", "index.html"))
}

func startServer() {
	// Create shutdown channel
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server
	go func() {
		slog.Info("Starting server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			slog.Error("Server error", "error", err)
			shutdownChan <- syscall.SIGTERM
		}
	}()

	// Wait for shutdown signal
	<-shutdownChan
	slog.Info("Shutting down server...")
}

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
		slog.Error("Failed to setup database", "error", err)
		os.Exit(1)
	}

	setupRoutes(db)
	startServer()
}
