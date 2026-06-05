package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/erp-system/fm-service/internal/api/handlers"
	"github.com/erp-system/fm-service/internal/config"
)

func SetupRoutes(
	router *gin.Engine,
	cfg *config.Config,
	accHandler *handlers.AccountHandler,
	txHandler *handlers.TransactionHandler,
	repHandler *handlers.ReportHandler,
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

		// Transaction routes
		transactions := v1.Group("/transactions")
		{
			transactions.GET("", txHandler.GetTransactions)
			transactions.POST("", txHandler.CreateTransaction)
			transactions.GET("/:id", txHandler.GetTransaction)
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