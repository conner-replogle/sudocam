package stream

import (
	"testing"
)

func TestCameraImplementations(t *testing.T) {
	options := CameraOptions{
		Width:          1280,
		Height:         720,
		Fps:            30,
		HorizontalFlip: true,
		VerticalFlip:   true,
		Rotation:       90,
		UseMjpeg:       true,
	}

	// Test Libcamera implementation
	libcam := &LibcameraCamera{useMjpeg: true}
	if cmd := libcam.GetCommand(); cmd != "libcamera-vid" {
		t.Errorf("Expected command 'libcamera-vid', got '%s'", cmd)
	}

	libcamArgs := libcam.GetArgs(options)
	expectedArgs := []string{
		"--width", "1280",
		"--height", "720",
		"--framerate", "30",
		"--codec", "mjpeg",
		"--hflip",
		"--vflip",
		"--rotation", "90",
	}
	for _, arg := range expectedArgs {
		if !containsArg(libcamArgs, arg) {
			t.Errorf("LibcameraCamera missing expected argument: %s", arg)
		}
	}

	// Test Raspivid implementation
	raspivid := &RaspividCamera{useMjpeg: true}
	if cmd := raspivid.GetCommand(); cmd != "raspivid" {
		t.Errorf("Expected command 'raspivid', got '%s'", cmd)
	}

	// Test FFmpeg implementation
	ffmpeg := &FFmpegCamera{useMjpeg: true}
	if cmd := ffmpeg.GetCommand(); cmd != "ffmpeg" {
		t.Errorf("Expected command 'ffmpeg', got '%s'", cmd)
	}

	ffmpegArgs := ffmpeg.GetArgs(options)
	expectedFFmpegArgs := []string{
		"-f", "v4l2",
		"-framerate", "30",
		"-video_size", "1280x720",
	}
	for _, arg := range expectedFFmpegArgs {
		if !containsArg(ffmpegArgs, arg) {
			t.Errorf("FFmpegCamera missing expected argument: %s", arg)
		}
	}
}

// Test NewCamera factory function
func TestNewCamera(t *testing.T) {
	// Test explicit camera type selection
	options := CameraOptions{CameraType: CameraTypeLibcamera}
	camera := NewCamera(options)
	if _, ok := camera.(*LibcameraCamera); !ok {
		t.Error("Expected LibcameraCamera type")
	}

	options.CameraType = CameraTypeRaspivid
	camera = NewCamera(options)
	if _, ok := camera.(*RaspividCamera); !ok {
		t.Error("Expected RaspividCamera type")
	}

	options.CameraType = CameraTypeFFmpeg
	camera = NewCamera(options)
	if _, ok := camera.(*FFmpegCamera); !ok {
		t.Error("Expected FFmpegCamera type")
	}
}

func containsArg(args []string, arg string) bool {
	for _, a := range args {
		if a == arg {
			return true
		}
	}
	return false
}
