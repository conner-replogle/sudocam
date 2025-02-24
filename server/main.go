package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	pb "messages/msgspb"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket" // Import the gorilla/websocket library
	"github.com/joho/godotenv"
	spa "github.com/roberthodgen/spa-server"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Define a global upgrader for WebSocket connections
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (for development).  In production, _restrict_ this!
	},
}

// Map to store WebSocket connections
var websocketConnections = make(map[string]*websocket.Conn)
var websocketConnectionsMutex sync.Mutex

// Add these at package level
var (
	// Add channel to coordinate shutdown
	shutdownChan = make(chan struct{})
	wg           sync.WaitGroup
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

type User struct {
	gorm.Model
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"password"`
}

type Camera struct {
	gorm.Model
	Code      string `json:"code" gorm:"unique"`
	UserID uint `json:"userID"`

}

// Login request structure
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// JWT claims structure
type Claims struct {
	Email string `json:"email"`
	UserID uint `json:"userID"`
	jwt.RegisteredClaims
}

// TURNCredentials represents the response from Cloudflare's TURN API
type TURNCredentials struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	TTL      int      `json:"ttl"`
	URIs     []string `json:"uris"`
}

type TokenResponse struct {
	Valid bool   `json:"valid"`
	Email string `json:"email,omitempty"`
}

type CameraAdd struct {
	UserID uint `json:"userID"`
	ServerURL string `json:"serverURL"`
	jwt.RegisteredClaims
}

var db *gorm.DB

func main() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	// Load .env file
	if err := godotenv.Load("../.env"); err != nil {
		slog.Error("Error loading .env file", "error", err)
	}

	dB, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db = dB
	db.AutoMigrate(&User{}, &Camera{})

	// Create a channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Add auth endpoints before the spa handler
	http.HandleFunc("/api/signup", handleSignup)
	http.HandleFunc("/api/login", handleLogin)

	// Add the WebSocket handler
	http.HandleFunc("/ws", handleWebSocket)

	// Add the TURN credentials endpoint
	http.HandleFunc("/api/turn", handleTURNCredentials)

	// Add new validation endpoint before the spa handler
	http.HandleFunc("/api/validate", validateToken)

	// Add new endpoint before spa handler
	http.HandleFunc("/api/cameras/generate", authMiddleware(handleGenerateCamera))

	// HTTP Server setup
	http.Handle("/", spa.SpaHandler("ui/dist", "index.html"))
	go func() {
		slog.Info("HTTP server listening on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			slog.Error("HTTP server error", "error", err)
			sigChan <- syscall.SIGTERM // Signal other goroutines to shut down
		}
	}()

	// Block until a signal is received
	<-sigChan
	slog.Info("\nReceived shutdown signal. Cleaning up...")

	// Signal all goroutines to stop
	close(shutdownChan)

	// Wait for goroutines to finish
	slog.Info("Waiting for goroutines to finish...")
	wg.Wait()
	slog.Info("All goroutines finished")

	// Close all WebSocket connections
	websocketConnectionsMutex.Lock()
	for id := range websocketConnections {
		websocketConnections[id].Close()
	}
	websocketConnectionsMutex.Unlock()
	slog.Info("Shutdown complete.")

}

// handleWebSocket handles WebSocket connections
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", "error", err)
		return
	}

	msg := &pb.Message{}
	readProtoMessage(conn, msg)
	slog.Info("received message", "message", msg)
	ident := msg.GetInitalization()
	if ident == nil {
		slog.Error("Invalid Init Message")
		conn.Close()
		return
	}
	id := ident.GetId()

	websocketConnectionsMutex.Lock()
	websocketConnections[id] = conn // Add the new connection to the map
	websocketConnectionsMutex.Unlock()

	defer func() {
		websocketConnectionsMutex.Lock()
		delete(websocketConnections, id) // Remove the connection when it's closed
		websocketConnectionsMutex.Unlock()
		conn.Close()
		slog.Info("WebSocket connection closed")
	}()
	slog.Info("WebSocket connection established")
	for {
		err := readProtoMessage(conn, msg)
		if err != nil {
			slog.Error("Error reading WebSocket message", "error", err)

			return
		}

		// Marshal the message back to binary
		b, err := proto.Marshal(msg)
		if err != nil {
			slog.Error("Error marshaling protobuf message", "error", err)
			continue
		}
		targ := websocketConnections[msg.To]
		if targ == nil {
			slog.Error("No target connection", "id", msg.To)
			continue
		}
		targ.WriteMessage(websocket.BinaryMessage, b)
		// Send the message back to the client
	}
}

func readProtoMessage(conn *websocket.Conn, message *pb.Message) error {
	messageType, p, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	// Only process binary messages
	if messageType == websocket.BinaryMessage {
		// Unmarshal the binary data into a protobuf message
		if err := proto.Unmarshal(p, message); err != nil {
			return err

		}

		return nil
	}
	return errors.New("invalid message type")
}

// handleTURNCredentials handles requests for TURN credentials
func handleTURNCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		slog.Error("Method not allowed", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	turnKeyID := os.Getenv("TURN_KEY_ID")
	turnKeyAPIToken := os.Getenv("TURN_KEY_API_TOKEN")

	if turnKeyID == "" || turnKeyAPIToken == "" {
		slog.Error("TURN credentials not configured")
		http.Error(w, "TURN credentials not configured", http.StatusInternalServerError)
		return
	}

	apiURL := fmt.Sprintf("https://rtc.live.cloudflare.com/v1/turn/keys/%s/credentials/generate", url.PathEscape(turnKeyID))

	// Create the request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(`{"ttl": 86400}`))
	if err != nil {
		slog.Error("Failed to create request", "error", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+turnKeyAPIToken)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to get TURN credentials", "error", err)
		http.Error(w, "Failed to get TURN credentials", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read and forward the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}
	slog.Debug("Forwarding response", "status", resp.StatusCode, "body", string(body))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func handleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		slog.Warn("Invalid method for signup", "method", r.Method, "ip", r.RemoteAddr)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		slog.Error("Failed to decode signup request", "error", err, "ip", r.RemoteAddr)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	slog.Debug("Processing signup request", "email", user.Email, "ip", r.RemoteAddr)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Failed to hash password", "error", err, "ip", r.RemoteAddr)
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	user.Password = string(hashedPassword)
	result := db.Create(&user)
	if result.Error != nil {
		slog.Error("Failed to create user",
			"error", result.Error,
			"email", user.Email,
			"ip", r.RemoteAddr,
		)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	slog.Info("User created successfully",
		"email", user.Email,
		"ip", r.RemoteAddr,
		"userId", user.ID,
	)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		slog.Warn("Invalid method for login", "method", r.Method, "ip", r.RemoteAddr)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		slog.Error("Failed to decode login request", "error", err, "ip", r.RemoteAddr)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	slog.Debug("Processing login request", "email", loginReq.Email, "ip", r.RemoteAddr)

	var user User
	result := db.Where("email = ?", loginReq.Email).First(&user)
	if result.Error != nil {
		slog.Warn("Login failed: user not found",
			"email", loginReq.Email,
			"ip", r.RemoteAddr,
		)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		slog.Warn("Login failed: invalid password",
			"email", loginReq.Email,
			"ip", r.RemoteAddr,
		)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Email: user.Email,
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		slog.Error("Failed to create JWT",
			"error", err,
			"email", user.Email,
			"ip", r.RemoteAddr,
		)
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	slog.Info("User logged in successfully",
		"email", user.Email,
		"ip", r.RemoteAddr,
		"userId", user.ID,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func validateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		slog.Warn("Invalid method for token validation", "method", r.Method, "ip", r.RemoteAddr)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenString := extractToken(r)
	if tokenString == "" {
		slog.Debug("Token validation failed: no token provided", "ip", r.RemoteAddr)
		json.NewEncoder(w).Encode(TokenResponse{Valid: false})
		return
	}

	claims, err := validateJWT(tokenString)
	if err != nil {
		slog.Debug("Token validation failed: invalid token",
			"error", err,
			"ip", r.RemoteAddr,
		)
		json.NewEncoder(w).Encode(TokenResponse{Valid: false})
		return
	}

	slog.Debug("Token validated successfully",
		"email", claims.Email,
		"ip", r.RemoteAddr,
	)

	json.NewEncoder(w).Encode(TokenResponse{
		Valid: true,
		Email: claims.Email,
	})
}

func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

func validateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// Middleware function to protect routes
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := extractToken(r)
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		_, err := validateJWT(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// Add new handler
func handleGenerateCamera(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user email from token
	tokenString := extractToken(r)
	claims, _ := validateJWT(tokenString)

	// Generate random code
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		slog.Error("Failed to generate random bytes", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}


	expirationTime := time.Now().Add(2 * time.Hour)
	data := &CameraAdd {
		UserID: claims.UserID,
		ServerURL: "http://100.117.177.44:8080",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		slog.Error("Failed to create JWT",
			"error", err,
		)
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"code": tokenString,
	})
}
