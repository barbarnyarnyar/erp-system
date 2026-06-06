package config

import (
	"os"
)

type Config struct {
	Server ServerConfig
	JWT    JWTConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  int // in minutes
	RefreshExpiry int // in hours
}

func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8000"),
			Env:  getEnv("ENV", "development"),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "super-secret-key-123"),
			AccessExpiry:  60, // 1 hour
			RefreshExpiry: 24, // 24 hours
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
