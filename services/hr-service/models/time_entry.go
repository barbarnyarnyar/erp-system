package models

import (
	"time"
)	

// TimeEntry represents an employee's recorded work time or leave.
type TimeEntry struct {
	ID         string    `json:"id"`           // Unique identifier for the time entry
	EmployeeID string    `json:"employee_id"`  // ID of the employee
	Date       time.Time `json:"date"`         // Date of the time entry
	Hours      float64   `json:"hours"`        // Number of hours worked or taken as leave
	Type       string    `json:"type"`         // Type of entry (e.g., "Work", "Vacation", "Sick Leave")
	Status     string    `json:"status"`       // Status (e.g., "Pending", "Approved", "Rejected")
	Notes      string    `json:"notes"`        // Any relevant notes for the entry
	CreatedAt  time.Time `json:"created_at"`   // Timestamp when the time entry was created
	UpdatedAt  time.Time `json:"updated_at"`   // Timestamp when the time entry was last updated
}