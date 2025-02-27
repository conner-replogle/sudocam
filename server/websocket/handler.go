package websocket

import (
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	pb "messages/msgspb"

	"server/middleware"
	"server/models"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

// ConnectionType identifies the type of WebSocket connection
type ConnectionType int

const (
	TypeCamera ConnectionType = iota
	TypeUser
)

// Connection stores information about a WebSocket connection
type Connection struct {
	Conn     *websocket.Conn
	Type     ConnectionType
	UserID   uint   // The user who owns this connection
	EntityID string // Camera UUID or user identifier
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Store connections with type identification
	connections      = make(map[string]*Connection)
	connectionsMutex sync.Mutex
)

func HandleWebSocket(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("WebSocket upgrade failed", "error", err)
			return
		}

		// Read initial message to identify connection type
		msg := &pb.Message{}
		if err := readProtoMessage(conn, msg); err != nil {
			slog.Error("Failed to read initialization message", "error", err)
			conn.Close()
			return
		}

		slog.Info("received initialization message", "message", msg)
		ident := msg.GetInitalization()
		if ident == nil {
			slog.Error("Invalid initialization message")
			conn.Close()
			return
		}

		id := ident.GetId()
		isUser := ident.GetIsUser()

		// Determine connection type and authenticate
		var connection *Connection
		var connectionType ConnectionType
		var userID uint

		// If accountID is provided, this is likely a user connection
		if isUser  {
			// Validate user token from request headers
			token := ident.GetToken()
			if token == "" {
				slog.Error("Missing authorization token for user connection")
				conn.Close()
				return
			}

			// Extract the token from "Bearer <token>"
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}

			claims, err := middleware.ValidateJWT(token)
			if err != nil {
				slog.Error("Invalid authorization token")
				conn.Close()
				return
			}

			userID = claims.UserID
			connectionType = TypeUser
			slog.Info("User connected via WebSocket", "user_id", userID)

			// Store user connection
			connection = &Connection{
				Conn:     conn,
				Type:     TypeUser,
				UserID:   userID,
				EntityID: id,
			}
		} else {
			// This is a camera connection
			// Verify camera exists in database and get its owner
			var camera models.Camera
			result := db.Where("camera_uuid = ?", id).First(&camera)
			if result.Error != nil {
				slog.Error("Camera not found in database", "camera_uuid", id, "error", result.Error)
				conn.Close()
				return
			}

			connectionType = TypeCamera
			userID = camera.UserID
			slog.Info("Camera connected via WebSocket", "camera_uuid", id, "user_id", userID)

			// Update camera status to online
			if err := updateCameraStatus(db, id, true); err != nil {
				slog.Error("Failed to update camera online status", "camera_uuid", id, "error", err)
			} else {
				slog.Info("Camera now online", "camera_uuid", id, "user_id", userID)
			}

			// Store camera connection
			connection = &Connection{
				Conn:     conn,
				Type:     TypeCamera,
				UserID:   userID,
				EntityID: id,
			}
		}

		// Store connection
		connectionsMutex.Lock()
		connections[id] = connection
		connectionsMutex.Unlock()

		defer func() {
			connectionsMutex.Lock()
			delete(connections, id)
			connectionsMutex.Unlock()

			// Update camera status to offline when connection closes (if it's a camera)
			if connectionType == TypeCamera {
				if err := updateCameraStatus(db, id, false); err != nil {
					slog.Error("Failed to update camera offline status", "camera_uuid", id, "error", err)
				} else {
					slog.Info("Camera now offline", "camera_uuid", id)
				}
			}

			conn.Close()
			slog.Info("WebSocket connection closed", "type", connectionType, "id", id)
		}()

		handleMessages(conn, msg, db, connection)
	}
}

func handleMessages(conn *websocket.Conn, msg *pb.Message, db *gorm.DB, sourceConn *Connection) {
	for {
		err := readProtoMessage(conn, msg)
		if err != nil {
			slog.Error("Error reading WebSocket message", "error", err)
			continue
		}

		// Get target connection
		connectionsMutex.Lock()
		targetConn, targetExists := connections[msg.To]
		connectionsMutex.Unlock()

		// Verify target connection exists
		if !targetExists {
			slog.Error("Target connection not found", "from", msg.From, "to", msg.To)
			continue
		}

		// Verify access permissions - both connections must be owned by the same user
		if sourceConn.UserID != targetConn.UserID {
			slog.Error("Access denied: entities belong to different users",
				"from", msg.From, "to", msg.To,
				"from_owner", sourceConn.UserID, "to_owner", targetConn.UserID)
			continue
		}

		// Forward the message
		b, err := proto.Marshal(msg)
		if err != nil {
			slog.Error("Error marshaling protobuf message", "error", err)
			continue
		}

		targetConn.Conn.WriteMessage(websocket.BinaryMessage, b)
	}
}

func readProtoMessage(conn *websocket.Conn, message *pb.Message) error {
	messageType, p, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	if messageType == websocket.BinaryMessage {
		if err := proto.Unmarshal(p, message); err != nil {
			return err
		}
		return nil
	}
	return errors.New("invalid message type")
}

// updateCameraStatus updates the online status and last online time for a camera
func updateCameraStatus(db *gorm.DB, cameraUUID string, online bool) error {
	now := time.Now()
	result := db.Model(&models.Camera{}).
		Where("camera_uuid = ?", cameraUUID).
		Updates(map[string]interface{}{
			"online_status": online,
			"last_online":   now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("camera not found")
	}

	return nil
}

// verifyCameraAccess checks if a user has access to a specific camera
func verifyCameraAccess(db *gorm.DB, cameraUUID string, userID uint) bool {
	var camera models.Camera
	result := db.Where("camera_uuid = ? AND user_id = ?", cameraUUID, userID).First(&camera)
	return result.Error == nil
}
