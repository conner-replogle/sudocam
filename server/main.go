package main

import (
	"errors"
	"log/slog"
	pb "messages/msgspb"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gorilla/websocket" // Import the gorilla/websocket library
	spa "github.com/roberthodgen/spa-server"
	"google.golang.org/protobuf/proto"
)

// Define a global upgrader for WebSocket connections
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (for development).  In production, _restrict_ this!
	},
}

// Map to store WebSocket connections
var websocketConnections = make(map[string]*websocket.Conn)
var websocketConnectionsMutex sync.Mutex

// Add these at package level
var (
	// Add channel to coordinate shutdown
	shutdownChan = make(chan struct{})
	wg           sync.WaitGroup
)

func main() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))
	// Create a channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// HTTP Server setup
	http.Handle("/", spa.SpaHandler("ui/dist", "index.html"))
	// Add the WebSocket handler
	http.HandleFunc("/ws", handleWebSocket)

	go func() {
		slog.Info("HTTP server listening on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			slog.Error("HTTP server error", "error", err)
			sigChan <- syscall.SIGTERM // Signal other goroutines to shut down
		}
	}()

	// Block until a signal is received
	<-sigChan
	slog.Info("\nReceived shutdown signal. Cleaning up...")

	// Signal all goroutines to stop
	close(shutdownChan)

	// Wait for goroutines to finish
	slog.Info("Waiting for goroutines to finish...")
	wg.Wait()
	slog.Info("All goroutines finished")

	// Close all WebSocket connections
	websocketConnectionsMutex.Lock()
	for id := range websocketConnections {
		websocketConnections[id].Close()
	}
	websocketConnectionsMutex.Unlock()
	slog.Info("Shutdown complete.")

}

// handleWebSocket handles WebSocket connections
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
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

	websocketConnectionsMutex.Lock()
	websocketConnections[id] = conn // Add the new connection to the map
	websocketConnectionsMutex.Unlock()

	defer func() {
		websocketConnectionsMutex.Lock()
		delete(websocketConnections, id) // Remove the connection when it's closed
		websocketConnectionsMutex.Unlock()
		conn.Close()
		slog.Info("WebSocket connection closed")
	}()
	slog.Info("WebSocket connection established")
	for {
		err := readProtoMessage(conn, msg)
		if err != nil {
			slog.Error("Error reading WebSocket message", "error", err)

			return
		}

		// Marshal the message back to binary
		b, err := proto.Marshal(msg)
		if err != nil {
			slog.Error("Error marshaling protobuf message", "error", err)
			continue
		}
		targ := websocketConnections[msg.To]
		if targ == nil {
			slog.Error("No target connection", "id", msg.To)
			continue
		}
		targ.WriteMessage(websocket.BinaryMessage, b)
		// Send the message back to the client
	}
}

func readProtoMessage(conn *websocket.Conn, message *pb.Message) error {
	messageType, p, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	// Only process binary messages
	if messageType == websocket.BinaryMessage {
		// Unmarshal the binary data into a protobuf message
		if err := proto.Unmarshal(p, message); err != nil {
			return err

		}

		return nil
	}
	return errors.New("invalid message type")
}
