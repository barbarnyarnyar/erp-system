// services/finance/internal/handlers/gl_handler.go
package handlers

import (
    "github.com/gin-gonic/gin"
    "github.com/your-org/erp-microservices/shared/utils"
)

// GetGLAccounts returns chart of accounts
func GetGLAccounts(c *gin.Context) {
    data := gin.H{
        "message":     "Finance Service - General Ledger",
        "endpoint":    "GET /api/v1/finance/gl",
        "description": "Chart of Accounts Management",
        "accounts": []gin.H{
            {
                "account_code": "1100",
                "account_name": "Cash and Cash Equivalents",
                "account_type": "ASSET",
                "balance":      125000.00,
            },
            {
                "account_code": "1200", 
                "account_name": "Accounts Receivable",
                "account_type": "ASSET",
                "balance":      45000.00,
            },
            {
                "account_code": "4100",
                "account_name": "Revenue",
                "account_type": "REVENUE",
                "balance":      350000.00,
            },
        },
    }
    utils.SuccessResponse(c, data)
}

// CreateJournalEntry creates a new journal entry
func CreateJournalEntry(c *gin.Context) {
    data := gin.H{
        "message":     "Finance Service - Create Journal Entry",
        "endpoint":    "POST /api/v1/finance/gl/journal-entries",
        "description": "Create new journal entry for financial transactions",
        "example": gin.H{
            "journal_entry_id": "je_12345",
            "reference":        "JE-2024-001",
            "status":          "POSTED",
            "total_debit":     5000.00,
            "total_credit":    5000.00,
        },
    }
    utils.SuccessResponse(c, data)
}

// GetJournalEntries retrieves journal entries
func GetJournalEntries(c *gin.Context) {
    data := gin.H{
        "message":     "Finance Service - Journal Entries",
        "endpoint":    "GET /api/v1/finance/gl/journal-entries",
        "description": "Retrieve journal entries with filtering",
        "entries": []gin.H{
            {
                "journal_entry_id": "je_12345",
                "reference":        "JE-2024-001",
                "description":      "Monthly depreciation",
                "posted_date":      "2024-01-31",
                "total_amount":     5000.00,
                "status":          "POSTED",
            },
        },
    }
    utils.SuccessResponse(c, data)
}