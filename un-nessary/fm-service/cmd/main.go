// services/finance/cmd/main.go
package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "erp-system/shared/utils"
)

func main() {
    serviceName := getEnv("SERVICE_NAME", "finance")
    port := getEnv("PORT", "8001")
    
    utils.InitLogger(serviceName)
    utils.Logger.Info("Starting Finance Service")

    r := gin.Default()

    // Health check
    r.GET("/health", func(c *gin.Context) {
        utils.SuccessResponse(c, "Finance Service is healthy", gin.H{
            "service": serviceName,
            "version": "1.0.0",
        })
    })

    // Hello World endpoints
    r.GET("/", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Finance Service!", gin.H{
            "service": "finance",
            "domain": "Financial Management",
        })
    })

    r.GET("/api/v1/finance/hello", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Finance API!", gin.H{
            "endpoints": []string{
                "GET /api/v1/finance/gl - General Ledger",
                "GET /api/v1/finance/ap - Accounts Payable", 
                "GET /api/v1/finance/ar - Accounts Receivable",
                "GET /api/v1/finance/reports - Financial Reports",
            },
        })
    })

    r.GET("/api/v1/finance/gl", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from General Ledger!", gin.H{
            "module": "General Ledger",
            "features": []string{"Chart of Accounts", "Journal Entries", "Trial Balance"},
        })
    })

    r.GET("/api/v1/finance/ap", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Accounts Payable!", gin.H{
            "module": "Accounts Payable",
            "features": []string{"Vendor Invoices", "Payments", "Aging Reports"},
        })
    })

    r.GET("/api/v1/finance/ar", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Accounts Receivable!", gin.H{
            "module": "Accounts Receivable",
            "features": []string{"Customer Invoices", "Collections", "Credit Management"},
        })
    })

    r.GET("/api/v1/finance/reports", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from Financial Reports!", gin.H{
            "module": "Financial Reports",
            "features": []string{"Balance Sheet", "Income Statement", "Cash Flow"},
        })
    })

    utils.Logger.WithField("port", port).Info("Finance service starting")
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
