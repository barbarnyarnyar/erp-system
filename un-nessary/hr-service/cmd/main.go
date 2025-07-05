// services/hr/go.mod
module github.com/sithuhlaing/erp-system/services/hr

go 1.21

require (
    erp-system/shared v0.0.0
    github.com/gin-gonic/gin v1.9.1
)

replace erp-system/shared => ../../shared

// services/hr/cmd/main.go
package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "erp-system/shared/utils"
)

func main() {
    serviceName := getEnv("SERVICE_NAME", "hr")
    port := getEnv("PORT", "8002")
    
    utils.InitLogger(serviceName)
    utils.Logger.Info("Starting HR Service")

    r := gin.Default()

    // Health check
    r.GET("/health", func(c *gin.Context) {
        utils.SuccessResponse(c, "HR Service is healthy", gin.H{
            "service": serviceName,
            "version": "1.0.0",
        })
    })

    // Hello World endpoints
    r.GET("/", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from HR Service!", gin.H{
            "service": "hr",
            "domain": "Human Resources",
        })
    })

    r.GET("/api/v1/hr/hello", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from HR API!", gin.H{
            "endpoints": []string{
                "GET /api/v1/hr/employees - Employee Management",
                "GET /api/v1/hr/payroll - Payroll Processing",
                "GET /api/v1/hr/time - Time & Attendance",
                "GET /api/v1/hr/training - Training & Skills",
            },
        })
    })

    r.GET("/api/v1/hr/employees", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Employee Management!", gin.H{
            "module": "Employee Management",
            "features": []string{"Employee Records", "Organizational Structure", "Employee Lookup"},
        })
    })

    r.GET("/api/v1/hr/payroll", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Payroll!", gin.H{
            "module": "Payroll Processing",
            "features": []string{"Salary Calculation", "Tax Processing", "Payment Processing"},
        })
    })

    r.GET("/api/v1/hr/time", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Time & Attendance!", gin.H{
            "module": "Time & Attendance",
            "features": []string{"Time Tracking", "Leave Management", "Overtime Calculation"},
        })
    })

    r.GET("/api/v1/hr/training", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Training & Skills!", gin.H{
            "module": "Training & Skills",
            "features": []string{"Training Programs", "Certifications", "Skills Assessment"},
        })
    })

    utils.Logger.WithField("port", port).Info("HR service starting")
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
