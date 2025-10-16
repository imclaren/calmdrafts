package config

import (
	"encoding/json"
	"os"
	"time"
)

// Config holds the application configuration
type Config struct {
	CheckInterval    time.Duration `json:"check_interval"`    // How often to check drafts (e.g., "1h", "30m")
	CleanupAge       time.Duration `json:"cleanup_age"`       // Age threshold for deleting empty drafts (default: 7 days)
	CredentialsPath  string        `json:"credentials_path"`  // Path to Google OAuth credentials JSON
	TokenPath        string        `json:"token_path"`        // Path to store OAuth token
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		CheckInterval:   1 * time.Hour,
		CleanupAge:      7 * 24 * time.Hour, // 7 days
		CredentialsPath: "credentials.json",
		TokenPath:       "token.json",
	}
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		// Return default config if file doesn't exist
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig saves configuration to a JSON file
func SaveConfig(path string, config *Config) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}
