// shared/utils/response.go
package utils

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
)

type Response struct {
    Status    string      `json:"status"`
    Message   string      `json:"message"`
    Data      interface{} `json:"data,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
}

func SuccessResponse(c *gin.Context, message string, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Status:    "success",
        Message:   message,
        Data:      data,
        Timestamp: time.Now(),
    })
}

func ErrorResponse(c *gin.Context, statusCode int, message string) {
    c.JSON(statusCode, Response{
        Status:    "error",
        Message:   message,
        Timestamp: time.Now(),
    })
}