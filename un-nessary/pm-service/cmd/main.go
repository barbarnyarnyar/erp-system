// services/projects/cmd/main.go
package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/sithuhlaing/erp-system/shared/utils"
)

func main() {
    serviceName := getEnv("SERVICE_NAME", "projects")
    port := getEnv("PORT", "8006")
    
    utils.InitLogger(serviceName)
    utils.Logger.Info("Starting Projects Service")

    r := gin.Default()

    // Health check
    r.GET("/health", func(c *gin.Context) {
        utils.SuccessResponse(c, "Projects Service is healthy", gin.H{
            "service": serviceName,
            "version": "1.0.0",
        })
    })

    // Hello World endpoints
    r.GET("/", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Projects Service!", gin.H{
            "service": "projects",
            "domain": "Project Management",
        })
    })

    r.GET("/api/v1/projects/hello", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Projects API!", gin.H{
            "endpoints": []string{
                "GET /api/v1/projects/projects - Project Management",
                "GET /api/v1/projects/tasks - Task Management",
                "GET /api/v1/projects/resources - Resource Management",
                "GET /api/v1/projects/time - Time Tracking",
            },
        })
    })

    r.GET("/api/v1/projects/projects", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Project Management!", gin.H{
            "module": "Project Management",
            "features": []string{"Project Planning", "Portfolio Management", "Deliverables"},
        })
    })

    r.GET("/api/v1/projects/tasks", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Task Management!", gin.H{
            "module": "Task Management",
            "features": []string{"Task Planning", "Dependencies", "Progress Tracking"},
        })
    })

    r.GET("/api/v1/projects/resources", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Resource Management!", gin.H{
            "module": "Resource Management",
            "features": []string{"Resource Allocation", "Capacity Planning", "Utilization"},
        })
    })

    r.GET("/api/v1/projects/time", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Time Tracking!", gin.H{
            "module": "Time Tracking",
            "features": []string{"Time Entry", "Billing", "Reporting"},
        })
    })

    utils.Logger.WithField("port", port).Info("Projects service starting")
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