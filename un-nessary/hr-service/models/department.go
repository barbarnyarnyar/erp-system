package models

import (
	"time"
)

// Department represents an organizational department.
type Department struct {
	ID        string    `json:"id"`           // Unique identifier for the department
	Name      string    `json:"name"`         // Department name (e.g., "Finance", "Sales", "IT")
	ManagerID string    `json:"manager_id"`   // ID of the employee who manages this department
	CreatedAt time.Time `json:"created_at"`   // Timestamp when the department record was created
	UpdatedAt time.Time `json:"updated_at"`   // Timestamp when the department record was last updated
}
