package websocket

import (
	"encoding/json"
	"log/slog"
	pb "messages/msgspb"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type ThreadSafeWriter struct {
	conn  *websocket.Conn
	mutex sync.Mutex
}

func newThreadSafeWriter(conn *websocket.Conn) *ThreadSafeWriter {
	return &ThreadSafeWriter{
		conn: conn,
	}
}

func (w *ThreadSafeWriter) writeMessage(messageType int, data []byte) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.conn.WriteMessage(messageType, data)
}

type WebsocketManager struct {
	Writer         *ThreadSafeWriter
	conn           *websocket.Conn
	CameraUuid     string
	ServerUrl      *url.URL
	WSServerURL    *url.URL
	connected      bool
	reconnecting   bool
	reconnectMutex sync.Mutex
	stopReconnect  chan struct{}
}

func NewWebsocketManager(u *url.URL, camera_uuid string) *WebsocketManager {
	websocketUrl, _ := u.Parse(u.String())
	websocketUrl.Scheme = "ws"
	if u.Scheme == "https" {
		websocketUrl.Scheme = "wss"
	}
	websocketUrl.Path = "/ws"

	manager := &WebsocketManager{
		CameraUuid:    camera_uuid,
		ServerUrl:     u,
		WSServerURL:   websocketUrl,
		connected:     false,
		reconnecting:  false,
		stopReconnect: make(chan struct{}),
	}

	// Establish initial connection
	manager.connect()
	return manager
}

func (manager *WebsocketManager) connect() bool {
	slog.Info("connecting to websocket", "url", manager.WSServerURL.String())

	c, _, err := websocket.DefaultDialer.Dial(manager.WSServerURL.String(), nil)
	if err != nil {
		slog.Error("dial failed:", "error", err)
		return false
	}

	writer := newThreadSafeWriter(c)
	manager.Writer = writer
	manager.conn = c
	manager.connected = true

	// Send initialization message
	err = SendProtoMessage(writer, &pb.Message{
		From: manager.CameraUuid,
		To:   "server",
		DataType: &pb.Message_Initalization{
			Initalization: &pb.Initalization{
				Id: manager.CameraUuid,
			},
		},
	})

	if err != nil {
		slog.Error("failed to send initialization message:", "error", err)
		manager.connected = false
		c.Close()
		return false
	}

	slog.Info("websocket connection established")
	return true
}

func (manager *WebsocketManager) startReconnectLoop() {
	manager.reconnectMutex.Lock()
	if manager.reconnecting {
		manager.reconnectMutex.Unlock()
		return
	}
	manager.reconnecting = true
	manager.reconnectMutex.Unlock()

	go func() {
		backoff := 1 * time.Second
		maxBackoff := 30 * time.Second

		for {
			select {
			case <-manager.stopReconnect:
				manager.reconnectMutex.Lock()
				manager.reconnecting = false
				manager.reconnectMutex.Unlock()
				return
			default:
				if manager.connected {
					manager.reconnectMutex.Lock()
					manager.reconnecting = false
					manager.reconnectMutex.Unlock()
					return
				}

				slog.Info("attempting to reconnect", "backoff", backoff.String())
				if manager.connect() {
					manager.reconnectMutex.Lock()
					manager.reconnecting = false
					manager.reconnectMutex.Unlock()
					return
				}

				// Wait before next attempt with exponential backoff
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
		}
	}()
}

func (manager *WebsocketManager) ensureConnected() bool {
	if manager.connected && manager.conn != nil {
		return true
	}

	// If not already trying to reconnect, start reconnection loop
	if !manager.reconnecting {
		manager.startReconnectLoop()
	}

	return false
}

func (manager *WebsocketManager) Close() {
	// Signal reconnect loop to stop if it's running
	manager.reconnectMutex.Lock()
	if manager.reconnecting {
		close(manager.stopReconnect)
	}
	manager.reconnectMutex.Unlock()

	if manager.conn != nil {
		manager.conn.Close()
		manager.connected = false
	}
}

func (manager *WebsocketManager) ReadMessage() (*pb.Message, error) {
	if !manager.ensureConnected() {
		return nil, websocket.ErrCloseSent
	}

	_, data, err := manager.conn.ReadMessage()
	if err != nil {
		slog.Error("read message error:", "error", err)
		manager.connected = false
		manager.startReconnectLoop()
		return nil, err
	}

	msg := &pb.Message{}
	err = proto.Unmarshal(data, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (manager *WebsocketManager) SendWebRTCMessage(payload any, to string) error {
	if !manager.ensureConnected() {
		return websocket.ErrCloseSent
	}

	message := &pb.Message{
		From: manager.CameraUuid,
		To:   to,
		DataType: &pb.Message_Webrtc{
			Webrtc: &pb.Webrtc{
				Data: func() string {
					data, err := json.Marshal(payload)
					if err != nil {
						panic(err)
					}
					return string(data)
				}(),
			},
		},
	}

	err := SendProtoMessage(manager.Writer, message)
	if err != nil {
		slog.Error("failed to send WebRTC message:", "error", err)
		manager.connected = false
		manager.startReconnectLoop()
	}
	return err
}

func SendProtoMessage(writer *ThreadSafeWriter, message *pb.Message) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	return writer.writeMessage(websocket.BinaryMessage, data)
}
