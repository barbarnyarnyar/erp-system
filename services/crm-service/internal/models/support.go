package models

import (
	"time"
)

// SupportTicket represents a customer support ticket
type SupportTicket struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	ContactID    uint       `json:"contact_id" gorm:"not null"`
	TicketNumber string     `json:"ticket_number" gorm:"uniqueIndex;not null"`
	Subject      string     `json:"subject" gorm:"not null"`
	Description  string     `json:"description"`
	Category     string     `json:"category"`
	Priority     string     `json:"priority" gorm:"default:'medium'"`
	Status       string     `json:"status" gorm:"default:'open'"`
	AssignedTo   *uint      `json:"assigned_to"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ClosedAt     *time.Time `json:"closed_at"`

	// Relations
	Contact Contact `json:"contact" gorm:"foreignKey:ContactID"`
}

// SupportResponse represents a response to a support ticket
type SupportResponse struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	TicketID   uint      `json:"ticket_id" gorm:"not null"`
	Response   string    `json:"response" gorm:"not null"`
	IsInternal bool      `json:"is_internal" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at"`

	// Relations
	Ticket SupportTicket `json:"ticket" gorm:"foreignKey:TicketID"`
}

// SupportTicketCreateRequest represents the request to create a new support ticket
type SupportTicketCreateRequest struct {
	ContactID   uint   `json:"contact_id" binding:"required"`
	Subject     string `json:"subject" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Priority    string `json:"priority"`
}

// SupportResponseCreateRequest represents the request to create a new support response
type SupportResponseCreateRequest struct {
	TicketID   uint   `json:"ticket_id" binding:"required"`
	Response   string `json:"response" binding:"required"`
	IsInternal bool   `json:"is_internal"`
}
