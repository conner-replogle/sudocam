package stream

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strconv"
	"sync"
)

const (
	readBufferSize = 4096
	bufferSizeKB   = 256
)

var nalSeparator = []byte{0, 0, 0, 1} //NAL break

var soiMarker = []byte{0xff, 0xd8}
var eoiMarker = []byte{0xff, 0xd9}

// CameraOptions sets the options for camera operation
type CameraOptions struct {
	Width           int
	Height          int
	Fps             int
	HorizontalFlip  bool
	VerticalFlip    bool
	Rotation        int
	BitRate         int
	AutoFocus       bool
	PostProcess     bool
	UseMjpeg        bool
	LatestFrameOnly bool       // Only send latest frame, dropping older ones if needed
	CameraType      CameraType // Type of camera to use
}

// CameraType represents the available camera backends
type CameraType string

const (
	CameraTypeAuto      CameraType = "auto"      // Auto-detect available camera
	CameraTypeRaspivid  CameraType = "raspivid"  // Legacy Raspberry Pi camera
	CameraTypeLibcamera CameraType = "libcamera" // Newer libcamera interface
	CameraTypeFFmpeg    CameraType = "ffmpeg"    // FFmpeg for various camera inputs
)

// Camera defines the interface that all camera implementations must satisfy
type Camera interface {
	// GetCommand returns the executable command for this camera type
	GetCommand() string

	// GetArgs returns the command line arguments based on the options
	GetArgs(options CameraOptions) []string

	// ReadOutput processes the output from the camera
	ReadOutput(writer io.Writer, stdout io.ReadCloser)
}

// Video streams the video for the camera to a writer
func Video(ctx context.Context, options CameraOptions, writer io.Writer) {
	cameraStarted := sync.Mutex{}
	slog.Info("Video: Starting camera")

	go startCamera(ctx, options, writer, &cameraStarted)
}

func startCamera(ctx context.Context, options CameraOptions, writer io.Writer, mutex *sync.Mutex) {
	mutex.Lock()
	defer mutex.Unlock()
	defer slog.Info("startCamera: Stopped camera")

	camera := NewCamera(options)

	command := camera.GetCommand()
	args := camera.GetArgs(options)

	cmdCtx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(cmdCtx, command, args...)
	slog.Debug("startCamera: Starting camera", slog.String("command", command), slog.Any("args", args))
	defer cmd.Wait()
	defer cancel()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		slog.Error("startCamera: Error getting stdout pipe", slog.Any("error", err))
		return
	}
	if err := cmd.Start(); err != nil {
		slog.Error("startCamera: Error starting camera", slog.Any("error", err))
		return
	}
	slog.Debug("startCamera: Started camera", slog.String("command", command), slog.Any("args", args))

	go func() {
		<-ctx.Done()
		cancel()
	}()

	camera.ReadOutput(writer, stdout)
}

// NewCamera creates a camera implementation based on the provided options
func NewCamera(options CameraOptions) Camera {
	if options.CameraType == CameraTypeAuto {
		// Auto-detect available camera
		if _, err := exec.LookPath("libcamera-vid"); err == nil {
			return &LibcameraCamera{options.UseMjpeg}
		} else if _, err := exec.LookPath("raspivid"); err == nil {
			return &RaspividCamera{options.UseMjpeg}
		} else if _, err := exec.LookPath("ffmpeg"); err == nil {
			return &FFmpegCamera{options.UseMjpeg}
		}
		// Default to libcamera even if not found - will show the appropriate error when trying to start
		return &FFmpegCamera{options.UseMjpeg}
	}

	switch options.CameraType {
	case CameraTypeLibcamera:
		return &LibcameraCamera{options.UseMjpeg}
	case CameraTypeFFmpeg:
		return &FFmpegCamera{options.UseMjpeg}
	case CameraTypeRaspivid:
		fallthrough
	default:
		return &FFmpegCamera{options.UseMjpeg}
	}
}

// LibcameraCamera implements Camera for the libcamera-vid command
type LibcameraCamera struct {
	useMjpeg bool
}

func (c *LibcameraCamera) GetCommand() string {
	return "libcamera-vid"
}

func (c *LibcameraCamera) GetArgs(options CameraOptions) []string {
	args := []string{
		"-t", "0", // Disable timeout
		"-o", "-", // Output to stdout
		"--width", strconv.Itoa(options.Width),
		"--height", strconv.Itoa(options.Height),
		"--framerate", strconv.Itoa(options.Fps),
		"-n", // No preview
	}

	if c.useMjpeg {
		args = append(args, "--codec", "mjpeg")
	} else {
		args = append(args, "--inline", "--profile", "baseline",
			"--low-latency")
	}

	if options.BitRate > 0 {
		args = append(args, "-b", strconv.Itoa(options.BitRate))
	} else {
		args = append(args, "-b", "5000000")
	}
	if options.AutoFocus {
		args = append(args, "--autofocus-mode", "continuous")
	}
	if options.HorizontalFlip {
		args = append(args, "--hflip")
	}
	if options.VerticalFlip {
		args = append(args, "--vflip")
	}
	if options.Rotation != 0 {
		args = append(args, "--rotation", strconv.Itoa(options.Rotation))
	}
	if options.PostProcess {
		args = append(args, "--post-process-file", "post.json")
	}

	return args
}

func (c *LibcameraCamera) ReadOutput(writer io.Writer, stdout io.ReadCloser) {
	if c.useMjpeg {
		readMjpeg(writer, stdout)
	} else {
		readH264(writer, stdout)
	}
}

// RaspividCamera implements Camera for the legacy raspivid command
type RaspividCamera struct {
	useMjpeg bool
}

func (c *RaspividCamera) GetCommand() string {
	return "raspivid"
}

func (c *RaspividCamera) GetArgs(options CameraOptions) []string {
	args := []string{
		"-t", "0", // Disable timeout
		"-o", "-", // Output to stdout
		"-w", strconv.Itoa(options.Width),
		"-h", strconv.Itoa(options.Height),
		"-fps", strconv.Itoa(options.Fps),
		"-n", // No preview
	}

	if c.useMjpeg {
		args = append(args, "-cd", "MJPEG")
	} else {
		args = append(args, "-pf", "baseline", "-ih")
	}

	if options.HorizontalFlip {
		args = append(args, "-hf")
	}
	if options.VerticalFlip {
		args = append(args, "-vf")
	}
	if options.Rotation != 0 {
		args = append(args, "-rot", strconv.Itoa(options.Rotation))
	}

	return args
}

func (c *RaspividCamera) ReadOutput(writer io.Writer, stdout io.ReadCloser) {
	if c.useMjpeg {
		readMjpeg(writer, stdout)
	} else {
		readH264(writer, stdout)
	}
}

// FFmpegCamera implements Camera for FFmpeg
type FFmpegCamera struct {
	useMjpeg bool
}

func (c *FFmpegCamera) GetCommand() string {
	return "ffmpeg"
}

func (c *FFmpegCamera) GetArgs(options CameraOptions) []string {
	// Use test source instead of real camera
	args := []string{
		"-f", "lavfi",
		"-i", "testsrc=size=" + fmt.Sprintf("%dx%d", options.Width, options.Height) +
			":rate=" + strconv.Itoa(options.Fps) +
			":duration=3600", // Generate for 1 hour (can adjust as needed)
	}

	// Add filters for rotation and flipping if needed
	filterComplex := ""
	if options.VerticalFlip && options.HorizontalFlip {
		filterComplex = "transpose=2,transpose=2"
	} else if options.VerticalFlip {
		filterComplex = "vflip"
	} else if options.HorizontalFlip {
		filterComplex = "hflip"
	}

	// Handle rotation
	if options.Rotation != 0 {
		if filterComplex != "" {
			filterComplex += ","
		}
		switch options.Rotation {
		case 90:
			filterComplex += "transpose=1"
		case 180:
			filterComplex += "transpose=2,transpose=2"
		case 270:
			filterComplex += "transpose=2"
		}
	}

	if filterComplex != "" {
		args = append(args, "-vf", filterComplex)
	}

	// Output format
	if c.useMjpeg {
		args = append(args,
			"-f", "mjpeg",
			"-q:v", "5") // Quality setting for MJPEG
	} else {
		args = append(args,
			"-c:v", "libx264",
			"-preset", "ultrafast",
			// "-tune", "zerolatency",
			"-force_key_frames", "expr:gte(t,n_forced*1)",
			"-pix_fmt", "yuv420p",
			"-f", "h264")

		// Add bitrate if specified
		if options.BitRate > 0 {
			args = append(args, "-b:v", strconv.Itoa(options.BitRate))
		}
	}

	// Silence ffmpeg output except errors
	args = append(args, "-loglevel", "error")

	// Output to stdout
	args = append(args, "-")

	return args
}

func (c *FFmpegCamera) ReadOutput(writer io.Writer, stdout io.ReadCloser) {
	if c.useMjpeg {
		readMjpeg(writer, stdout)
	} else {
		readH264(writer, stdout)
	}
}

func readMjpeg(writer io.Writer, stdout io.ReadCloser) {
	const (
		minFrameSize = 1024 // Minimum reasonable JPEG size
		bufSize      = bufferSizeKB * 1024
	)

	buffer := make([]byte, bufSize)
	p := make([]byte, readBufferSize)

	// Ring buffer management
	head := 0
	tail := 0
	size := 0
	frameStart := -1
	latestFrameOnly := true

	// For latest-frame-only mode
	var pendingFrames [][]byte
	if latestFrameOnly {
		pendingFrames = make([][]byte, 0, 2) // Pre-allocate small capacity
	}

	// Function to get available data size
	available := func() int {
		return size
	}

	// Function to add data to the ring buffer
	push := func(data []byte) int {
		n := len(data)
		if n > bufSize-size {
			n = bufSize - size
		}

		// Copy data in one or two operations
		if head+n <= bufSize {
			copy(buffer[head:], data[:n])
		} else {
			chunk1 := bufSize - head
			copy(buffer[head:], data[:chunk1])
			copy(buffer[0:], data[chunk1:n])
		}

		head = (head + n) % bufSize
		size += n
		return n
	}

	// Function to discard data from the ring buffer
	discard := func(n int) {
		if n > size {
			n = size
		}
		tail = (tail + n) % bufSize
		size -= n

		// Reset frame start tracking after discarding data
		if frameStart >= 0 {
			frameStart -= n
			if frameStart < 0 {
				frameStart = -1
			}
		}
	}

	// Function to peek at data in the ring buffer
	peekAt := func(pos int, pattern []byte) bool {
		if pos < 0 || pos+len(pattern) > size {
			return false
		}

		realPos := (tail + pos) % bufSize
		for i := 0; i < len(pattern); i++ {
			bytePos := (realPos + i) % bufSize
			if buffer[bytePos] != pattern[i] {
				return false
			}
		}
		return true
	}

	// Function to search for a pattern in the ring buffer
	search := func(start int, end int, pattern []byte) int {
		if start < 0 {
			start = 0
		}
		if end > size {
			end = size
		}
		if start >= end {
			return -1
		}

		patLen := len(pattern)
		for i := start; i <= end-patLen; i++ {
			if peekAt(i, pattern) {
				return i
			}
		}
		return -1
	}

	// Function to extract data from the ring buffer
	extract := func(start int, end int) []byte {
		if start < 0 || end > size || start >= end {
			return nil
		}

		result := make([]byte, end-start)
		realStart := (tail + start) % bufSize

		if realStart+len(result) <= bufSize {
			copy(result, buffer[realStart:realStart+len(result)])
		} else {
			firstChunk := bufSize - realStart
			copy(result[:firstChunk], buffer[realStart:])
			copy(result[firstChunk:], buffer[:len(result)-firstChunk])
		}

		return result
	}

	// Latest-frame only: flush writes the latest frame and discards older ones
	flushLatestFrame := func() {
		if len(pendingFrames) == 0 {
			return
		}

		// Write only the most recent frame
		latestFrame := pendingFrames[len(pendingFrames)-1]
		if _, err := writer.Write(latestFrame); err != nil {
			slog.Error("Error writing latest frame", slog.Any("error", err))
		}

		// Log how many frames were skipped
		if len(pendingFrames) > 1 {
			slog.Debug("Skipped frames for latest-only mode",
				slog.Int("skipped", len(pendingFrames)-1))
		}

		// Clear the pending frames
		pendingFrames = pendingFrames[:0]
	}

	for {
		n, err := stdout.Read(p)
		if err != nil {
			if err == io.EOF {
				slog.Debug("readMjpeg: EOF")
				return
			}
			slog.Error("readMjpeg: Error reading from camera; ignoring", slog.Any("error", err))
			continue
		}

		if n == 0 {
			continue
		}

		// Add data to buffer
		pushed := push(p[:n])
		if pushed < n {
			slog.Warn("Buffer overflow, some data discarded", slog.Int("discarded", n-pushed))
		}

		// Process frames while we have enough data
		for available() > 0 {
			// Find start of frame if needed
			if frameStart == -1 {
				frameStart = search(0, available(), soiMarker)
				if frameStart == -1 {
					// No start marker found, discard all but last potential marker
					discard(available() - 1)
					break
				}
			}

			// Look for end marker after the start marker
			endIdx := search(frameStart+2, available(), eoiMarker)
			if endIdx == -1 {
				// End marker not found yet
				// Keep at least frameStart bytes in buffer, discard the rest
				if frameStart > 0 {
					discard(frameStart)
					frameStart = 0
				}
				break
			}

			// We have a complete frame
			frameEnd := endIdx + len(eoiMarker)
			frameSize := frameEnd - frameStart

			if frameSize >= minFrameSize {
				// Extract the frame
				frame := extract(frameStart, frameEnd)

				if latestFrameOnly {
					// Store frame for later processing
					pendingFrames = append(pendingFrames, frame)
				} else {
					// Normal mode: write frame immediately
					if _, err := writer.Write(frame); err != nil {
						slog.Error("Error writing frame", slog.Any("error", err))
					}
				}
			} else {
				slog.Debug("Rejecting suspicious frame", slog.Int("size", frameSize))
			}

			// Discard the processed frame
			discard(frameEnd)
			frameStart = -1
		}

		// For latest-frame-only mode, flush pending frames after processing all available data
		if latestFrameOnly && len(pendingFrames) > 0 {
			flushLatestFrame()
		}

		// Safety check - if buffer is getting too full without finding frames, reset
		if available() > bufSize*3/4 {
			slog.Warn("Buffer getting full without complete frames, resetting",
				slog.Int("available", available()))
			discard(available())
			frameStart = -1
		}
	}
}

func readH264(writer io.Writer, stdout io.ReadCloser) {
	p := make([]byte, readBufferSize)
	buffer := make([]byte, bufferSizeKB*1024)
	currentPos := 0
	NALlen := len(nalSeparator)

	for {
		n, err := stdout.Read(p)
		if err != nil {
			if err == io.EOF {
				slog.Debug("readH264: EOF")
				return
			}
			slog.Error("readH264: Error reading from camera; ignoring", slog.Any("error", err))
			continue
		}

		copied := copy(buffer[currentPos:], p[:n])
		startPosSearch := currentPos - NALlen
		endPos := currentPos + copied

		if startPosSearch < 0 {
			startPosSearch = 0
		}
		nalIndex := bytes.Index(buffer[startPosSearch:endPos], nalSeparator)

		currentPos = endPos
		if nalIndex > 0 {
			nalIndex += startPosSearch

			// Broadcast before the NAL
			broadcast := make([]byte, nalIndex)
			copy(broadcast, buffer)
			writer.Write(broadcast)

			// Shift
			copy(buffer, buffer[nalIndex:currentPos])
			currentPos = currentPos - nalIndex
		}
	}
}
