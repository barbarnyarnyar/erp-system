// services/scm/cmd/main.go
package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/sithuhlaing/erp-system/shared/utils"
)

func main() {
    serviceName := getEnv("SERVICE_NAME", "scm")
    port := getEnv("PORT", "8003")
    
    utils.InitLogger(serviceName)
    utils.Logger.Info("Starting SCM Service")

    r := gin.Default()

    // Health check
    r.GET("/health", func(c *gin.Context) {
        utils.SuccessResponse(c, "SCM Service is healthy", gin.H{
            "service": serviceName,
            "version": "1.0.0",
        })
    })

    // Hello World endpoints
    r.GET("/", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from SCM Service!", gin.H{
            "service": "scm",
            "domain": "Supply Chain Management",
        })
    })

    r.GET("/api/v1/scm/hello", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from SCM API!", gin.H{
            "endpoints": []string{
                "GET /api/v1/scm/products - Product Management",
                "GET /api/v1/scm/vendors - Vendor Management",
                "GET /api/v1/scm/inventory - Inventory Management",
                "GET /api/v1/scm/procurement - Procurement",
            },
        })
    })

    r.GET("/api/v1/scm/products", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Product Management!", gin.H{
            "module": "Product Management",
            "features": []string{"Product Master Data", "Categories", "Pricing"},
        })
    })

    r.GET("/api/v1/scm/vendors", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Vendor Management!", gin.H{
            "module": "Vendor Management",
            "features": []string{"Vendor Master Data", "Contracts", "Performance Tracking"},
        })
    })

    r.GET("/api/v1/scm/inventory", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Inventory Management!", gin.H{
            "module": "Inventory Management",
            "features": []string{"Stock Levels", "Warehouse Management", "Movements"},
        })
    })

    r.GET("/api/v1/scm/procurement", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Procurement!", gin.H{
            "module": "Procurement",
            "features": []string{"Purchase Orders", "Requisitions", "Vendor Selection"},
        })
    })

    utils.Logger.WithField("port", port).Info("SCM service starting")
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