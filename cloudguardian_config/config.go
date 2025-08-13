package cloudguardian_config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type CloudGuardianConfig struct {
	ApiUrl          string `json:"api_url"`                     // URL of the Cloud Gardian API
	ApiKey          string `json:"api_key"`                     // API key for authentication
	HostSecurityKey string `json:"host_security_key,omitempty"` // Optional host security key
	Debug           bool   `json:"debug"`                       // Debug mode flag
}

// DefaultConfig returns a default configuration for Cloud Gardian.
func DefaultConfig() *CloudGuardianConfig {
	return &CloudGuardianConfig{
		ApiUrl: "https://api.cloud-guardian.net/cloudguardian-api/v1/",
		ApiKey: "",
		Debug:  false,
	}
}

// Validate checks if the configuration is valid.
func (config *CloudGuardianConfig) Validate() error {
	if config.ApiUrl == "" {
		return fmt.Errorf("api_url cannot be empty")
	}
	if !strings.HasSuffix(config.ApiUrl, "/") {
		return fmt.Errorf("api_url must end with a /")
	}
	if !strings.HasPrefix(config.ApiUrl, "http://") && !strings.HasPrefix(config.ApiUrl, "https://") {
		return fmt.Errorf("api_url must start with http:// or https://")
	}
	if config.ApiKey != "" && len(config.ApiKey) != 16 {
		return fmt.Errorf("api_key must be exactly 16 characters long")
	}
	return nil
}

// LoadConfig loads configuration from a JSON file.
// It reads the file, unmarshals the JSON, and validates the configuration.
//
// Parameters:
//   - filename: The path to the configuration file
//
// Returns:
//   - *CloudGuardianConfig: The loaded configuration
//   - error: Any error that occurred during loading or validation
func LoadConfig(filename string) (*CloudGuardianConfig, error) {
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
	// Ensure the API URL ends with a slash
	if !strings.HasSuffix(config.ApiUrl, "/") {
		config.ApiUrl += "/"
	}
	return config, nil
}

// Save saves the configuration to a JSON file.
// It validates the configuration before saving and only includes non-default values.
//
// Parameters:
//   - filename: The path where to save the configuration file
//
// Returns:
//   - error: Any error that occurred during validation or saving
func (config *CloudGuardianConfig) Save(filename string) error {

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	defaultApiUrl := DefaultConfig().ApiUrl

	configFileContent := map[string]any{
		"api_key": config.ApiKey,
	}

	if config.ApiUrl != defaultApiUrl {
		configFileContent["api_url"] = config.ApiUrl
	}

	if config.HostSecurityKey != "" {
		configFileContent["host_security_key"] = config.HostSecurityKey
	}

	if config.Debug {
		configFileContent["debug"] = true
	}

	jsonData, err := json.MarshalIndent(configFileContent, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

var (
	ErrConfigNotFound = errors.New("configuration file not found")
)

type InvalidConfigError struct {
	Msg      string
	Location string
	Err      error
}

// Error returns a formatted error message for the InvalidConfigError.
func (e *InvalidConfigError) Error() string {
	return fmt.Sprintf("Invalid configuration in location %s: %s: %v", e.Location, e.Msg, e.Err)
}

// Unwrap returns the underlying error for error unwrapping.
func (e *InvalidConfigError) Unwrap() error {
	return e.Err
}

// FindAndLoadConfig attempts to find and load a configuration file from multiple locations.
// It searches in the current directory, user config directory, and system-wide config location.
//
// Returns:
//   - *CloudGuardianConfig: The loaded configuration if found
//   - error: ErrConfigNotFound if no config file is found, or other errors during loading
func FindAndLoadConfig() (*CloudGuardianConfig, error) {
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
