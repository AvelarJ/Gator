package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	Database_url string `json:"db_url"`
	Current_user string `json:"current_user_name"`
}

func Read() (Config, error) {
	config_location, err := getConfigLocation()
	if err != nil {
		return Config{}, err
	}

	file, err := os.Open(config_location)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func (c *Config) SetUser(user string) {
	c.Current_user = user

	write(*c)
}

// Helper to write the config to disk
func write(config Config) error {
	config_location, err := getConfigLocation()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(config_location, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return err
	}

	return nil
}

// Helper to get the config location in the home directory
func getConfigLocation() (string, error) {
	home_dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home_dir, configFileName), nil
}
