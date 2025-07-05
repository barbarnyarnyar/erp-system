// shared/utils/response.go
package utils

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type APIResponse struct {
	Status    string      `json:"status"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	RequestID string      `json:"request_id"`
	Timestamp time.Time   `json:"timestamp"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Status:    "success",
		Data:      data,
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

func ErrorResponse(c *gin.Context, statusCode int, code, message, details string) {
	c.JSON(statusCode, APIResponse{
		Status: "error",
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

func getRequestID(c *gin.Context) string {
	if requestID := c.GetString("request_id"); requestID != "" {
		return requestID
	}
	return uuid.New().String()
}