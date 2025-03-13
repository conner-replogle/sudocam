package setup

import (
	"bufio"
	"bytes"
	"camera/config"
	// "camera/stream"
	// "context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
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

// ScanQRCode attempts to decode a QR code from the provided image
func ScanQRCode(img image.Image) *gozxing.Result {
	// prepare BinaryBitmap
	bmp, _ := gozxing.NewBinaryBitmapFromImage(img)

	// decode image
	qrReader := qrcode.NewQRCodeReader()
	result, _ := qrReader.Decode(bmp, nil)
	return result
}

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
	// opt := stream.CameraOptions{
	// 	Width:      1280,
	// 	Height:     720,
	// 	Fps:        30,
	// 	AutoFocus:  true,
	// 	UseMjpeg:   true,
	// 	CameraType: stream.CameraTypeAuto,
	// }

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel() // Ensure cleanup when function exits

	r, w := io.Pipe()

	var debugServer *MJPEGServer
	// var streamWriter io.Writer = w

	// Set up debug HTTP server if debug mode is enabled
	if debugMode {
		debugServer = NewMJPEGServer(8080)
		debugServer.Start()

		// Create a multi-writer to send data to both the QR scanner and the debug server
		// streamWriter = io.MultiWriter(w, debugServer.Writer())
		slog.Info("Debug mode enabled - view stream at http://localhost:8080")
	}

	// stream.Video(ctx, opt, streamWriter)

	for {
		img, err := jpeg.Decode(r)
		if err != nil {
			slog.Info("Error decoding image", slog.Any("error", err))
			continue
		}

		result := ScanQRCode(img)
		if result != nil {
			slog.Info("QR Code found")
			jwtToken := result.GetText()

			r.Close() // Close the pipe reader
			w.Close() // Close the pipe writer

			return processJWT(jwtToken)
		}
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
