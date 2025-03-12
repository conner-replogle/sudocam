package jwtmsg

import (
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// CameraAdd contains the information needed to add a camera
type CameraAdd struct {
	jwt.RegisteredClaims
	UserID       string `json:"userId"`
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


type AuthClaims struct {
	Email      string     `json:"email"`
	EntityID   string     `json:"entityID"`
	EntityType EntityType `json:"entityType"`
	jwt.RegisteredClaims
}

type EntityType string

const (
	EntityTypeUser   EntityType = "user"
	EntityTypeCamera EntityType = "camera"
)

// MarshalJSON implements json.Marshaler interface
func (e EntityType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(e))
}

// UnmarshalJSON implements json.Unmarshaler interface
func (e *EntityType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch EntityType(s) {
	case EntityTypeUser, EntityTypeCamera:
		*e = EntityType(s)
		return nil
	default:
		return fmt.Errorf("invalid entity type: %s", s)
	}
}

// String returns the string representation of the entity type
func (e EntityType) String() string {
	return string(e)
}