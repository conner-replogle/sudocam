package record

import (
	"camera/config"
	"camera/websocket"
	"context"
	"fmt"
	"io"
	"log/slog"
	"messages/msgspb"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Recorder handles recording video streams to HLS segments
type Recorder struct {
	cameraID  string
	recordDir string
	cmd       *exec.Cmd
	active    bool
	writer    io.WriteCloser
	websocket *websocket.WebsocketManager
}

// NewRecorder creates a new instance of Recorder
func NewRecorder(cfg *config.Config) *Recorder {
	// Ensure the record directory exists
	recordDir := cfg.RecordDir
	cameraID := cfg.CameraUuid
	if recordDir == "" {
		recordDir = "recordings"
	}

	fullDir := filepath.Join(recordDir, cameraID)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		slog.Error("Failed to create recording directory", "error", err)
		return nil
	}

	return &Recorder{
		cameraID:  cameraID,
		recordDir: recordDir,
		active:    false,
	}
}

// SetWebsocketManager sets the websocket manager reference
func (r *Recorder) SetWebsocketManager(ws *websocket.WebsocketManager) {
	r.websocket = ws
}

// Start begins the recording process
func (r *Recorder) Start(ctx context.Context) (io.Writer, error) {
	if r.active {
		return r.writer, nil
	}

	// Create subdirectory for current recording session
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	sessionDir := filepath.Join(r.recordDir, r.cameraID, timestamp)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	sessionDir = filepath.Join(sessionDir, "index.m3u8")

	// err := ffmpeg.Input("pipe:",
	// 		ffmpeg.KwArgs{"f": "h264",
	// 		}).
	// 		Output("recordings/", ffmpeg.KwArgs{"pix_fmt": "yuv420p"}).
	// 		OverWriteOutput().
	// 		WithInput(w).
	// 		Run()
	// Setup ffmpeg command

	// Log the command for debugging
	slog.Info("Starting ffmpeg recording", "command", r.cmd.String())

	// Get the stdin pipe for ffmpeg
	stdin, err := r.cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get ffmpeg stdin pipe: %w", err)
	}
	r.writer = stdin

	// Setup logging for ffmpeg output
	// r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr

	// Start the ffmpeg process
	if err := r.cmd.Start(); err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	r.active = true

	// Monitor the process in a goroutine
	go func() {
		if err := r.cmd.Wait(); err != nil {
			slog.Error("ffmpeg process exited with error", "error", err)
		} else {
			slog.Info("ffmpeg process completed successfully")
		}
		r.active = false
	}()

	return r.writer, nil
}

func (r *Recorder) HandleRecordRequest(msg *msgspb.RecordRequest) error {

	//list the directory under the cameraID
	files, err := os.ReadDir(filepath.Join(r.recordDir, r.cameraID))
	if err != nil {
		return fmt.Errorf("failed to list directory: %w", err)
	}

	// Create a list of video ranges
	var videoRanges []*msgspb.VideoRange
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		videoRanges = append(videoRanges, &msgspb.VideoRange{
			FileName:  file.Name(),
			StartTime: 0,
			EndTime:   0,
		})
	}

	// Send the response back

	return r.websocket.SendMessage(&msgspb.Message{
		From: r.cameraID,
		To:   "server",
		DataType: &msgspb.Message_RecordResponse{
			RecordResponse: &msgspb.RecordResponse{

				Id:      msg.Id,
				Records: videoRanges,
			},
		},
	})

}

// HandleRequest processes an HLS file request and sends the file data back
func (r *Recorder) HandleRequest(msg *msgspb.HLSRequest) error {
	if r.websocket == nil {
		return fmt.Errorf("websocket manager not set")
	}

	slog.Info("Handling HLS request", "filename", msg.FileName)

	filename := filepath.Join(r.recordDir, r.cameraID, msg.FileName)

	// For safety, normalize the path and check it's still within the recordings directory
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	recordingRoot, err := filepath.Abs(filepath.Join(r.recordDir, r.cameraID))
	if err != nil {
		return fmt.Errorf("failed to get recordings root: %w", err)
	}

	// Security check: ensure the file is inside the recordings directory
	if !isSubPath(recordingRoot, absPath) {
		slog.Warn("Security: Attempted path traversal", "requested_path", msg.FileName)
		return fmt.Errorf("invalid file path - attempted path traversal")
	}

	// Check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filename)
	}

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read all of the bytes
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	slog.Info("Sending HLS response", "size_bytes", len(fileBytes))

	// Send the response back
	err = r.websocket.SendMessage(&msgspb.Message{
		From: r.cameraID,
		To:   "server",
		DataType: &msgspb.Message_HlsResponse{
			HlsResponse: &msgspb.HLSResponse{
				FileName: msg.FileName,
				Data:     fileBytes,
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to send HLS response: %w", err)
	}

	return nil
}

// isSubPath checks if child is a subdirectory of parent
func isSubPath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel == "." || rel != ".." && !filepath.IsAbs(rel) && !strings.HasPrefix(rel, "..")
}

// Stop ends the current recording
func (r *Recorder) Stop() error {
	if !r.active {
		return nil
	}

	// Close the writer to signal EOF to ffmpeg
	if err := r.writer.Close(); err != nil {
		return fmt.Errorf("failed to close ffmpeg input: %w", err)
	}

	// Wait for the process to exit
	r.active = false
	return nil
}

// IsActive returns whether recording is currently active
func (r *Recorder) IsActive() bool {
	return r.active
}
