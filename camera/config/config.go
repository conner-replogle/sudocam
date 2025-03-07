package config

import (
	"encoding/json"
	"os"
)

// Config holds the camera configuration
type Config struct {
	CameraUuid string `json:"cameraUUID"`
	CameraName string `json:"cameraName"`
	Addr       string `json:"addr"`
	RecordDir  string `json:"record_dir"`
	// Add any other configuration fields here
}

// LoadConfig loads the configuration from a JSON file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set default record directory if not specified
	if config.RecordDir == "" {
		config.RecordDir = "recordings"
	}

	return &config, nil
}

func DeleteConfig(filename string) error {
	return os.Remove(filename)
}

// SaveConfig saves the configuration to a JSON file
func (c *Config) SaveConfig(filename string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
