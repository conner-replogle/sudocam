package models

type CameraStatusUpdate struct {
	CameraUUID   string `json:"cameraUUID"`
	OnlineStatus bool   `json:"onlineStatus"`
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
