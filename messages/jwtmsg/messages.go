package jwtmsg

import "github.com/golang-jwt/jwt/v5"

// CameraAdd contains the information needed to add a camera
type CameraAdd struct {
	jwt.RegisteredClaims
	UserID       uint `json:"userId"`
	FriendlyName string `json:"friendlyName"`
	ServerURL    string `json:"serverUrl"`
	WifiNetwork  string `json:"wifiNetwork,omitempty"`
	WifiPassword string `json:"wifiPassword,omitempty"`
}

// RegisterCamera is the request sent to register a camera
type RegisterCamera struct {
	CameraUUID   string `json:"cameraUUID"`
	Token        string `json:"token"`
	FriendlyName string `json:"friendlyName"`
}
