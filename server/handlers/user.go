package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"server/middleware"
	"server/models"

	"gorm.io/gorm"
)


func UsersCameras(db *gorm.DB,) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		userID := r.Context().Value(middleware.ContextUserKey).(uint)
		// if userID != -1{
		// 	slog.Error("Invalid user ID", "userID", userID)
		// 	http.Error(w, "Invalid user ID", http.StatusBadRequest)
		// 	return
		// }

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