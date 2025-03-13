package stream

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
)

const maxFrameSize = 1920 * 1080 / 2

// serveH264Socket creates a Unix socket server that reads h264 packets and writes them to the provided writer
func serveH264Socket(socketPath string, videoTrack *webrtc.TrackLocalStaticSample) error {
	// Remove socket if it already exists
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			return fmt.Errorf("failed to remove existing socket: %w", err)
		}
	}

	// Create the socket server
	listener, err := net.Listen("unixpacket", socketPath)
	if err != nil {
		return fmt.Errorf("failed to create socket server: %w", err)
	}
	defer func() {
		listener.Close()
		os.Remove(socketPath) // Clean up socket file
	}()

	// Set permissions for the socket
	if err := os.Chmod(socketPath, 0666); err != nil {
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	slog.Info("H264 socket server started", "path", socketPath)

	// Handle connections until context is cancelled
	connection, err := listener.Accept()
	slog.Debug("Accepted connection", connection.RemoteAddr().String())

	inboundPacket := make([]byte, maxFrameSize)
	lastFrame := time.Now()
	for {
		n, err := connection.Read(inboundPacket)
		if err != nil {
			slog.Error("error during read: %v", err)
			return err
		}
		now := time.Now()
		sinceLastFrame := now.Sub(lastFrame)
		lastFrame = now
		err = videoTrack.WriteSample(media.Sample{Data: inboundPacket[:n], Duration: sinceLastFrame})
		if err != nil {
			slog.Error("error writing sample: %v", err)
		}

	}

}

// Video streams the video for the camera to a writer
func Video(videoTrack *webrtc.TrackLocalStaticSample) {
	socketPath := "/tmp/h264_stream.sock"
	slog.Info("Video: Starting h264 stream server on socket", "path", socketPath)

	go func() {
		if err := serveH264Socket(socketPath, videoTrack); err != nil {
			slog.Error("Error serving h264 socket", "error", err)
		}
	}()

	// Wait for the goroutine to complete
}
