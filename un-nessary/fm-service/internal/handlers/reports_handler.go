// services/finance/internal/handlers/reports_handler.go
package handlers

import (
    "github.com/gin-gonic/gin"
    "github.com/your-org/erp-microservices/shared/utils"
)

// GetBalanceSheet returns balance sheet
func GetBalanceSheet(c *gin.Context) {
    data := gin.H{
        "message":     "Finance Service - Balance Sheet",
        "endpoint":    "GET /api/v1/finance/reports/balance-sheet",
        "description": "Balance sheet financial report",
        "balance_sheet": gin.H{
            "assets": gin.H{
                "current_assets": gin.H{
                    "cash":                125000.00,
                    "accounts_receivable": 45000.00,
                    "inventory":          75000.00,
                    "total":              245000.00,
                },
                "fixed_assets": gin.H{
                    "equipment":              200000.00,
                    "accumulated_depreciation": -50000.00,
                    "total":                  150000.00,
                },
                "total_assets": 395000.00,
            },
            "liabilities": gin.H{
                "current_liabilities": gin.H{
                    "accounts_payable": 35000.00,
                    "accrued_expenses": 15000.00,
                    "total":           50000.00,
                },
                "total_liabilities": 50000.00,
            },
            "equity": gin.H{
                "retained_earnings": 345000.00,
                "total_equity":     345000.00,
            },
        },
    }
    utils.SuccessResponse(c, data)
}

// GetIncomeStatement returns income statement
func GetIncomeStatement(c *gin.Context) {
    data := gin.H{
        "message":     "Finance Service - Income Statement",
        "endpoint":    "GET /api/v1/finance/reports/income-statement",
        "description": "Income statement financial report",
    }
    utils.SuccessResponse(c, data)
}

// GetCashFlow returns cash flow statement
func GetCashFlow(c *gin.Context) {
    data := gin.H{
        "message":     "Finance Service - Cash Flow Statement",
        "endpoint":    "GET /api/v1/finance/reports/cash-flow",
        "description": "Cash flow statement financial report",
    }
    utils.SuccessResponse(c, data)
}