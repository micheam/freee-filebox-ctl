// Package config provides functionality to load application configuration
// from a TOML file following the XDG Base Directory specification.
package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const (
	appName        = "ffbox"
	configFileName = "config.toml"
)

//go:embed config.example.toml
var defaultConfig []byte

// Config represents the application configuration
type Config struct {
	OAuth2 struct {
		// TokenFile is the path to the file where OAuth2 tokens are stored
		TokenFile string `toml:"token_file"`
		// LocalAddr is the local address for OAuth2 callback server
		LocalAddr string `toml:"local_addr"`
	} `toml:"oauth2"`

	Freee struct {
		// CompanyID is the default freee company ID to use for operations
		CompanyID int64 `toml:"company_id"`
	} `toml:"freee"`
}

func (c *Config) Marshal() ([]byte, error) {
	return toml.Marshal(c)
}

// Default returns a Config with default values
func Default() *Config {
	cfg := &Config{}
	if err := toml.Unmarshal(defaultConfig, cfg); err != nil {
		panic(fmt.Sprintf("unmarshal default config: %v", err)) // should not happen
	}
	return cfg
}

func InitConfigFile() error {
	configPath := ConfigPath()
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	defaultConfig := Default()
	data, err := toml.Marshal(defaultConfig)
	if err != nil {
		panic(fmt.Sprintf("marshal default config: %v", err)) // should not happen
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// Load loads configuration from file following XDG Base Directory specification
// It searches for config.toml in the following order:
// 1. $XDG_CONFIG_HOME/ffbox/config.toml
// 2. $HOME/.config/ffbox/config.toml
//
// If no config file is found, it returns the default configuration.
func Load() (*Config, error) {
	configPath, err := findConfigFile()
	if err != nil {
		return nil, fmt.Errorf("no config file found: %w", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := Default()
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return cfg, nil
}

// findConfigFile searches for config.toml following XDG Base Directory specification
func findConfigFile() (string, error) {
	configDirs := getConfigDirs()

	for _, dir := range configDirs {
		configPath := filepath.Join(dir, appName, configFileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	return "", fmt.Errorf("config file not found")
}

// getConfigDirs returns config directories in order of preference
// following XDG Base Directory specification
func getConfigDirs() []string {
	var dirs []string

	// 1. $XDG_CONFIG_HOME
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		dirs = append(dirs, xdgConfigHome)
	}

	// 2. $HOME/.config
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".config"))
	}

	return dirs
}

// ConfigPath returns the expected config file path
// This can be used to inform users where to place the config file
func ConfigPath() string {
	dirs := getConfigDirs()
	if len(dirs) > 0 {
		return filepath.Join(dirs[0], appName, configFileName)
	}
	return configFileName
}
