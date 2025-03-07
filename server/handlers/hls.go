package handlers

import (
	"log/slog"
	"messages/msgspb"
	"net/http"
	"server/models"
	"server/websocket"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Create a map to store response channels for camera HLS requests
var (
	hlsResponseChannels     = make(map[string]chan []byte)
	hlsResponseChannelMutex sync.Mutex
)

// RegisterHLSResponseHandler registers a handler for HLS responses from cameras
func RegisterHLSResponseHandler() {
	websocket.RegisterMessageHandler("hlsResponse", func(msg *msgspb.Message) {
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
}

// ServeHLSContent handles requests for HLS content from cameras
func ServeHLSContent(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract camera ID and file path from the URL
		// URL format: /api/cameras/{camera_id}/video/{filepath}
		path := r.URL.Path
		parts := strings.Split(path, "/")

		// Ensure path has enough segments
		if len(parts) < 5 {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}

		cameraID := parts[3]
		filePath := strings.Join(parts[5:], "/")

		slog.Info("Received HLS request", "camera_id", cameraID, "file_path", filePath)

		// Find camera in database to verify it exists
		var camera models.Camera
		if err := db.Where("camera_uuid = ?", cameraID).First(&camera).Error; err != nil {
			slog.Error("Camera not found", "camera_id", cameraID, "error", err)
			http.Error(w, "Camera not found", http.StatusNotFound)
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
		err := websocket.SendMessageToClient(cameraID, &msgspb.Message{
			From: "server",
			To:   cameraID,
			DataType: &msgspb.Message_HlsRequest{
				HlsRequest: &msgspb.HLSRequest{
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
