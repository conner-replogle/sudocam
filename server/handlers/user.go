package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"server/middleware"
	"server/models"

	"gorm.io/gorm"
)

// SafeUserResponse represents a user without sensitive fields
type SafeUserResponse struct {
	ID       uint   `json:"id"`
	Name string `json:"name"`
	Email    string `json:"email"`
	// Add other non-sensitive fields as needed
}

func Me(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := r.Context().Value(middleware.ContextUserKey).(uint)
		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			slog.Error("Error fetching user", "error", err)
			http.Error(w, "Error fetching user", http.StatusInternalServerError)
			return
		}

		// Create a safe response without the password hash
		safeUser := SafeUserResponse{
			ID:       user.ID,
			Name: user.Name,
			Email:    user.Email,
			// Add other fields as needed
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(safeUser)
	}
}

func UsersCameras(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {

			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := r.Context().Value(middleware.ContextUserKey).(uint)

		var cameras []models.Camera
		if err := db.Where("user_id = ?", userID).Find(&cameras).Error; err != nil {
			slog.Error("Error fetching cameras", "error", err)
			http.Error(w, "Error fetching cameras", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cameras)
	}
}
