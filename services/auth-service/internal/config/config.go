package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strings"
)

type Config struct {
	Server ServerConfig
	JWT    JWTConfig
	Kafka  KafkaConfig
	TLS    TLSConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type TLSConfig struct {
	Enabled  bool
	CertFile string
	KeyFile  string
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
			Secret:        getJWTSecret(),
			AccessExpiry:  60, // 1 hour
			RefreshExpiry: 24, // 24 hours
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(brokers, ","),
		},
		TLS: TLSConfig{
			Enabled:  getEnv("TLS_ENABLED", "false") == "true",
			CertFile: getEnv("TLS_CERT_FILE", ""),
			KeyFile:  getEnv("TLS_KEY_FILE", ""),
		},
	}, nil
}

func getJWTSecret() string {
	if val := os.Getenv("JWT_SECRET"); val != "" {
		return val
	}
	// Generate a random per-run secret for development
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		log.Fatalf("Failed to generate random JWT secret: %v", err)
	}
	secret := hex.EncodeToString(buf)
	log.Printf("WARNING: JWT_SECRET not set. Generated random per-run secret. Tokens will be invalid after restart.")
	return secret
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
