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
    Logger.SetLevel(logrus.InfoLevel)
    Logger.WithField("service", serviceName).Info("Logger initialized")
}
