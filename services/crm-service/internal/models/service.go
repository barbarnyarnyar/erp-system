package models

import (
	"time"
)

// Service represents a service offering
type Service struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Price       float64   `json:"price"`
	Currency    string    `json:"currency" gorm:"default:'USD'"`
	Duration    int       `json:"duration"` // in minutes
	Status      string    `json:"status" gorm:"default:'active'"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ServiceRequest represents a service request from a customer
type ServiceRequest struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	ContactID   uint       `json:"contact_id" gorm:"not null"`
	ServiceID   uint       `json:"service_id" gorm:"not null"`
	Title       string     `json:"title" gorm:"not null"`
	Description string     `json:"description"`
	Priority    string     `json:"priority" gorm:"default:'medium'"`
	Status      string     `json:"status" gorm:"default:'open'"`
	RequestedAt time.Time  `json:"requested_at"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Notes       string     `json:"notes"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relations
	Contact Contact `json:"contact" gorm:"foreignKey:ContactID"`
	Service Service `json:"service" gorm:"foreignKey:ServiceID"`
}

// ServiceCreateRequest represents the request to create a new service
type ServiceCreateRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Duration    int     `json:"duration"`
}

// ServiceRequestCreateRequest represents the request to create a new service request
type ServiceRequestCreateRequest struct {
	ContactID   uint       `json:"contact_id" binding:"required"`
	ServiceID   uint       `json:"service_id" binding:"required"`
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	Notes       string     `json:"notes"`
}
