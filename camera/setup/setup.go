package setup

import (
	"bufio"
	"bytes"
	"camera/config"
	"camera/stream"
	"context"
	"time"

	// "camera/stream"
	// "context"
	"encoding/json"
	"fmt"

	"image/jpeg"
	"log/slog"
	"messages/jwtmsg"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// setupWifi configures WiFi using the provided network name and password
func setupWifi(network, password string) error {
	if network == "" {
		return nil // No WiFi setup needed
	}

	slog.Info("Setting up WiFi", "network", network)

	// This is a simplified example - actual implementation will depend on your OS/platform
	cmd := exec.Command("sh", "-c", fmt.Sprintf(`nmcli device wifi connect "%s" password "%s"`, network, password))
	output, err := cmd.CombinedOutput()

	if err != nil {
		slog.Error("Failed to connect to WiFi", "error", err, "output", string(output))
		return err
	}

	slog.Info("Successfully connected to WiFi network", "network", network)
	return nil
}

// RunSetupWithQRCode runs the setup process by scanning a QR code from the camera
func RunSetupWithQRCode(debugMode bool) *config.Config {
	var mjpegServer *MJPEGServer = nil
	if debugMode {
		mjpegServer = NewMJPEGServer(8080)
		mjpegServer.Start()
	}

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure we clean up resources

	// Channel to receive the config once QR code is processed
	configCh := make(chan *config.Config, 1)

	// Channel to hold the latest JPEG frame
	latestFrame := make(chan []byte, 1) // Buffered channel to hold the latest frame

	// Create a JPEG handler that sends the latest frame to the channel
	jpegHandler := func(data []byte, duration time.Duration) bool {
		// If in debug mode, send frames to MJPEG server
		if debugMode && mjpegServer != nil {
			mjpegServer.writer1.Write(data)
		}

		select {
		case latestFrame <- data: // Send the new frame
		default: // Drop the new frame if the channel is full (already processing)
			slog.Debug("Dropping frame, processing previous")
		}
		return true // Continue processing
	}

	// Start the video stream with our handler
	stream.Video(ctx, jpegHandler, stream.JPEGMedia)

	// Launch a goroutine to process frames from the channel
	go func() {
		qrReader := qrcode.NewQRCodeReader()
		hints := map[gozxing.DecodeHintType]interface{}{}
		for {
			select {
			case frameData := <-latestFrame:
				slog.Debug("Processing JPEG frame", "size", len(frameData))
				// data,err := bimg.NewImage(frameData).Resize(640, 480)

				// if err != nil {
				// 	slog.Info("Error resizing image", "error", err)
				// 	continue // Continue processing next frame
				// }
				img, err := jpeg.Decode(bytes.NewReader(frameData))

				if err != nil {
					slog.Info("Error decoding jpeg image", "error", err)
					continue // Continue processing next frame
				}
				if debugMode {
					mjpegServer.writer2.Write(frameData)
				}

				bmp, _ := gozxing.NewBinaryBitmapFromImage(img)
			
				qrReader.Reset()
				result, err := qrReader.Decode(bmp, hints)
				if err != nil {
					slog.Info("Error decoding QR code", "error", err)
					continue // Continue processing next frame
				}

				slog.Info("QR Code found", "text_length", len(result.GetText()))
				jwtToken := result.GetText()

				// Process the JWT and send the resulting config
				if config := processJWT(jwtToken); config != nil {
					configCh <- config
					cancel() // Stop the video stream
					return   // Exit goroutine
				}

			case <-ctx.Done():
				slog.Info("Stopping frame processing due to context cancellation")
				return // Exit goroutine
			}
		}
	}()

	// Wait for either the config to be received or a timeout
	select {
	case config := <-configCh:
		slog.Info("Configuration received, stopping setup")
		return config
	case <-time.After(60*5 * time.Second):
		slog.Error("QR code scan timed out after 120 seconds")
		cancel() // Stop the video stream
		return nil
	}
}

// RunSetupWithManualInput runs the setup process by accepting a JWT token from stdin
func RunSetupWithManualInput() *config.Config {
	fmt.Println("\n==== Debug Setup Mode ====")
	fmt.Println("Paste your JWT token (then press Enter):")

	reader := bufio.NewReader(os.Stdin)
	jwtToken, err := reader.ReadString('\n')
	if err != nil {
		slog.Error("Error reading from stdin", "error", err)
		return nil
	}

	// Trim whitespace and newlines
	jwtToken = strings.TrimSpace(jwtToken)

	if jwtToken == "" {
		slog.Error("Empty JWT token provided")
		return nil
	}

	return processJWT(jwtToken)
}

// processJWT handles the common JWT processing logic for both setup methods
func processJWT(jwtToken string) *config.Config {
	// Parse the claims from the JWT
	claims := &jwtmsg.CameraAdd{}
	_, _, err := jwt.NewParser().ParseUnverified(jwtToken, claims)
	if err != nil {
		slog.Error("Failed to parse JWT", "error", err)
		return nil
	}

	slog.Info("JWT processed", slog.Any("claims", claims))

	// Setup WiFi if provided
	if claims.WifiNetwork != "" {
		if err := setupWifi(claims.WifiNetwork, claims.WifiPassword); err != nil {
			slog.Error("WiFi setup failed", "error", err)
			// Continue anyway - might be already connected or using ethernet
		}
	}

	// Generate a UUID for this camera
	cameraUUID := uuid.New()

	// Register with the server
	register := &jwtmsg.RegisterCamera{
		CameraUUID:   cameraUUID.String(),
		Token:        jwtToken,
		FriendlyName: claims.FriendlyName,
	}

	url, err := url.Parse(claims.ServerURL)
	if err != nil {
		slog.Error("Failed to parse URL", "error", err)
		return nil
	}

	apiURL := fmt.Sprintf("%s/api/cameras/register", url)

	jsonData, err := json.Marshal(register)
	if err != nil {
		slog.Error("Failed to marshal JSON", "error", err)
		return nil
	}

	// Create the request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("Failed to create request", "error", err)
		return nil
	}

	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to register camera", "error", err)
		return nil
	}

	if resp.StatusCode != http.StatusCreated {
		slog.Error("Unexpected status code", "status", resp.StatusCode)
		return nil
	}
	jwtToken = resp.Header.Get("Authorization")

	return &config.Config{
		Addr:       claims.ServerURL,
		CameraUuid: cameraUUID.String(),
		CameraName: claims.FriendlyName,
		Token:      jwtToken,
	}
}

// RunSetup chooses the appropriate setup method based on debug flag
func RunSetup(debugMode bool) *config.Config {
	if debugMode {
		return RunSetupWithManualInput()
	} else {
		return RunSetupWithQRCode(false)
	}
}

// RunSetupWithDebug runs the setup with additional debug options
func RunSetupWithDebug(visualDebug bool) *config.Config {
	return RunSetupWithQRCode(visualDebug)
}
