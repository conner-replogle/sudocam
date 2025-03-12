package config

import (
	"encoding/json"
	"fmt"
	pb "messages/msgspb"
	"net/http"
)

func GetUpdatedUserConfig(config *Config) (*pb.UserConfig, error) {
	// Update the user configuration
	// Return the updated user configuration

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/cameras/%s/config", config.Addr, config.CameraUuid), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", config.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config : %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	var userConfig pb.UserConfig
	if err := json.NewDecoder(resp.Body).Decode(&userConfig); err != nil {
		return nil, fmt.Errorf("failed to decode User Config: %w", err)
	}

	return &userConfig, nil

}
