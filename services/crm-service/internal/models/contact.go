package models

import (
	"time"
)

// Contact represents a customer contact
type Contact struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	FirstName   string    `json:"first_name" gorm:"not null"`
	LastName    string    `json:"last_name" gorm:"not null"`
	Email       string    `json:"email" gorm:"uniqueIndex;not null"`
	Phone       string    `json:"phone"`
	Company     string    `json:"company"`
	Position    string    `json:"position"`
	Department  string    `json:"department"`
	Address     string    `json:"address"`
	City        string    `json:"city"`
	State       string    `json:"state"`
	Country     string    `json:"country"`
	PostalCode  string    `json:"postal_code"`
	LeadSource  string    `json:"lead_source"`
	Status      string    `json:"status" gorm:"default:'active'"`
	Notes       string    `json:"notes"`
	Tags        string    `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ContactCreateRequest represents the request to create a new contact
type ContactCreateRequest struct {
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Phone       string `json:"phone"`
	Company     string `json:"company"`
	Position    string `json:"position"`
	Department  string `json:"department"`
	Address     string `json:"address"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	PostalCode  string `json:"postal_code"`
	LeadSource  string `json:"lead_source"`
	Notes       string `json:"notes"`
	Tags        string `json:"tags"`
}

// ContactUpdateRequest represents the request to update a contact
type ContactUpdateRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email" binding:"email"`
	Phone       string `json:"phone"`
	Company     string `json:"company"`
	Position    string `json:"position"`
	Department  string `json:"department"`
	Address     string `json:"address"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	PostalCode  string `json:"postal_code"`
	LeadSource  string `json:"lead_source"`
	Status      string `json:"status"`
	Notes       string `json:"notes"`
	Tags        string `json:"tags"`
} 