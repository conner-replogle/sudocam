// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

//go:build !js
// +build !js

// play-from-disk demonstrates how to send video and/or audio to your browser from files saved to disk.
package main

import (
	"camera/config"
	"camera/record"
	"camera/setup"
	"camera/stream"
	"camera/webrtc"
	"camera/websocket"
	"context"
	"flag"
	"io"
	"log/slog"
	"messages/msgspb"
	"net/url"
	"os"
	"time"

	"google.golang.org/protobuf/proto"
)

var debugMode = flag.Bool("debug", false, "Run in debug mode with random UUID and name")
var debugSetupMode = flag.Bool("debug-setup", false, "Stream QR Code Video to http://localhost:8080")
var configFile = flag.String("config", "config.json", "Directory to save config file")

func main() { //nolint
	// Set up logging
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	// Parse command-line flags
	flag.Parse()

	var cfg *config.Config
	var err error

	// Normal mode - load from config file or run setup
	if !*debugMode {
		cfg, err = config.LoadConfig(*configFile)
	}
	if err != nil || *debugMode {
		slog.Info("No config file found")
		if *debugSetupMode {
			cfg = setup.RunSetupWithDebug(true)
		} else {
			cfg = setup.RunSetup(*debugMode)
		}
		if cfg == nil {
			slog.Error("Setup failed")
			os.Exit(1)
		}
		slog.Info("Setup Completed")
		if !*debugMode {
			cfg.SaveConfig(*configFile)
			slog.Info("Config saved")
		} else {
			slog.Info("Config not saved (--no-save-config flag used)")
		}
	}
	slog.Info("Config loaded")

	userConfig, err := config.GetUpdatedUserConfig(cfg)
	if err != nil {
		slog.Error("Failed to get updated user config", "error", err)
	}
	if !proto.Equal(&cfg.UserConfig, userConfig) {
		slog.Info("User config updated")

		proto.Reset(&cfg.UserConfig)
		proto.Merge(&cfg.UserConfig, userConfig)
		if !*debugMode {
			cfg.SaveConfig(*configFile)
			slog.Info("Config saved")
		} else {
			slog.Info("Config not saved (--no-save-config flag used)")
		}
	} else {
		slog.Info("User config not updated")
	}

	// Create a new WebsocketManager
	u, err := url.Parse(cfg.Addr)
	if err != nil {
		slog.Error("Invalid server URL", "error", err)
		os.Exit(1)
	}

	ws := websocket.NewWebsocketManager(u, cfg)
	defer ws.Close()

	rtc := webrtc.NewWebRTCManager(ws)

	Run(ws, cfg, rtc)
	slog.Info("User config update Reinitalize with new config")
	time.Sleep(2 * time.Second)

}

func Run(ws *websocket.WebsocketManager, cfg *config.Config, rtc *webrtc.WebRTCManager) {
	// Initialize the recorder
	recorder := record.NewRecorder(cfg)
	recorder.SetWebsocketManager(ws)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cleanup when function exits
	// Create pipes for WebRTC and recording
	rtcReader, rtcWriter := io.Pipe()

	// // Start the recorder
	recordWriter := io.Discard
	if cfg.UserConfig.RecordingType == msgspb.RecordingType_RECORDING_TYPE_UNSPECIFIED || cfg.UserConfig.RecordingType == msgspb.RecordingType_RECORDING_TYPE_OFF {
		slog.Info("Recording Off")
	} else {
		record, err := recorder.Start(ctx)
		if err != nil {
			slog.Error("Failed to start recorder", "error", err)
			os.Exit(1)
		}
		recordWriter = record
	}

	// Use MultiWriter to send video stream to both WebRTC and the recorder
	multiWriter := io.MultiWriter(rtcWriter, recordWriter)

	// Start camera stream
	stream.Video(ctx, stream.CameraOptions{
		Width:       1920,
		Height:      1080,
		Fps:         30,
		BitRate:     10000000,
		AutoFocus:   true,
		PostProcess: true,

		CameraType: stream.CameraTypeAuto,
	}, multiWriter)

	// Start WebRTC stream
	rtc.StartCamera(rtcReader)

	// Setup clean shutdown
	defer func() {
		if err := recorder.Stop(); err != nil {
			slog.Error("Error stopping recorder", "error", err)
		}
	}()

	// Main message handling loop
	for {
		message, err := ws.ReadMessage()
		if err != nil {
			slog.Error("Error reading message", "error", err)
			// Consider adding a reconnection mechanism here
			time.Sleep(5 * time.Second)
			continue
		}

		switch message.DataType.(type) {
		case *msgspb.Message_Webrtc:
			if err = rtc.HandleMessage(message.GetWebrtc(), message.From); err != nil {
				slog.Error("Error handling WebRTC message", "error", err)
			}
		case *msgspb.Message_HlsRequest:
			if err = recorder.HandleRequest(message.GetHlsRequest()); err != nil {
				slog.Error("Error handling message", "error", err)
			}
		case *msgspb.Message_RecordRequest:
			if err = recorder.HandleRecordRequest(message.GetRecordRequest()); err != nil {
				slog.Error("Error handling message", "error", err)
			}

		case *msgspb.Message_UserConfig:
			// Update user config
			userConfig := message.GetUserConfig()

			if proto.Equal(&cfg.UserConfig, userConfig) {
				continue
			}

			if cfg.UserConfig.RecordingType != userConfig.RecordingType {
				recorder.Stop()

			}

			proto.Merge(&cfg.UserConfig, userConfig)
			if !*debugMode {
				cfg.SaveConfig(*configFile)
				slog.Info("Config saved")
			} else {
				slog.Info("Config not saved (--no-save-config flag used)")
			}

		}

	}
}
