package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTPAddr string

	DatabaseURL string

	JWTIssuer       string
	JWTAccessSecret string
	JWTRefreshSecret string
	AccessTTL       time.Duration
	RefreshTTL      time.Duration
}

func FromEnv() (Config, error) {
	cfg := Config{
		HTTPAddr:         env("HTTP_ADDR", ":8082"),
		DatabaseURL:      env("DATABASE_URL", ""),
		JWTIssuer:        env("JWT_ISSUER", "pipelineapp"),
		JWTAccessSecret:  env("JWT_ACCESS_SECRET", ""),
		JWTRefreshSecret: env("JWT_REFRESH_SECRET", ""),
		AccessTTL:        envDuration("JWT_ACCESS_TTL", 15*time.Minute),
		RefreshTTL:       envDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTAccessSecret == "" {
		return Config{}, fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	if cfg.JWTRefreshSecret == "" {
		return Config{}, fmt.Errorf("JWT_REFRESH_SECRET is required")
	}
	return cfg, nil
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
