package utils

import "log"

// LogPublishErr standardizes the error logging for event publications.
func LogPublishErr(service, topic string, err error) {
	if err != nil {
		log.Printf("[%s] ERROR: failed to publish event %s: %v", service, topic, err)
	}
}
