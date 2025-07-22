package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/erp-system/fm-service/internal/api/handlers"
	"github.com/erp-system/fm-service/internal/config"
)

func SetupRoutes(router *gin.Engine, cfg *config.Config) {
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
			accounts.GET("", handlers.GetAccounts)
			accounts.POST("", handlers.CreateAccount)
			accounts.GET("/:id", handlers.GetAccount)
			accounts.PUT("/:id", handlers.UpdateAccount)
			accounts.DELETE("/:id", handlers.DeleteAccount)
			accounts.GET("/:id/balance", handlers.GetAccountBalance)
		}

		// Transaction routes
		transactions := v1.Group("/transactions")
		{
			transactions.GET("", handlers.GetTransactions)
			transactions.POST("", handlers.CreateTransaction)
			transactions.GET("/:id", handlers.GetTransaction)
			transactions.POST("/:id/post", handlers.PostTransaction)
			transactions.POST("/:id/reverse", handlers.ReverseTransaction)
		}

		// Reports routes
		reports := v1.Group("/reports")
		{
			reports.GET("/balance-sheet", handlers.GetBalanceSheet)
			reports.GET("/income-statement", handlers.GetIncomeStatement)
			reports.GET("/cash-flow", handlers.GetCashFlow)
		}
	}
}