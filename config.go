package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	HRMURL            string `json:"hrm_url"`
	EmployeeCode      string `json:"employee_code"`
	RegistrationKey   string `json:"registration_key"`
	Token             string `json:"token"`
	DeviceUUID        string `json:"device_uuid"`
	HeartbeatInterval int    `json:"heartbeat_interval"`
	ActivityInterval  int    `json:"activity_interval"`
	Registered        bool   `json:"registered"`
}

// getConfigPath resolves the local path to store the configuration.
func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".employee-agent")
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// LoadConfig loads config settings from config.json.
func LoadConfig() (Config, error) {
	var config Config
	path, err := getConfigPath()
	if err != nil {
		return config, err
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config intervals if file doesn't exist
			config.HeartbeatInterval = 30
			config.ActivityInterval = 60
			config.Registered = false
			return config, nil
		}
		return config, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&config)
	return config, err
}

// SaveConfig writes config settings to config.json.
func SaveConfig(config Config) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}
