// services/finance/cmd/main.go
package main

import (
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/your-org/erp-microservices/shared/utils"
)

// GetARDashboard returns accounts receivable dashboard
func GetARDashboard(c *gin.Context) {
    data := gin.H{
        "message":     "Finance Service - Accounts Receivable Dashboard",
        "endpoint":    "GET /api/v1/finance/ar",
        "description": "Accounts Receivable overview and metrics",
        "summary": gin.H{
            "total_outstanding": 85000.00,
            "current":          70000.00,
            "30_days":          10000.00,
            "60_days":          3000.00,
            "90_plus":          2000.00,
            "customer_count":   25,
        },
    }
    utils.SuccessResponse(c, data)
}

// CreateCustomerInvoice creates a customer invoice
func CreateCustomerInvoice(c *gin.Context) {
    data := gin.H{
        "message":     "Finance Service - Create Customer Invoice",
        "endpoint":    "POST /api/v1/finance/ar/invoices",
        "description": "Create customer invoice for accounts receivable",
        "example": gin.H{
            "invoice_id":     "ar_inv_12345",
            "customer_id":    "CUST-001",
            "invoice_number": "INV-2024-001",
            "total_amount":   5500.00,
            "status":        "SENT",
        },
    }
    utils.SuccessResponse(c, data)
}

// GetCustomerInvoices retrieves customer invoices
func GetCustomerInvoices(c *gin.Context) {
    data := gin.H{
        "message":     "Finance Service - Customer Invoices",
        "endpoint":    "GET /api/v1/finance/ar/invoices",
        "description": "Retrieve customer invoices with aging and status",
    }
    utils.SuccessResponse(c, data)
}

// RecordPaymentReceipt records customer payment
func RecordPaymentReceipt(c *gin.Context) {
    data := gin.H{
        "message":     "Finance Service - Record Payment Receipt",
        "endpoint":    "POST /api/v1/finance/ar/receipts",
        "description": "Record customer payment receipt",
    }
    utils.SuccessResponse(c, data)
}