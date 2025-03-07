package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"messages/jwtmsg"
	"server/middleware"
	"server/models"
)

// CameraAddRequest represents the request body for camera add endpoint
type CameraAddRequest struct {
	FriendlyName string `json:"friendly_name"`
	WifiNetwork  string `json:"wifi_network"`
	WifiPassword string `json:"wifi_password"`
}

// DeleteCameraRequest represents the request body for camera deletion endpoint
type DeleteCameraRequest struct {
	CameraUUID string `json:"camera_uuid"`
}

// UpdateCameraRequest represents the request body for camera update endpoint
type UpdateCameraRequest struct {
	CameraUUID string `json:"camera_uuid"`
	Name       string `json:"name"`
}

func HandleGenerateCamera(jwtKey []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req CameraAddRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate camera name
		if req.FriendlyName == "" {
			http.Error(w, "Camera name is required", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value(middleware.ContextUserKey).(uint)
		serverURL := os.Getenv("SERVER_URL")
		expirationTime := time.Now().Add(2 * time.Hour)

		data := &jwtmsg.CameraAdd{
			UserID:       userID,
			ServerURL:    serverURL,
			FriendlyName: req.FriendlyName,
			WifiNetwork:  req.WifiNetwork,
			WifiPassword: req.WifiPassword,
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
			slog.Error("Failed to decode request body", slog.Any("error", err))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		token, err := jwt.ParseWithClaims(register.Token, &jwtmsg.CameraAdd{}, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			slog.Error("Failed to parse JWT", slog.Any("error", err))
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*jwtmsg.CameraAdd)
		if !ok {
			slog.Error("Failed to parse claims")
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Create the camera with the friendly name from the claims
		camera := models.Camera{
			CameraUUID: register.CameraUUID,
			UserID:     claims.UserID,
			Name:       claims.FriendlyName,
			// Location:     "Default", // You can set a default location or leave it null
			IsOnline: true,
		}

		if err := db.Create(&camera).Error; err != nil {
			slog.Error("Failed to create camera", slog.Any("error", err))
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
		camera.LastSeen = &now
		camera.IsOnline = statusUpdate.OnlineStatus

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
		camera.LastSeen = &now
		camera.IsOnline = true

		if err := db.Save(&camera).Error; err != nil {
			http.Error(w, "Failed to update camera status", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}

// UpdateCamera handles requests to update camera information
// Currently supports updating the camera name
func UpdateCamera(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req UpdateCameraRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate camera UUID and name
		if req.CameraUUID == "" {
			http.Error(w, "Camera UUID is required", http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			http.Error(w, "Camera name is required", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value(middleware.ContextUserKey).(uint)

		// Find the camera
		var camera models.Camera
		if err := db.Where("camera_uuid = ?", req.CameraUUID).First(&camera).Error; err != nil {
			http.Error(w, "Camera not found", http.StatusNotFound)
			return
		}

		// Check if the user is authorized to update this camera
		if camera.UserID != userID {
			slog.Warn("Unauthorized update attempt",
				"requester_id", userID,
				"camera_owner_id", camera.UserID)
			http.Error(w, "Unauthorized to update this camera", http.StatusForbidden)
			return
		}

		// Update the camera name
		camera.Name = req.Name

		if err := db.Save(&camera).Error; err != nil {
			slog.Error("Failed to update camera", slog.Any("error", err))
			http.Error(w, "Error updating camera", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Camera updated successfully",
		})
	}
}

// DeleteCamera handles requests to delete a camera
// Only the authenticated user who owns the camera can delete it
func DeleteCamera(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req DeleteCameraRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate camera UUID
		if req.CameraUUID == "" {
			http.Error(w, "Camera UUID is required", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value(middleware.ContextUserKey).(uint)
		// Find the camera
		var camera models.Camera
		if err := db.Where("camera_uuid = ?", req.CameraUUID).First(&camera).Error; err != nil {
			http.Error(w, "Camera not found", http.StatusNotFound)
			return
		}

		// Check if the user is authorized to delete this camera
		if camera.UserID != userID {
			slog.Warn("Unauthorized deletion attempt",
				"requester_id", userID,
				"camera_owner_id", camera.UserID)
			http.Error(w, "Unauthorized to delete this camera", http.StatusForbidden)
			return
		}

		// Delete the camera
		if err := db.Delete(&camera).Error; err != nil {
			slog.Error("Failed to delete camera", slog.Any("error", err))
			http.Error(w, "Error deleting camera", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Camera deleted successfully",
		})
	}
}
