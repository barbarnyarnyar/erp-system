// --- main.go ---
// This file serves as the entry point for the application,
// initializes the Gin router, and sets up the routes.
package main

import (
	"fmt"
	"pm-serivce/routes" // Replace with your actual module name
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Gin router with default middleware
	router := gin.Default()

	// Setup routes for the API
	routes.SetupBookRoutes(router)

	// Run the server on port 8080
	// You can access it at http://localhost:8080
	port := ":8080"
	fmt.Printf("Server running on http://localhost%s\n", port)
	router.Run(port)
}