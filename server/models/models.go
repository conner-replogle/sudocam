package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `json:"email" gorm:"unique"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Camera struct {
	gorm.Model
	CameraUUID string     `json:"cameraUUID" gorm:"unique"`
	UserID     uint       `json:"userID"`
	Name       string     `json:"name"`
	LastSeen   *time.Time `json:"lastSeen"`
	IsOnline   bool       `json:"isOnline" gorm:"default:false"`
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
