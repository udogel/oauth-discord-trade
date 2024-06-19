package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Port         string `json:"localhost_port"`
	TimeoutTime  int    `json:"timeout_time"`
}

func main() {
	config, err := readConfig("config.json")
	if err != nil {
		log.Fatal("Error loading configuration file: ", err)
	}

	if err := validateConfig(config); err != nil {
		log.Fatal("Invalid configuration: ", err)
	}

	timeoutTime := config.TimeoutTime
	if timeoutTime == 0 {
		timeoutTime = 60
	}

	userAuth, err := AuthenticateUser(config)
	if err != nil {
		log.Fatal("Authentication failed: ", err)
	}
	log.Print("Successfully received server response. Token type: ", userAuth.Type)
}

func readConfig(path string) (Config, error) {
	var config Config
	configFile, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	if err = json.Unmarshal(configFile, &config); err != nil {
		return config, fmt.Errorf("failed to unmarshall JSON: %w", err)
	}

	return config, nil
}

func validateConfig(config Config) error {
	if config.Port == "" || config.ClientID == "" || config.ClientSecret == "" || config.TimeoutTime == 0 {
		return fmt.Errorf("all config fields must be set")
	}
	return nil
}
