package config

import (
	"os"
)

type Config struct {
	Port           string
	CRMServiceURL  string
	SCMServiceURL  string
	FMServiceURL   string
}

func Load() *Config {
	port := getEnv("PORT", "8085")
	crmURL := getEnv("CRM_SERVICE_URL", "http://crm-service:8002")
	scmURL := getEnv("SCM_SERVICE_URL", "http://scm-service:8006")
	fmURL := getEnv("FM_SERVICE_URL", "http://fm-service:8001")

	return &Config{
		Port:          port,
		CRMServiceURL: crmURL,
		SCMServiceURL: scmURL,
		FMServiceURL:  fmURL,
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
