// services/manufacturing/cmd/main.go
package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "erp-system/shared/utils"
)

func main() {
    serviceName := getEnv("SERVICE_NAME", "manufacturing")
    port := getEnv("PORT", "8004")
    
    utils.InitLogger(serviceName)
    utils.Logger.Info("Starting Manufacturing Service")

    r := gin.Default()

    // Health check
    r.GET("/health", func(c *gin.Context) {
        utils.SuccessResponse(c, "Manufacturing Service is healthy", gin.H{
            "service": serviceName,
            "version": "1.0.0",
        })
    })

    // Hello World endpoints
    r.GET("/", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Manufacturing Service!", gin.H{
            "service": "manufacturing",
            "domain": "Manufacturing Management",
        })
    })

    r.GET("/api/v1/manufacturing/hello", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Manufacturing API!", gin.H{
            "endpoints": []string{
                "GET /api/v1/manufacturing/boms - Bill of Materials",
                "GET /api/v1/manufacturing/production - Production Orders",
                "GET /api/v1/manufacturing/quality - Quality Control",
                "GET /api/v1/manufacturing/workcenters - Work Centers",
            },
        })
    })

    r.GET("/api/v1/manufacturing/boms", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Bill of Materials!", gin.H{
            "module": "Bill of Materials",
            "features": []string{"BOM Structure", "Component Lists", "Routing"},
        })
    })

    r.GET("/api/v1/manufacturing/production", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Production Orders!", gin.H{
            "module": "Production Orders",
            "features": []string{"Production Planning", "Shop Floor Control", "MRP"},
        })
    })

    r.GET("/api/v1/manufacturing/quality", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Quality Control!", gin.H{
            "module": "Quality Control",
            "features": []string{"Quality Plans", "Inspections", "Test Results"},
        })
    })

    r.GET("/api/v1/manufacturing/workcenters", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Work Centers!", gin.H{
            "module": "Work Centers",
            "features": []string{"Capacity Planning", "Scheduling", "Utilization"},
        })
    })

    utils.Logger.WithField("port", port).Info("Manufacturing service starting")
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