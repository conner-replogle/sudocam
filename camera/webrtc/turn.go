package webrtc

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pion/webrtc/v4"
)
type Turn struct {
	IceServers webrtc.ICEServer `json:"iceServers"`
}

func fetchTURNCredentials(addr string) (*webrtc.ICEServer, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/turn", addr))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch TURN credentials: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var creds Turn
	if err := json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		return nil, fmt.Errorf("failed to decode TURN credentials: %w", err)
	}

	return &creds.IceServers, nil
}