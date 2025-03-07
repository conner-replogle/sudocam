package setup

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
)

// MJPEGServer provides a simple HTTP server that streams MJPEG content
type MJPEGServer struct {
	port     int
	writer   io.Writer
	clients  map[chan []byte]bool
	lock     sync.Mutex
	boundary string
}

// NewMJPEGServer creates a new MJPEG streaming server
func NewMJPEGServer(port int) *MJPEGServer {
	server := &MJPEGServer{
		port:     port,
		clients:  make(map[chan []byte]bool),
		boundary: "sudocamboundary",
	}

	// Create a custom writer that broadcasts to all clients
	pr, pw := io.Pipe()
	server.writer = pw

	// Process incoming frames
	go func() {
		buffer := make([]byte, 1024*1024) // 1MB buffer
		for {
			n, err := pr.Read(buffer)
			if err != nil {
				if err != io.EOF {
					slog.Error("Error reading from pipe", "error", err)
				}
				return
			}

			// Clone the data for each client
			frameData := make([]byte, n)
			copy(frameData, buffer[:n])

			// Send to all clients
			server.lock.Lock()
			for clientChan := range server.clients {
				select {
				case clientChan <- frameData:
					// Frame sent successfully
				default:
					// Client can't keep up, drop the frame
				}
			}
			server.lock.Unlock()
		}
	}()

	return server
}

// Writer returns the writer interface where MJPEG data should be written
func (s *MJPEGServer) Writer() io.Writer {
	return s.writer
}

// Start launches the HTTP server
func (s *MJPEGServer) Start() {
	// Handle the MJPEG stream
	http.HandleFunc("/stream", s.handleStream)

	// Serve a simple HTML page with the stream embedded
	http.HandleFunc("/", s.handleIndex)

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%d", s.port)
		slog.Info("Starting debug MJPEG server", "address", "http://localhost"+addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			slog.Error("Failed to start debug server", "error", err)
		}
	}()
}

// handleStream processes incoming client requests for the MJPEG stream
func (s *MJPEGServer) handleStream(w http.ResponseWriter, r *http.Request) {
	// Set headers for MJPEG streaming
	w.Header().Set("Content-Type", fmt.Sprintf("multipart/x-mixed-replace; boundary=%s", s.boundary))
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "close")
	w.Header().Set("Pragma", "no-cache")

	// Create a channel for this client
	clientChan := make(chan []byte, 10) // Buffer up to 10 frames

	// Register the new client
	s.lock.Lock()
	s.clients[clientChan] = true
	s.lock.Unlock()

	// Clean up when the client disconnects
	defer func() {
		s.lock.Lock()
		delete(s.clients, clientChan)
		s.lock.Unlock()
		close(clientChan)
	}()

	// Stream data to the client
	for {
		select {
		case frameData, ok := <-clientChan:
			if !ok {
				return // Channel closed
			}

			// Write frame with MJPEG format
			_, err := fmt.Fprintf(w, "--%s\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n",
				s.boundary, len(frameData))
			if err != nil {
				return
			}

			if _, err := w.Write(frameData); err != nil {
				return
			}

			if _, err := fmt.Fprintf(w, "\r\n"); err != nil {
				return
			}

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

		case <-r.Context().Done():
			return // Request canceled
		}
	}
}

// handleIndex serves a simple HTML page with the stream embedded
func (s *MJPEGServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>SudoCam QR Code Scanner Debug</title>
			<style>
				body { font-family: Arial, sans-serif; text-align: center; margin: 20px; }
				h1 { color: #333; }
				.container { max-width: 800px; margin: 0 auto; }
				.stream { width: 100%%; max-width: 640px; border: 1px solid #ddd; }
			</style>
		</head>
		<body>
			<div class="container">
				<h1>SudoCam QR Code Scanner Debug</h1>
				<p>This page shows the live camera feed being used for QR code scanning.</p>
				<img class="stream" src="/stream" alt="Camera Stream" />
			</div>
		</body>
		</html>
	`)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
