package routes

import (
	"github.com/erp-system/fm-service/internal/api/handlers"
	"github.com/erp-system/fm-service/internal/config"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	router *gin.Engine,
	cfg *config.Config,
	accHandler *handlers.AccountHandler,
	txHandler *handlers.TransactionHandler,
	repHandler *handlers.ReportHandler,
	invHandler *handlers.InvoiceHandler,
	payHandler *handlers.PaymentHandler,
	billHandler *handlers.VendorBillHandler,
) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "fm-service",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Account routes
		accounts := v1.Group("/accounts")
		{
			accounts.GET("", accHandler.GetAccounts)
			accounts.POST("", accHandler.CreateAccount)
			accounts.GET("/:id", accHandler.GetAccount)
			accounts.PUT("/:id", accHandler.UpdateAccount)
			accounts.DELETE("/:id", accHandler.DeleteAccount)
			accounts.GET("/:id/balance", accHandler.GetAccountBalance)
		}

		// Journal Entries routes
		journalEntries := v1.Group("/journal-entries")
		{
			journalEntries.GET("", txHandler.GetTransactions)
			journalEntries.POST("", txHandler.CreateTransaction)
			journalEntries.GET("/:id", txHandler.GetTransaction)
			journalEntries.PUT("/:id", txHandler.UpdateTransaction)
			journalEntries.DELETE("/:id", txHandler.DeleteTransaction)
		}

		// Invoices routes
		invoices := v1.Group("/invoices")
		{
			invoices.GET("", invHandler.GetInvoices)
			invoices.POST("", invHandler.CreateInvoice)
			invoices.GET("/:id", invHandler.GetInvoice)
			invoices.PUT("/:id", invHandler.UpdateInvoice)
			invoices.DELETE("/:id", invHandler.DeleteInvoice)
			invoices.POST("/:id/send", invHandler.SendInvoice)
			invoices.GET("/:id/lines", invHandler.GetInvoiceLines)
		}

		// Payments routes
		payments := v1.Group("/payments")
		{
			payments.GET("", payHandler.GetPayments)
			payments.POST("", payHandler.RecordPayment)
			payments.GET("/:id", payHandler.GetPayment)
		}


		// Bank Statements routes
		bankStatements := v1.Group("/bank-statements")
		{
			bankStatements.GET("/:id/lines", payHandler.GetBankStatementLines)
		}

		// Vendor Bills routes
		vendorBills := v1.Group("/vendor-bills")
		{
			vendorBills.GET("", billHandler.GetVendorBills)
			vendorBills.POST("", billHandler.CreateVendorBill)
			vendorBills.GET("/:id/lines", billHandler.GetVendorBillLines)
		}

		// Reports routes
		reports := v1.Group("/reports")
		{
			reports.GET("/balance-sheet", repHandler.GetBalanceSheet)
			reports.GET("/income-statement", repHandler.GetIncomeStatement)
			reports.GET("/cash-flow", repHandler.GetCashFlow)
		}
	}
}
