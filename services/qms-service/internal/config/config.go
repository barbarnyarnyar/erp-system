package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
	TLS      TLSConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	SSLMode  string
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
			Port: getEnv("PORT", "8009"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			Username: getEnv("DB_USERNAME", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Database: getEnv("DB_DATABASE", "erp_quality"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(brokers, ","),
			GroupID: getEnv("KAFKA_GROUP_ID", "qms-service"),
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

func getEnvAsInt(key string, defaultValue int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}
