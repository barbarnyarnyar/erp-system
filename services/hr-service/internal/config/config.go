package config

import (
	"os"
	"strings"
)

type Config struct {
	Server ServerConfig
	Kafka  KafkaConfig
}

type ServerConfig struct {
	Port string
	Env  string
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
			Port: getEnv("PORT", "8003"),
			Env:  getEnv("ENV", "development"),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(brokers, ","),
			GroupID: getEnv("KAFKA_GROUP_ID", "hr-service"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
