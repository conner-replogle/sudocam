package websocket

import (
	"errors"

	"log/slog"
	"net/http"
	"sync"
	"time"

	"messages/jwtmsg"
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
	UserID   string // The user who owns this connection
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
	connections         = make(map[string]*Connection)
	connectionsMutex    sync.Mutex
	messageHandlers     = make(map[string]func(*pb.Message))
	messageHandlerMutex sync.Mutex
)

func RegisterMessageHandler(messageType string, handler func(*pb.Message)) {
	messageHandlerMutex.Lock()
	defer messageHandlerMutex.Unlock()
	messageHandlers[messageType] = handler
}

func HandleWebSocket(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("WebSocket upgrade failed", "error", err)
			return
		}
		auth_token := r.URL.Query().Get("auth")
		if auth_token == "" {
			auth_token = r.Header.Get("Authorization")
		}
		if auth_token == "" {
			slog.Error("Missing auth query/header parameter")
			conn.Close()
			return
		}
		slog.Info("WebSocket connection received", "auth_token", auth_token) //TODO DELTE
		// Validate  token from request headers

		// Extract the token from "Bearer <token>"
		if len(auth_token) > 7 && auth_token[:7] == "Bearer " {
			auth_token = auth_token[7:]
		}

		claims, err := middleware.ValidateJWT(auth_token)
		if err != nil {
			slog.Error("Invalid authorization token")
			conn.Close()
			return
		}
		id := claims.EntityID
		isUser := claims.EntityType == jwtmsg.EntityTypeUser
		// Determine connection type and authenticate
		var connection *Connection

		// If accountID is provided, this is likely a user connection
		if isUser {

			slog.Info("User connected via WebSocket", "user_id", id)

			// Store user connection
			connection = &Connection{
				Conn:     conn,
				Type:     TypeUser,
				UserID:   id,
				EntityID: id,
			}
		} else {
			// This is a camera connection
			// Verify camera exists in database and get its owner
			var camera models.Camera
			result := db.Where("id = ?", id).First(&camera)
			if result.Error != nil {

				SendProtoMessage(conn, &pb.Message{DataType: &pb.Message_Response{Response: &pb.Response{Success: false, Message: "Camera not found in database"}}})

				slog.Error("Camera not found in database", "camera_id", id, "error", result.Error)
				conn.Close()

				return
			}

			// Update camera status to online
			if err := updateCameraStatus(db, id, true); err != nil {
				slog.Error("Failed to update camera online status", "camera_id", id, "error", err)
			} else {
				slog.Info("Camera now online", "camera_id", id, "user_id", camera.UserID)
				err = SendRefreshToClient(camera.UserID)
				if err != nil {
					slog.Error("Failed to send refresh to user", "error", err)
				}

			}

			// Store camera connection
			connection = &Connection{
				Conn:     conn,
				Type:     TypeCamera,
				UserID:   camera.UserID,
				EntityID: id,
			}
		}

		// Store connection
		connectionsMutex.Lock()
		connections[id] = connection
		connectionsMutex.Unlock()
		slog.Info("WebSocket connection stored", "isUser", isUser, "id", id)

		defer func() {
			connectionsMutex.Lock()
			delete(connections, id)
			connectionsMutex.Unlock()

			// Update camera status to offline when connection closes (if it's a camera)
			if !isUser {
				if err := updateCameraStatus(db, id, false); err != nil {
					slog.Error("Failed to update camera offline status", "camera_uuid", id, "error", err)
				} else {
					slog.Info("Camera now offline", "camera_uuid", id)
					SendRefreshToClient(connection.UserID)

				}

			}

			conn.Close()
			slog.Info("WebSocket connection closed", "isUser", isUser, "id", id)
		}()

		handleMessages(conn, connection)
	}
}

func handleMessages(conn *websocket.Conn, sourceConn *Connection) {
	msg := &pb.Message{}
	for {
		err := readProtoMessage(conn, msg)
		if err != nil {
			slog.Error("Error reading WebSocket message", "error", err)
			return
		}
		if msg.To != "server" {
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
			msg.From = sourceConn.EntityID
			if msg.GetWebrtc() != nil {
				// Forward WebRTC messages to the specified recipient
				if msg.To != "" && msg.To != "server" {
					err := SendMessageToClient(msg.To, msg)
					if err != nil {
						slog.Error("Failed to forward WebRTC message", "error", err)
					}
				} else {
					slog.Error("Invalid recipient", "to", msg.To)
				}
			}
		} else {
			if msg.GetHlsResponse() != nil {
				messageHandlerMutex.Lock()
				handler := messageHandlers["hlsResponse"]
				messageHandlerMutex.Unlock()

				if handler != nil {
					handler(msg)
				} else {
					slog.Error("No handler for HLS response")
				}
			}
			if msg.GetRecordResponse() != nil {
				messageHandlerMutex.Lock()
				handler := messageHandlers["recordResponse"]
				messageHandlerMutex.Unlock()

				if handler != nil {
					handler(msg)
				} else {
					slog.Error("No handler for record response")
				}
			}

		}

	}
}

func SendProtoMessage(conn *websocket.Conn, message *pb.Message) error {
	b, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.BinaryMessage, b)
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
		Where("id = ?", cameraUUID).
		Updates(map[string]interface{}{
			"is_online": online,
			"last_seen": now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("camera not found")
	}

	return nil
}

// SendMessageToClient sends a protobuf message to a specific client
func SendMessageToClient(clientID string, message *pb.Message) error {
	connectionsMutex.Lock()
	client, exists := connections[clientID]
	connectionsMutex.Unlock()

	if !exists {
		return nil // Client not connected, silently ignore
	}

	// Marshal the protobuf message
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	// Send the binary message
	return client.Conn.WriteMessage(websocket.BinaryMessage, data)
}

func SendRefreshToClient(clientID string) error {
	return SendMessageToClient(clientID, &pb.Message{
		DataType: &pb.Message_TriggerRefresh{
			TriggerRefresh: &pb.TriggerRefresh{},
		},
	})
}
