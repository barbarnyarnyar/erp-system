// shared/utils/logger.go
package utils

import (
	"os"
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func InitLogger(serviceName string) {
	Logger = logrus.New()
	Logger.SetOutput(os.Stdout)
	Logger.SetFormatter(&logrus.JSONFormatter{})
	
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)
	
	Logger.WithField("service", serviceName).Info("Logger initialized")
}