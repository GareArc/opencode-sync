package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/GareArc/opencode-sync/internal/paths"
)

// Config represents the opencode-sync configuration
type Config struct {
	Repo       RepoConfig       `json:"repo"`
	Encryption EncryptionConfig `json:"encryption"`
	Sync       SyncConfig       `json:"sync"`
}

// RepoConfig holds Git repository configuration
type RepoConfig struct {
	URL    string `json:"url"`
	Branch string `json:"branch"`
}

// EncryptionConfig holds encryption settings
type EncryptionConfig struct {
	Enabled bool   `json:"enabled"`
	KeyFile string `json:"keyFile,omitempty"`
}

// SyncConfig holds sync behavior settings
type SyncConfig struct {
	IncludeAuth    bool     `json:"includeAuth"`
	IncludeMcpAuth bool     `json:"includeMcpAuth"`
	Exclude        []string `json:"exclude,omitempty"`
}

// Default returns a default configuration
func Default() *Config {
	p, _ := paths.Get()
	keyFile := ""
	if p != nil {
		keyFile = p.KeyFile()
	}

	return &Config{
		Repo: RepoConfig{
			Branch: "main",
		},
		Encryption: EncryptionConfig{
			Enabled: false,
			KeyFile: keyFile,
		},
		Sync: SyncConfig{
			IncludeAuth:    false,
			IncludeMcpAuth: false,
			Exclude:        []string{"node_modules", "*.log", "bun.lock"},
		},
	}
}

// Load loads the configuration from the default location
func Load() (*Config, error) {
	p, err := paths.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get paths: %w", err)
	}

	configFile := p.ConfigFile()

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, nil // No config yet
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// Save saves the configuration to the default location
func Save(cfg *Config) error {
	p, err := paths.Get()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	// Ensure config directory exists
	if err := os.MkdirAll(p.ConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configFile := p.ConfigFile()
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Repo.URL == "" {
		return fmt.Errorf("repo.url is required")
	}

	if c.Sync.IncludeAuth && !c.Encryption.Enabled {
		return fmt.Errorf("sync.includeAuth requires encryption.enabled to be true")
	}

	if c.Sync.IncludeMcpAuth && !c.Encryption.Enabled {
		return fmt.Errorf("sync.includeMcpAuth requires encryption.enabled to be true")
	}

	return nil
}

// KeyFileExists checks if the encryption key file exists
func (c *Config) KeyFileExists() bool {
	if c.Encryption.KeyFile == "" {
		return false
	}

	// Expand ~ to home directory
	keyFile := c.Encryption.KeyFile
	if len(keyFile) > 0 && keyFile[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		keyFile = filepath.Join(home, keyFile[1:])
	}

	_, err := os.Stat(keyFile)
	return err == nil
}
