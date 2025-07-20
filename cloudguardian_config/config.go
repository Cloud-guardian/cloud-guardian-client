package cloudguardian_config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type CloudGardianConfig struct {
	ApiUrl string `json:"api_url"` // URL of the Cloud Gardian API
	ApiKey string `json:"api_key"` // API key for authentication
	Debug  bool   `json:"debug"`   // Debug mode flag
}

// DefaultConfig returns a default configuration for Cloud Gardian.
func DefaultConfig() *CloudGardianConfig {
	return &CloudGardianConfig{
		ApiUrl: "https://api.cloud-guardian.net/cloudguardian-api/v1/",
		ApiKey: "",
		Debug:  false,
	}
}

// Validate checks if the configuration is valid.
func (c *CloudGardianConfig) Validate() error {
	if c.ApiUrl == "" {
		return fmt.Errorf("api_url cannot be empty")
	}
	if c.ApiKey != "" && len(c.ApiKey) != 32 {
		return fmt.Errorf("api_key must be exactly 32 characters long")
	}
	return nil
}

// Load config from a file.
func LoadConfig(filename string) (*CloudGardianConfig, error) {
	config := DefaultConfig()
	// Load config from json file:
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	err = json.Unmarshal(jsonData, config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return config, nil
}

var (
	ErrConfigNotFound = errors.New("configuration file not found")
)

type InvalidConfigError struct {
	Msg      string
	Location string
	Err      error
}

func (e *InvalidConfigError) Error() string {
	return fmt.Sprintf("Invalid configuration in location %s: %s: %v", e.Location, e.Msg, e.Err)
}

func (e *InvalidConfigError) Unwrap() error {
	return e.Err
}

// Try to find the config file:
func FindAndLoadConfig() (*CloudGardianConfig, error) {
	// check the following locations:
	// 1. Current directory
	// 2. ~/.config/cloud-guardian.json
	// 3. /etc/cloud-guardian.json
	locations := []string{
		"cloud-guardian.json",                              // Current directory
		os.Getenv("HOME") + "/.config/cloud-guardian.json", // User config
		"/etc/cloud-guardian.json",                         // System-wide config
	}
	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			config, err := LoadConfig(loc)
			if err != nil {
				return nil, &InvalidConfigError{
					Msg:      "Failed to load config",
					Location: loc,
					Err:      err,
				}
			}
			return config, nil
		}
	}
	return nil, ErrConfigNotFound
}
