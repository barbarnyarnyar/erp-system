// shared/config/config.go
package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName string
	Port        string
	Database    DatabaseConfig
	Redis       RedisConfig
	RabbitMQ    RabbitMQConfig
	JWT         JWTConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type RabbitMQConfig struct {
	URL      string
	Exchange string
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

func Load() *Config {
	return &Config{
		ServiceName: getEnv("SERVICE_NAME", "erp-service"),
		Port:        getEnv("PORT", "8080"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "erp_user"),
			Password: getEnv("DB_PASSWORD", "erp_password"),
			DBName:   getEnv("DB_NAME", "erp_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		RabbitMQ: RabbitMQConfig{
			URL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			Exchange: getEnv("RABBITMQ_EXCHANGE", "erp.events"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key"),
			Expiration: time.Hour * 24,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}