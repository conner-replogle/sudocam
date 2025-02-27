package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"messages/jwtmsg"
	"server/middleware"
	"server/models"
)

func HandleGenerateCamera(jwtKey []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		tokenString := middleware.ExtractToken(r)
		claims, _ := middleware.ValidateJWT(tokenString)
		serverURL := os.Getenv("SERVER_URL")
		expirationTime := time.Now().Add(2 * time.Hour)
		data := &jwtmsg.CameraAdd{
			UserID:    claims.UserID,
			ServerURL: serverURL,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, "Error creating token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"code": tokenString})
	}
}

func HandleRegisterCamera(db *gorm.DB, jwtKey []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		register := &jwtmsg.RegisterCamera{}
		if err := json.NewDecoder(r.Body).Decode(register); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		token, err := jwt.ParseWithClaims(register.Token, &jwtmsg.CameraAdd{}, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*jwtmsg.CameraAdd)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		camera := models.Camera{
			CameraUUID: register.CameraUUID,
			UserID:     claims.UserID,
		}

		if err := db.Create(&camera).Error; err != nil {
			http.Error(w, "Error registering camera", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Camera registered successfully"})
	}
}

// UpdateCameraStatus handles requests to update a camera's online status
func UpdateCameraStatus(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var statusUpdate models.CameraStatusUpdate
		if err := json.NewDecoder(r.Body).Decode(&statusUpdate); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Find the camera
		var camera models.Camera
		if err := db.Where("camera_uuid = ?", statusUpdate.CameraUUID).First(&camera).Error; err != nil {
			http.Error(w, "Camera not found", http.StatusNotFound)
			return
		}

		// Update the camera's status
		now := time.Now()
		camera.LastOnline = &now
		camera.OnlineStatus = statusUpdate.OnlineStatus

		if err := db.Save(&camera).Error; err != nil {
			http.Error(w, "Failed to update camera status", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}

// PingCamera is a simple endpoint for cameras to ping the server to maintain "online" status
func PingCamera(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cameraUUID := r.URL.Query().Get("uuid")
		if cameraUUID == "" {
			http.Error(w, "Camera UUID is required", http.StatusBadRequest)
			return
		}

		// Find the camera
		var camera models.Camera
		if err := db.Where("camera_uuid = ?", cameraUUID).First(&camera).Error; err != nil {
			http.Error(w, "Camera not found", http.StatusNotFound)
			return
		}

		// Update the camera's last online timestamp and set status to online
		now := time.Now()
		camera.LastOnline = &now
		camera.OnlineStatus = true

		if err := db.Save(&camera).Error; err != nil {
			http.Error(w, "Failed to update camera status", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}
