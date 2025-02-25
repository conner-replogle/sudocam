package websocket

import (
	"errors"
	"log/slog"
	"net/http"
	"sync"

	pb "messages/msgspb"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	connections     = make(map[string]*websocket.Conn)
	connectionsMutex sync.Mutex
)

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", "error", err)
		return
	}

	msg := &pb.Message{}
	readProtoMessage(conn, msg)
	slog.Info("received message", "message", msg)
	ident := msg.GetInitalization()
	if ident == nil {
		slog.Error("Invalid Init Message")
		conn.Close()
		return
	}
	id := ident.GetId()

	connectionsMutex.Lock()
	connections[id] = conn
	connectionsMutex.Unlock()

	defer func() {
		connectionsMutex.Lock()
		delete(connections, id)
		connectionsMutex.Unlock()
		conn.Close()
		slog.Info("WebSocket connection closed")
	}()

	handleMessages(conn, msg)
}

func handleMessages(conn *websocket.Conn, msg *pb.Message) {
	for {
		err := readProtoMessage(conn, msg)
		if err != nil {
			slog.Error("Error reading WebSocket message", "error", err)
			return
		}

		b, err := proto.Marshal(msg)
		if err != nil {
			slog.Error("Error marshaling protobuf message", "error", err)
			continue
		}

		connectionsMutex.Lock()
		targ := connections[msg.To]
		connectionsMutex.Unlock()

		if targ == nil {
			slog.Error("No target connection", "id", msg.To)
			continue
		}
		targ.WriteMessage(websocket.BinaryMessage, b)
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
