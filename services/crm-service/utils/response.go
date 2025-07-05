package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// StandardResponse represents a standard API response
type StandardResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Service   string      `json:"service"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Service   string    `json:"service"`
	Status    string    `json:"status"`
	Port      string    `json:"port"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
}

// ResponseHelper provides helper methods for API responses
type ResponseHelper struct {
	serviceName string
}

// NewResponseHelper creates a new response helper
func NewResponseHelper(serviceName string) *ResponseHelper {
	return &ResponseHelper{
		serviceName: serviceName,
	}
}

// Success sends a successful response
func (r *ResponseHelper) Success(c *gin.Context, message string, data interface{}) {
	logger := GetLogger(c)
	requestID := r.getRequestID(c)
	
	response := StandardResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Service:   r.serviceName,
		RequestID: requestID,
		Timestamp: time.Now(),
	}
	
	logger.Info("Success response: %s", message)
	c.JSON(http.StatusOK, response)
}

// Error sends an error response
func (r *ResponseHelper) Error(c *gin.Context, statusCode int, message string, err error) {
	logger := GetLogger(c)
	requestID := r.getRequestID(c)
	
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
		logger.Error("Error response: %s - %s", message, errorMessage)
	} else {
		logger.Error("Error response: %s", message)
	}
	
	response := StandardResponse{
		Success:   false,
		Message:   message,
		Error:     errorMessage,
		Service:   r.serviceName,
		RequestID: requestID,
		Timestamp: time.Now(),
	}
	
	c.JSON(statusCode, response)
}

// BadRequest sends a 400 Bad Request response
func (r *ResponseHelper) BadRequest(c *gin.Context, message string) {
	r.Error(c, http.StatusBadRequest, message, nil)
}

// Unauthorized sends a 401 Unauthorized response
func (r *ResponseHelper) Unauthorized(c *gin.Context, message string) {
	r.Error(c, http.StatusUnauthorized, message, nil)
}

// NotFound sends a 404 Not Found response
func (r *ResponseHelper) NotFound(c *gin.Context, message string) {
	r.Error(c, http.StatusNotFound, message, nil)
}

// InternalServerError sends a 500 Internal Server Error response
func (r *ResponseHelper) InternalServerError(c *gin.Context, message string, err error) {
	r.Error(c, http.StatusInternalServerError, message, err)
}

// Health sends a health check response
func (r *ResponseHelper) Health(c *gin.Context, port string) {
	requestID := r.getRequestID(c)
	
	response := HealthResponse{
		Service:   r.serviceName,
		Status:    "healthy",
		Port:      port,
		RequestID: requestID,
		Timestamp: time.Now(),
	}
	
	c.JSON(http.StatusOK, response)
}

// getRequestID extracts request ID from context
func (r *ResponseHelper) getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// ValidateJSON validates JSON binding and sends error response if validation fails
func (r *ResponseHelper) ValidateJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		r.BadRequest(c, "Invalid JSON payload: "+err.Error())
		return false
	}
	return true
}