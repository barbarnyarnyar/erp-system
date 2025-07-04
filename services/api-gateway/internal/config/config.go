// File: api-gateway/internal/config/config.go
package config

import (
	"os"
)

type Config struct {
	Port      string
	JWTSecret string
	Services  ServiceConfig
}

type ServiceConfig struct {
	AuthService string
	FMService   string
	HRService   string
	SCMService  string
	MService    string
	CRMService  string
	PMService   string
}

func Load() (*Config, error) {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", "your-super-secret-key"),
		Services: ServiceConfig{
			AuthService: getEnv("AUTH_SERVICE_URL", "http://auth-service:8090"),
			FMService:   getEnv("FM_SERVICE_URL", "http://fm-service:8081"),
			HRService:   getEnv("HR_SERVICE_URL", "http://hr-service:8082"),
			SCMService:  getEnv("SCM_SERVICE_URL", "http://scm-service:8083"),
			MService:    getEnv("M_SERVICE_URL", "http://m-service:8084"),
			CRMService:  getEnv("CRM_SERVICE_URL", "http://crm-service:8085"),
			PMService:   getEnv("PM_SERVICE_URL", "http://pm-service:8086"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
