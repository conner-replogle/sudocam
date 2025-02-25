package jwtmsg

import (
	"github.com/golang-jwt/jwt/v5"
)

type CameraAdd struct {
	UserID uint `json:"userID"`
	ServerURL string `json:"serverURL"`
	jwt.RegisteredClaims
}

type RegisterCamera struct {
	CameraUUID string `json:"cameraUUID"`
	Token string `json:"token"`
}