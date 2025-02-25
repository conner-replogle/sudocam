package setup

import (
	"bytes"
	"camera/config"
	"camera/stream"
	"context"
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

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

func ScanQRCode(img image.Image) *gozxing.Result {

	// prepare BinaryBitmap
	bmp, _ := gozxing.NewBinaryBitmapFromImage(img)

	// decode image
	qrReader := qrcode.NewQRCodeReader()
	result, _ := qrReader.Decode(bmp, nil)
	return result
}

func RunSetup() *config.Config {
	opt := stream.CameraOptions{
		Width:        1920,
		Height:       1080,
		Fps:          15,
		UseLibcamera: true,
		AutoFocus:    true,
		UseMjpeg:     true,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cleanup when function exits

	r, w := io.Pipe()
	stream.Video(ctx, opt, w)

	counter := 0

	for {
		img, err := jpeg.Decode(r)
		if err != nil {
			slog.Info("Error decoding image", slog.Any("error", err))
			continue
		}

		// Write the image to a file with an incrementing number
		fileName := fmt.Sprintf("image_%d.jpg", counter)
		file, err := os.Create(fileName)
		if err != nil {
			slog.Info("Error creating file", slog.Any("error", err))
			continue
		}
		defer file.Close()

		err = jpeg.Encode(file, img, nil)
		if err != nil {
			slog.Info("Error encoding image to file", slog.Any("error", err))
			continue
		}

		counter++

		result := ScanQRCode(img)
		if result != nil {
			slog.Info("QR Code found")
			result.GetText()
			claims := &jwtmsg.CameraAdd{}
			jwt.NewParser().ParseUnverified(result.GetText(),claims )
			slog.Info("QR Code", slog.Any("claims", claims))

			CameraUUID := uuid.New()


			register := &jwtmsg.RegisterCamera{
				CameraUUID:CameraUUID.String() ,
				Token: result.GetText(),

			}
			url,err := url.Parse(claims.ServerURL);
			if err != nil {
				slog.Error("Failed to parse URL", "error", err)
				continue
			}

			apiURL := fmt.Sprintf("%s/api/cameras/register", url)

			jsonData, err := json.Marshal(register)
			if err != nil {
				slog.Error("Failed to marshal JSON", "error", err)
				continue
			}


			// Create the request
			req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
			if err != nil {
				slog.Error("Failed to create request", "error", err)
				continue
			}
		
			req.Header.Set("Content-Type", "application/json")
		
			// Make the request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				slog.Error("Failed to get TURN credentials", "error", err)
			    continue
			}

			if resp.StatusCode != http.StatusCreated {
				slog.Error("Unexpected status code", "status", resp.StatusCode)
				continue
			}


			
			r.Close() // Close the pipe reader
			w.Close() // Close the pipe writer
			return &config.Config{
				Addr:       claims.ServerURL,
				CameraUuid: CameraUUID.String(),
			}
		} else {
			slog.Info("QR Code not found")
		}
	}

	
}
