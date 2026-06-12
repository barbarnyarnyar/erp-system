package config

import (
	"os"
	"strings"
)

type Config struct {
	Server ServerConfig
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

type KafkaConfig struct {
	Brokers []string
	GroupID string
}

func Load() (*Config, error) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8008"),
			Env:  getEnv("ENV", "development"),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(brokers, ","),
			GroupID: getEnv("KAFKA_GROUP_ID", "plm-service"),
		},
		TLS: TLSConfig{
			Enabled:  getEnv("TLS_ENABLED", "false") == "true",
			CertFile: getEnv("TLS_CERT_FILE", ""),
			KeyFile:  getEnv("TLS_KEY_FILE", ""),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
