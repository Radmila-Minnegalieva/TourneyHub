package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	HTTPAddress string
	DatabaseURL string
	CORSOrigins []string
}

func Load() (Config, error) {
	cfg := Config{HTTPAddress: os.Getenv("HTTP_ADDRESS"), DatabaseURL: os.Getenv("DATABASE_URL")}
	if cfg.HTTPAddress == "" {
		cfg.HTTPAddress = ":8080"
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	origins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if origins == "" {
		origins = "http://localhost:5173,http://localhost:8081"
	}
	for _, origin := range strings.Split(origins, ",") {
		if value := strings.TrimSpace(origin); value != "" {
			cfg.CORSOrigins = append(cfg.CORSOrigins, value)
		}
	}
	return cfg, nil
}
