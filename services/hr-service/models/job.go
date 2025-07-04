package models

import (
	"time"
)



// Job represents a specific job title or position within the organization.
type Job struct {
	ID          string    `json:"id"`           // Unique identifier for the job
	Title       string    `json:"title"`        // Job title (e.g., "Software Engineer", "Marketing Manager")
	Description string    `json:"description"`  // A brief description of the job responsibilities
	DepartmentID string   `json:"department_id"` // ID of the department this job belongs to
	BaseSalary  float64   `json:"base_salary"`  // Base salary for this job role
	CreatedAt   time.Time `json:"created_at"`   // Timestamp when the job record was created
	UpdatedAt   time.Time `json:"updated_at"`   // Timestamp when the job record was last updated
}