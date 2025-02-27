package models

type CameraStatusUpdate struct {
	CameraUUID   string `json:"cameraUUID"`
	OnlineStatus bool   `json:"onlineStatus"`
}
