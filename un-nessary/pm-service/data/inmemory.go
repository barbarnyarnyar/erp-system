// --- data/inmemory.go ---
// This file simulates a database by holding the in-memory data.
// In a real application, this would be replaced by database interactions.
package data

import "your_project_name/models" // Replace with your actual module name

// Books is an in-memory slice to simulate a database.
var Books = []models.Book{
	{ID: 1, Title: "The Hitchhiker's Guide to the Galaxy", Author: "Douglas Adams"},
	{ID: 2, Title: "1984", Author: "George Orwell"},
	{ID: 3, Title: "Pride and Prejudice", Author: "Jane Austen"},
}

// NextBookID generates a simple incrementing ID for new books.
// In a real application, this would be handled by the database.
func NextBookID() int {
	return len(Books) + 1
}
