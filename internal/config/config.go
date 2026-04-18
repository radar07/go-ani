package config

import (
	"os"
	"path/filepath"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	// Player settings
	Player  string
	Quality string
	Mode    string

	// Paths
	DownloadDir string
	HistoryPath string

	// Behavior
	NoDetach bool
}

// LoadConfig creates Config with defaults from environment variables
func LoadConfig() *Config {
	cfg := &Config{
		Player:      getEnv("ANI_PLAYER", "mpv"),
		Quality:     getEnv("ANI_QUALITY", "best"),
		Mode:        getEnv("ANI_MODE", "sub"),
		DownloadDir: getEnv("ANI_DOWNLOAD_DIR", "."),
		NoDetach:    getEnvBool("ANI_NO_DETACH", false),
	}

	// Set history path
	configDir := getConfigDir()
	cfg.HistoryPath = filepath.Join(configDir, "history.json")

	return cfg
}

// EnsureHistoryDir creates config directory if it doesn't exist
func (c *Config) EnsureHistoryDir() error {
	dir := filepath.Dir(c.HistoryPath)
	return os.MkdirAll(dir, 0755)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getConfigDir() string {
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return filepath.Join(configHome, "go-ani")
	}

	// Fallback to ~/.config/go-ani
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", "go-ani")
	}

	// Last resort: current directory
	return ".go-ani"
}
