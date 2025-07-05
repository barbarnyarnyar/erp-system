// --- models/book.go ---
// This file defines the structure for the Book resource.
package models

// Book represents a book in our database.
type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title" binding:"required"` // Added binding for validation
	Author string `json:"author" binding:"required"`
}
