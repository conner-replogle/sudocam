package handlers

import (
	"encoding/json"
	"log/slog"
	pb "messages/msgspb"
	"net/http"
	"server/middleware"
	"server/models"
	"server/websocket"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Create a map to store response channels for camera HLS requests
var (
	hlsResponseChannels     = make(map[string]chan []byte)
	hlsResponseChannelMutex sync.Mutex

	recordResponseChannels     = make(map[string]chan pb.RecordResponse)
	recordResponseChannelMutex sync.Mutex
)

// RegisterHLSResponseHandler registers a handler for HLS responses from cameras
func RegisterHLSResponseHandler() {
	websocket.RegisterMessageHandler("hlsResponse", func(msg *pb.Message) {
		if msg.GetHlsResponse() == nil {
			return
		}

		cameraID := msg.From
		hlsResponseChannelMutex.Lock()
		defer hlsResponseChannelMutex.Unlock()

		// Check if there's a channel waiting for this response
		if ch, exists := hlsResponseChannels[cameraID+msg.GetHlsResponse().FileName]; exists {
			// Send the data to the channel
			select {
			case ch <- msg.GetHlsResponse().Data:
				// Data sent successfully
			default:
				// Channel not ready, this shouldn't happen
				slog.Warn("HLS response channel not ready for camera", "camera_id", cameraID)
			}
		} else {
			slog.Warn("Received HLS response but no channel exists", "camera_id", cameraID)
		}
	})

	websocket.RegisterMessageHandler("recordResponse", func(msg *pb.Message) {
		if msg.GetRecordResponse() == nil {
			return
		}

		cameraID := msg.From
		recordResponseChannelMutex.Lock()
		defer recordResponseChannelMutex.Unlock()

		// Check if there's a channel waiting for this response
		if ch, exists := recordResponseChannels[cameraID+strconv.Itoa(int(msg.GetRecordResponse().Id))]; exists {
			// Send the data to the channel
			select {
			case ch <- *msg.GetRecordResponse():
				// Data sent successfully
			default:
				// Channel not ready, this shouldn't happen
				slog.Warn("Record response channel not ready for camera", "camera_id", cameraID)
			}
		} else {
			slog.Warn("Received Record response but no channel exists", "camera_id", cameraID)
		}
	})
}

// ServeHLSContent handles requests for HLS content from cameras
func ServeHLSContent(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract camera ID and file path from the URL
		// URL format: /api/cameras/{camera_id}/video/{filepath}
		cameraID := r.PathValue("id")
		filePath := r.PathValue("filepath")

		slog.Info("Received HLS request", "camera_id", cameraID, "file_path", filePath)
		userID := r.Context().Value(middleware.ContextUserKey).(string)
		// Find camera in database to verify it exists
		var camera models.Camera
		if err := db.Where("id = ?", cameraID).First(&camera).Error; err != nil {
			slog.Error("Camera not found", "camera_id", cameraID, "error", err)
			http.Error(w, "Camera not found", http.StatusNotFound)
			return
		}

		if camera.UserID != userID {
			slog.Warn("Unauthorized access attempt", "user_id", userID, "camera_owner_id", camera.UserID)
			http.Error(w, "Unauthorized access", http.StatusForbidden)
			return
		}

		// Check if camera is online
		if !camera.IsOnline {
			slog.Error("Camera is offline", "camera_id", cameraID)
			http.Error(w, "Camera is offline", http.StatusServiceUnavailable)
			return
		}

		// Create a channel to receive the HLS response
		responseChan := make(chan []byte, 1)
		channelID := cameraID + filePath
		// Register the channel to receive the response
		hlsResponseChannelMutex.Lock()
		hlsResponseChannels[channelID] = responseChan
		hlsResponseChannelMutex.Unlock()

		// Clean up the channel when done
		defer func() {
			hlsResponseChannelMutex.Lock()
			delete(hlsResponseChannels, channelID)
			hlsResponseChannelMutex.Unlock()
			close(responseChan)
		}()

		// Send the request to the camera
		err := websocket.SendMessageToClient(cameraID, &pb.Message{
			From: "server",
			To:   cameraID,
			DataType: &pb.Message_HlsRequest{
				HlsRequest: &pb.HLSRequest{
					FileName: filePath,
				},
			},
		})

		if err != nil {
			slog.Error("Failed to send HLS request to camera", "camera_id", cameraID, "error", err)
			http.Error(w, "Failed to communicate with camera", http.StatusInternalServerError)
			return
		}

		// Wait for the response with a timeout
		select {
		case data := <-responseChan:
			// Set appropriate content type based on file extension
			contentType := "video/mp2t"
			if strings.HasSuffix(filePath, ".m3u8") {
				contentType = "application/vnd.apple.mpegurl"
			}
			w.Header().Set("Content-Type", contentType)

			// Set CORS headers to allow cross-origin requests
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

			w.Write(data)

		case <-time.After(10 * time.Second):
			// Timeout waiting for camera response
			slog.Error("Timeout waiting for HLS response from camera", "camera_id", cameraID)
			http.Error(w, "Timeout waiting for camera response", http.StatusGatewayTimeout)
		}
	}
}

// ServeHLSContent handles requests for HLS content from cameras
func VideoList(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract camera ID and file path from the URL
		// URL format: /api/cameras/{camera_id}/list
		cameraID := r.PathValue("id")

		slog.Info("Received List request", "camera_id", cameraID)
		userID := r.Context().Value(middleware.ContextUserKey).(string)
		// Find camera in database to verify it exists
		var camera models.Camera
		if err := db.Where("id = ?", cameraID).First(&camera).Error; err != nil {
			slog.Error("Camera not found", "camera_id", cameraID, "error", err)
			http.Error(w, "Camera not found", http.StatusNotFound)
			return
		}

		if camera.UserID != userID {
			slog.Warn("Unauthorized access attempt", "user_id", userID, "camera_owner_id", camera.UserID)
			http.Error(w, "Unauthorized access", http.StatusForbidden)
			return
		}

		// Check if camera is online
		if !camera.IsOnline {
			slog.Error("Camera is offline", "camera_id", cameraID)
			http.Error(w, "Camera is offline", http.StatusServiceUnavailable)
			return
		}
		request := &pb.RecordRequest{}
		json.NewDecoder(r.Body).Decode(&request)

		// Create a channel to receive the Record response
		responseChan := make(chan pb.RecordResponse, 1)
		id := time.Now().Unix()
		request.Id = id

		channelID := cameraID + strconv.Itoa(int(id))
		// Register the channel to receive the response
		recordResponseChannelMutex.Lock()
		recordResponseChannels[channelID] = responseChan
		recordResponseChannelMutex.Unlock()

		// Clean up the channel when done
		defer func() {
			recordResponseChannelMutex.Lock()
			delete(recordResponseChannels, channelID)
			recordResponseChannelMutex.Unlock()
			close(responseChan)
		}()

		// Send the request to the camera
		err := websocket.SendMessageToClient(cameraID, &pb.Message{
			From: "server",
			To:   cameraID,
			DataType: &pb.Message_RecordRequest{
				RecordRequest: request,
			},
		})

		if err != nil {
			slog.Error("Failed to send Record request to camera", "camera_id", cameraID, "error", err)
			http.Error(w, "Failed to communicate with camera", http.StatusInternalServerError)
			return
		}

		// Wait for the response with a timeout
		select {
		case data := <-responseChan:
			// Set appropriate content type based on file extension
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(data)

		case <-time.After(10 * time.Second):
			// Timeout waiting for camera response
			slog.Error("Timeout waiting for HLS response from camera", "camera_id", cameraID)
			http.Error(w, "Timeout waiting for camera response", http.StatusGatewayTimeout)
		}
	}
}
