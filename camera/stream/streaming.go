package stream

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os/exec"
	"strconv"
	"sync"
)

const (
	readBufferSize = 4096
	bufferSizeKB   = 256

	legacyCommand    = "raspivid"
	libcameraCommand = "libcamera-vid"
)

var nalSeparator = []byte{0, 0, 0, 1} //NAL break

var soiMarker = []byte{0xff, 0xd8}
var eoiMarker = []byte{0xff, 0xd9}

// CameraOptions sets the options to send to raspivid
type CameraOptions struct {
	Width               int
	Height              int
	Fps                 int
	HorizontalFlip      bool
	VerticalFlip        bool
	Rotation            int
	AutoFocus           bool
	PostProcess		    bool
	UseMjpeg            bool
	UseLibcamera        bool // Set to true to enable libcamera, otherwise use legacy raspivid stack
	AutoDetectLibCamera bool // Set to true to automatically detect if libcamera is available. If true, UseLibcamera is ignored.
}

// Video streams the video for the Raspberry Pi camera to a websocket
func Video(ctx context.Context, options CameraOptions, writer io.Writer) {
	cameraStarted := sync.Mutex{}
	slog.Info("Video: Starting camera")

	go startCamera(ctx, options, writer, &cameraStarted)
}

func startCamera(ctx context.Context, options CameraOptions, writer io.Writer, mutex *sync.Mutex) {
	mutex.Lock()
	defer mutex.Unlock()
	defer slog.Info("startCamera: Stopped camera")

	args := []string{
		"-t", "0", // Disable timeout
		"-o", "-", // Output to stdout
		"--flush", // Flush output files immediately
		"--width", strconv.Itoa(options.Width),
		"--height", strconv.Itoa(options.Height),
		"--framerate", strconv.Itoa(options.Fps),
		"-n", // Do not show a preview window
	}
	if options.UseMjpeg {
		args = append(args, "--codec", "mjpeg")
	} else {
		args = append(args, "--inline", "--profile", "baseline",
	"--low-latency",   "-b", "1000000")
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
		args = append(args, "--rotation")
		args = append(args, strconv.Itoa(options.Rotation))
	}
	if options.PostProcess {
		args = append(args, "--post-process-file", "post.json")
	}

	command := determineCameraCommand(options)

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

	if options.UseMjpeg {
		readMjpeg(writer, stdout, command)
	} else {
		readH264(writer, stdout, command)
	}
}

func readMjpeg(writer io.Writer, stdout io.ReadCloser, command string) {
	p := make([]byte, readBufferSize)
	buffer := make([]byte, bufferSizeKB*1024)
	currentPos := 0
	frameStart := -1
	minFrameSize := 1024 // Minimum reasonable JPEG size

	for {
		n, err := stdout.Read(p)
		if err != nil {
			if err == io.EOF {
				slog.Debug("startCamera: EOF", slog.String("command", command))
				return
			}
			slog.Error("startCamera: Error reading from camera; ignoring", slog.Any("error", err))
			continue
		}

		// Copy new data into buffer
		copied := copy(buffer[currentPos:], p[:n])
		endPos := currentPos + copied

		// Search entire buffer for markers
		if frameStart == -1 {
			if idx := bytes.Index(buffer[:endPos], soiMarker); idx >= 0 {
				frameStart = idx
				slog.Debug("Found SOI marker", slog.Int("position", frameStart))
			}
		}

		if frameStart >= 0 {
			// Only search from frame start to current end
			searchBuffer := buffer[frameStart:endPos]
			if idx := bytes.Index(searchBuffer, eoiMarker); idx >= 0 {
				frameEnd := frameStart + idx + len(eoiMarker)
				frameSize := frameEnd - frameStart

				if frameSize >= minFrameSize {
					slog.Debug("Found complete JPEG frame",
						slog.Int("size", frameSize),
						slog.Int("start", frameStart),
						slog.Int("end", frameEnd))

					// Write complete JPEG frame
					if _, err := writer.Write(buffer[frameStart:frameEnd]); err != nil {
						slog.Error("Error writing frame", slog.Any("error", err))
					}

					// Shift remaining data to start of buffer
					copy(buffer, buffer[frameEnd:endPos])
					currentPos = endPos - frameEnd
				} else {
					slog.Debug("Rejecting suspicious frame", slog.Int("size", frameSize))
					currentPos = 0 // Reset on invalid frame
				}
				frameStart = -1
				continue
			}
		}

		currentPos = endPos

		// Buffer safety check
		if currentPos > bufferSizeKB*1024-readBufferSize*2 {
			slog.Warn("Buffer getting full, resetting", slog.Int("pos", currentPos))
			currentPos = 0
			frameStart = -1
		}
	}
}

func readH264(writer io.Writer, stdout io.ReadCloser, command string) {
	p := make([]byte, readBufferSize)
	buffer := make([]byte, bufferSizeKB*1024)
	currentPos := 0
	NALlen := len(nalSeparator)

	for {
		n, err := stdout.Read(p)
		if err != nil {
			if err == io.EOF {
				slog.Debug("startCamera: EOF", slog.String("command", command))
				return
			}
			slog.Error("startCamera: Error reading from camera; ignoring", slog.Any("error", err))
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

func determineCameraCommand(options CameraOptions) string {
	if options.AutoDetectLibCamera {
		_, err := exec.LookPath(libcameraCommand)
		if err == nil {
			return libcameraCommand
		}
		return legacyCommand
	}

	if options.UseLibcamera {
		return libcameraCommand
	} else {
		return legacyCommand
	}
}
