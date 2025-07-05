// --- handlers/book_handler.go ---
// This file contains the business logic and handlers for book-related operations.
package handlers

import (
	"net/http"
	"strconv"

	"your_project_name/data"   // Replace with your actual module name
	"your_project_name/models" // Replace with your actual module name

	"github.com/gin-gonic/gin"
)

// GetBooks handles GET requests to retrieve all books.
// @Summary Get all books
// @Description Get a list of all books
// @Produce json
// @Success 200 {array} models.Book
// @Router /books [get]
func GetBooks(c *gin.Context) {
	c.JSON(http.StatusOK, data.Books)
}

// GetBookByID handles GET requests to retrieve a single book by its ID.
// @Summary Get a book by ID
// @Description Get a single book by its ID
// @Produce json
// @Param id path int true "Book ID"
// @Success 200 {object} models.Book
// @Failure 400 {object} gin.H "Invalid book ID"
// @Failure 404 {object} gin.H "Book not found"
// @Router /books/{id} [get]
func GetBookByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	for _, book := range data.Books {
		if book.ID == id {
			c.JSON(http.StatusOK, book)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "Book not found"})
}

// CreateBook handles POST requests to create a new book.
// @Summary Create a new book
// @Description Add a new book to the collection
// @Accept json
// @Produce json
// @Param book body models.Book true "Book object to be created"
// @Success 201 {object} models.Book
// @Failure 400 {object} gin.H "Invalid input"
// @Router /books [post]
func CreateBook(c *gin.Context) {
	var newBook models.Book
	// Use ShouldBindJSON for automatic validation based on 'binding' tags in struct
	if err := c.ShouldBindJSON(&newBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newBook.ID = data.NextBookID()
	data.Books = append(data.Books, newBook)

	c.JSON(http.StatusCreated, newBook)
}

// UpdateBook handles PUT requests to update an existing book.
// @Summary Update an existing book
// @Description Update a book's title and author by its ID
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Param book body models.Book true "Updated book object"
// @Success 200 {object} models.Book
// @Failure 400 {object} gin.H "Invalid input"
// @Failure 404 {object} gin.H "Book not found"
// @Router /books/{id} [put]
func UpdateBook(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	var updatedBook models.Book
	if err := c.ShouldBindJSON(&updatedBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	found := false
	for i, book := range data.Books {
		if book.ID == id {
			data.Books[i].Title = updatedBook.Title
			data.Books[i].Author = updatedBook.Author
			c.JSON(http.StatusOK, data.Books[i])
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"message": "Book not found"})
	}
}

// DeleteBook handles DELETE requests to remove a book.
// @Summary Delete a book
// @Description Delete a book by its ID
// @Produce json
// @Param id path int true "Book ID"
// @Success 204 "No Content"
// @Failure 400 {object} gin.H "Invalid input"
// @Failure 404 {object} gin.H "Book not found"
// @Router /books/{id} [delete]
func DeleteBook(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	foundIndex := -1
	for i, book := range data.Books {
		if book.ID == id {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Book not found"})
		return
	}

	// Remove the book from the slice
	data.Books = append(data.Books[:foundIndex], data.Books[foundIndex+1:]...)
	c.Status(http.StatusNoContent) // 204 No Content for successful deletion
}