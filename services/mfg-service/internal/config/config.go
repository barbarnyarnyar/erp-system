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
			Port: getEnv("PORT", "8004"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			Username: getEnv("DB_USERNAME", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Database: getEnv("DB_DATABASE", "mfg_service"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(brokers, ","),
			GroupID: getEnv("KAFKA_GROUP_ID", "m-service"),
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

func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

