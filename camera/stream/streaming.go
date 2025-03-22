package stream

import (
	"context"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
)

const maxFrameSize = 1920 * 1080 / 2
const maxCachedIFrames = 5 // Store last 5 I-frames

// MediaType represents the type of media being streamed
type MediaType string

const (
	// H264Media represents an H264 video stream
	H264Media MediaType = "h264"
	// JPEGMedia represents a JPEG image stream
	JPEGMedia MediaType = "jpeg"
)

// PacketCallback is a function that will be called when packets are received
// Return false to stop processing and cancel the stream
type PacketCallback func(data []byte, duration time.Duration) bool

// connectToUnixSocket connects to a Unix socket and forwards media packets via the callback
func connectToUnixSocket(ctx context.Context, socketPath string, callback PacketCallback, mediaType MediaType) error {
	// Check if socket file exists
	if _, err := os.Stat(socketPath); err != nil {
		slog.Error("Socket file doesn't exist", "path", socketPath, "error", err)
		return err
	}

	conn, err := net.Dial("unixpacket", socketPath)
	if err != nil {
		slog.Error("Failed to connect to socket", "path", socketPath, "error", err)
		return err
	}
	defer conn.Close()

	slog.Info("Connected to video socket", "path", socketPath, "type", mediaType)

	inboundPacket := make([]byte, maxFrameSize)
	lastFrame := time.Now()

	// Create a channel for socket error handling
	errCh := make(chan error, 1)

	// Set up a goroutine to handle context cancellation
	go func() {
		<-ctx.Done()
		// Close the connection to unblock any pending Read calls
		conn.Close()
		errCh <- ctx.Err()
	}()

	for {
		// Set up a channel to signal when Read is complete
		readDone := make(chan struct{})

		// Read in a separate goroutine to handle cancellation
		var n int
		var readErr error

		go func() {
			n, readErr = conn.Read(inboundPacket)
			close(readDone)
		}()

		// Wait for either read completion or context cancellation
		select {
		case <-readDone:
			if readErr != nil {
				slog.Error("Error during read", "error", readErr)
				return readErr
			}
		case err := <-errCh:
			slog.Info("Stream cancelled", "error", err)
			return err
		}

		now := time.Now()
		sinceLastFrame := now.Sub(lastFrame)
		lastFrame = now

		// Create a copy of the packet to prevent data race when processing asynchronously
		packetCopy := make([]byte, n)
		copy(packetCopy, inboundPacket[:n])

		// Call the callback with the packet data
		// If callback returns false, stop processing
		if !callback(packetCopy, sinceLastFrame) {
			slog.Info("Stream cancelled by callback")
			return nil
		}

		// Check if context was cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue processing
		}
	}
}

// Video streams the video from a Unix socket and processes it via the provided callback
func Video(ctx context.Context, callback PacketCallback, mediaType MediaType) {
	var socketPath string

	// Select the appropriate socket based on media type
	switch mediaType {
	case H264Media:
		socketPath = "/tmp/h264_stream.sock"
	case JPEGMedia:
		socketPath = "/tmp/jpeg_stream.sock"
	default:
		socketPath = "/tmp/h264_stream.sock" // Default to H264
		mediaType = H264Media
	}

	slog.Info("Starting Unix socket client", "path", socketPath, "mediaType", mediaType)

	// Connect to the socket
	go func() {
		if err := connectToUnixSocket(ctx, socketPath, callback, mediaType); err != nil {
			// Don't reconnect if context was cancelled
			if ctx.Err() != nil {
				return
			}

			slog.Error("Error streaming from socket", "error", err, "path", socketPath)

			// Reconnection logic
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				Video(ctx, callback, mediaType)
			}
		}
	}()
}

// H264VideoHandler returns a PacketCallback that writes H264 data to a WebRTC track
func H264VideoHandler(videoTrack *webrtc.TrackLocalStaticSample) PacketCallback {
	return func(data []byte, duration time.Duration) bool {
		// Convert the raw packet data to a WebRTC media sample
		err := videoTrack.WriteSample(media.Sample{
			Data:     data,
			Duration: duration,
		})

		// Return false to stop processing if write failed
		return err == nil
	}
}

// CreateH264VideoStream is a convenience wrapper that sets up an H264 stream to a WebRTC track
func CreateH264VideoStream(ctx context.Context, videoTrack *webrtc.TrackLocalStaticSample) {
	Video(ctx, H264VideoHandler(videoTrack), H264Media)
}
