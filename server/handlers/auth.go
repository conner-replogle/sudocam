package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"server/middleware"
	"server/models"

	"gorm.io/gorm"
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

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		user.Password = string(hashedPassword)
		if err := db.Create(&user).Error; err != nil {
			http.Error(w, "Error creating user", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
	}
}

// HandleLogin processes user login and sets JWT as an HTTP-only cookie
func HandleLogin(db *gorm.DB, jwtKey []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var creds models.LoginRequest
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		var user models.User
		if err := db.Where("email = ?", creds.Email).First(&user).Error; err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Create a JWT token
		expirationTime := time.Now().Add(24 * time.Hour)
		claims := &middleware.Claims{
			UserID: user.ID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set the token as an HTTP-only cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    tokenString,
			Expires:  expirationTime,
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
			// Secure: true, // Enable in production with HTTPS
		})

		// Also return the token in the response for backward compatibility
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token": tokenString,
		})
	}
}
// HandleLogout clears the authentication cookie
func HandleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Clear the auth cookie by setting an expired cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    "",
			Expires:  time.Now().Add(-1 * time.Hour), // Set to past time
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
			// Secure: true, // Enable in production with HTTPS
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}
}
func ValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenString := middleware.ExtractToken(r)
	if tokenString == "" {
		json.NewEncoder(w).Encode(models.TokenResponse{Valid: false})
		return
	}

	claims, err := middleware.ValidateJWT(tokenString)
	if err != nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    "",
			Expires:  time.Now().Add(-1 * time.Hour), // Set to past time
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
			// Secure: true, // Enable in production with HTTPS
		})
		json.NewEncoder(w).Encode(models.TokenResponse{Valid: false})
		return
	}


	json.NewEncoder(w).Encode(models.TokenResponse{
		Valid: true,
		Email: claims.Email,
	})
}
