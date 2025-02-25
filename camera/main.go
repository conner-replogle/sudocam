// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

//go:build !js
// +build !js

// play-from-disk demonstrates how to send video and/or audio to your browser from files saved to disk.
package main

import (
	"camera/config"
	"camera/setup"
	"camera/webrtc"
	"camera/websocket"
	"flag"
	"log/slog"
	"net/url"
	"os"
)



func main() { //nolint

	// Assert that we have an audio or video file
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	flag.Parse()
	config,err := config.LoadConfig("config.json")
	if err != nil {
		slog.Info("No config file found")

		config = setup.RunSetup()
		slog.Info("Setup Completed")
		config.SaveConfig("config.json")
		slog.Info("Config saved")
	}
	slog.Info("Config loaded")


	// Create a new WebsocketManager
	u,err := url.Parse(config.Addr)
	if err != nil {
		panic(err)
	}
	ws := websocket.NewWebsocketManager(u, config.CameraUuid)
	defer ws.Close()
	rtc := webrtc.NewWebRTCManager(ws)
	rtc.StartCamera()

	// Create a video track
	
	for {
		message, err := ws.ReadMessage()
		if err != nil {
			slog.Error("Error reading message", err)
		}

		webRtc := message.GetWebrtc()

		if webRtc != nil {
			err = rtc.HandleMessage(webRtc,message.From)
			if err != nil {
				slog.Error("Error handling message", err)
			}
		}


		
	}
}


