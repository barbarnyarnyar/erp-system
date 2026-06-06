package utils

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type LoggerEntry struct {
	serviceName string
	fields      map[string]interface{}
	err         error
}

var Logger *LoggerEntry

func InitLogger(serviceName string) {
	Logger = &LoggerEntry{
		serviceName: serviceName,
		fields:      make(map[string]interface{}),
	}
}

func (l *LoggerEntry) WithField(key string, value interface{}) *LoggerEntry {
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields[key] = value
	return &LoggerEntry{
		serviceName: l.serviceName,
		fields:      newFields,
		err:         l.err,
	}
}

func (l *LoggerEntry) WithError(err error) *LoggerEntry {
	return &LoggerEntry{
		serviceName: l.serviceName,
		fields:      l.fields,
		err:         err,
	}
}

func (l *LoggerEntry) Info(message string, args ...interface{}) {
	l.log("INFO", message, args...)
}

func (l *LoggerEntry) Error(message string, args ...interface{}) {
	l.log("ERROR", message, args...)
}

func (l *LoggerEntry) Debug(message string, args ...interface{}) {
	l.log("DEBUG", message, args...)
}

func (l *LoggerEntry) Warn(message string, args ...interface{}) {
	l.log("WARN", message, args...)
}

func (l *LoggerEntry) log(level, message string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	var formattedMessage string
	if len(args) > 0 {
		formattedMessage = fmt.Sprintf(message, args...)
	} else {
		formattedMessage = message
	}

	fieldsStr := ""
	if len(l.fields) > 0 {
		fieldsStr = fmt.Sprintf(" fields=%v", l.fields)
	}

	errStr := ""
	if l.err != nil {
		errStr = fmt.Sprintf(" error=%v", l.err)
	}

	logMessage := fmt.Sprintf("[%s] [%s] [%s]%s%s %s", timestamp, level, l.serviceName, fieldsStr, errStr, formattedMessage)
	log.Println(logMessage)
}

func GetLogger(c *gin.Context) *LoggerEntry {
	if logger, exists := c.Get("logger"); exists {
		if l, ok := logger.(*LoggerEntry); ok {
			return l
		}
	}
	if Logger != nil {
		return Logger
	}
	return &LoggerEntry{
		serviceName: "unknown",
		fields:      make(map[string]interface{}),
	}
}