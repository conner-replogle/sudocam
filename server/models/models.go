package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"password"`
}

type Camera struct {
	gorm.Model
	CameraUUID   string     `json:"cameraUUID" gorm:"unique"`
	UserID       uint       `json:"userID"`
	FriendlyName string     `json:"friendlyName"`
	LastOnline   *time.Time `json:"lastOnline"`
	OnlineStatus bool       `json:"onlineStatus" gorm:"default:false"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Valid bool   `json:"valid"`
	Email string `json:"email,omitempty"`
}

type TURNCredentials struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	TTL      int      `json:"ttl"`
	URIs     []string `json:"uris"`
}
