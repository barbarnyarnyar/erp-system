// --- routes/book_routes.go ---
// This file defines the API routes for book resources and links them to handlers.
package routes

import (
	"your_project_name/handlers" // Replace with your actual module name
	"github.com/gin-gonic/gin"
)

// SetupBookRoutes configures the book-related API endpoints.
func SetupBookRoutes(router *gin.Engine) {
	bookRoutes := router.Group("/books")
	{
		bookRoutes.GET("/", handlers.GetBooks)
		bookRoutes.GET("/:id", handlers.GetBookByID)
		bookRoutes.POST("/", handlers.CreateBook)
		bookRoutes.PUT("/:id", handlers.UpdateBook)
		bookRoutes.DELETE("/:id", handlers.DeleteBook)
	}
}