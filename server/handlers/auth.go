package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	
	"your-project-path/middleware"
	"your-project-path/models"
)

func HandleSignup(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var user models.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// ... rest of signup logic ...
	}
}

func HandleLogin(db *gorm.DB, jwtKey []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ... login logic ...
	}
}

func ValidateToken(w http.ResponseWriter, r *http.Request) {
	// ... token validation logic ...
}
