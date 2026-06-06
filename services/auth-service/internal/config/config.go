package config

import (
	"os"
	"strings"
)

type Config struct {
	Server ServerConfig
	JWT    JWTConfig
	Kafka  KafkaConfig
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

type KafkaConfig struct {
	Brokers []string
}

func Load() (*Config, error) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

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
		Kafka: KafkaConfig{
			Brokers: strings.Split(brokers, ","),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
