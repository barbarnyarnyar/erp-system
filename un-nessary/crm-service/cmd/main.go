// services/crm/cmd/main.go
package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "erp-system/shared/utils"
)

func main() {
    serviceName := getEnv("SERVICE_NAME", "crm")
    port := getEnv("PORT", "8005")
    
    utils.InitLogger(serviceName)
    utils.Logger.Info("Starting CRM Service")

    r := gin.Default()

    // Health check
    r.GET("/health", func(c *gin.Context) {
        utils.SuccessResponse(c, "CRM Service is healthy", gin.H{
            "service": serviceName,
            "version": "1.0.0",
        })
    })

    // Hello World endpoints
    r.GET("/", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from CRM Service!", gin.H{
            "service": "crm",
            "domain": "Customer Relationship Management",
        })
    })

    r.GET("/api/v1/crm/hello", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from CRM API!", gin.H{
            "endpoints": []string{
                "GET /api/v1/crm/customers - Customer Management",
                "GET /api/v1/crm/leads - Lead Management",
                "GET /api/v1/crm/opportunities - Sales Opportunities",
                "GET /api/v1/crm/orders - Sales Orders",
            },
        })
    })

    r.GET("/api/v1/crm/customers", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Customer Management!", gin.H{
            "module": "Customer Management",
            "features": []string{"Customer Records", "Contact Management", "Account History"},
        })
    })

    r.GET("/api/v1/crm/leads", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Lead Management!", gin.H{
            "module": "Lead Management",
            "features": []string{"Lead Capture", "Qualification", "Conversion"},
        })
    })

    r.GET("/api/v1/crm/opportunities", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Sales Opportunities!", gin.H{
            "module": "Sales Opportunities",
            "features": []string{"Pipeline Management", "Forecasting", "Quotes"},
        })
    })

    r.GET("/api/v1/crm/orders", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Sales Orders!", gin.H{
            "module": "Sales Orders",
            "features": []string{"Order Processing", "Fulfillment", "Customer Service"},
        })
    })

    utils.Logger.WithField("port", port).Info("CRM service starting")
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