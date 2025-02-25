package config

import (
	"encoding/json"
	// "errors"
	"os"
)



type Config struct {
	Addr string `json:"addr"`
	CameraUuid string `json:"camera_uuid"`
}


func NewConfig(addr string, camera_uuid string) *Config {
	return &Config{
		Addr: addr,
		CameraUuid: camera_uuid,
	}
}

func LoadConfig(file string) (*Config,error) {
	// Load the configuration from a file
	// return nil,errors.New("testing")

	fileContent, err := os.ReadFile(file)
	if err != nil {
		return nil,err
	}

	var config Config
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		return nil,err
	}

	return &config,nil
}

func (c *Config) SaveConfig(file string) error {
	// Save the configuration to a file
	

	configJson, err := json.Marshal(c)
	if err != nil {
		return err
	}

	err = os.WriteFile(file, configJson, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}