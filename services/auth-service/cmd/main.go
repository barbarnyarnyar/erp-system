// services/auth-service/cmd/main.go
package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/sithuhlaing/erp-system/shared/utils"
)

func main() {
    serviceName := getEnv("SERVICE_NAME", "auth")
    port := getEnv("PORT", "8000")
    
    utils.InitLogger(serviceName)
    utils.Logger.Info("Starting Auth Service")

    r := gin.Default()

    // Health check
    r.GET("/health", func(c *gin.Context) {
        utils.SuccessResponse(c, "Auth Service is healthy", gin.H{
            "service": serviceName,
            "version": "1.0.0",
        })
    })

    // Hello World endpoints
    r.GET("/", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Auth Service!", gin.H{
            "service": "auth",
            "domain": "Authentication & Authorization",
        })
    })

    r.GET("/api/v1/auth/hello", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Auth API!", gin.H{
            "endpoints": []string{
                "POST /api/v1/auth/login - User Login",
                "POST /api/v1/auth/register - User Registration",
                "POST /api/v1/auth/refresh - Refresh Token",
                "POST /api/v1/auth/logout - User Logout",
            },
        })
    })

    r.GET("/api/v1/auth/login", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Login!", gin.H{
            "module": "User Authentication",
            "features": []string{"JWT Tokens", "Session Management", "Password Validation"},
        })
    })

    r.GET("/api/v1/auth/users", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from User Management!", gin.H{
            "module": "User Management",
            "features": []string{"User Profiles", "Role Management", "Permissions"},
        })
    })

    utils.Logger.WithField("port", port).Info("Auth service starting")
    if err := r.Run(":" + port); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
