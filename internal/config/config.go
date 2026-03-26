package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	AccessToken string
}

func Load(customPath string) (*Config, error) {
	// Load .env files in order (later values override earlier)
	homeDir, _ := os.UserHomeDir()
	defaultPath := filepath.Join(homeDir, ".config", "hubspot", ".env")

	paths := []string{defaultPath, ".env"}
	if customPath != "" {
		paths = append(paths, customPath)
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			_ = godotenv.Load(p)
		}
	}

	token := os.Getenv("HUBSPOT_ACCESS_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("HUBSPOT_ACCESS_TOKEN not set.\n\nCreate ~/.config/hubspot/.env with:\n  HUBSPOT_ACCESS_TOKEN=your_token\n\nOr run: hubspot configure")
	}

	return &Config{AccessToken: token}, nil
}
