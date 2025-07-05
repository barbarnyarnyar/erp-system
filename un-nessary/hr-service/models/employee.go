package models

import (
	"time"
)

// EmployeeStatus represents the current employment status of an employee.
type EmployeeStatus string

const (
	Active    EmployeeStatus = "Active"
	Inactive  EmployeeStatus = "Inactive"
	OnLeave   EmployeeStatus = "On Leave"
	Terminated EmployeeStatus = "Terminated"
)

// Employee represents a single employee in the organization.
// This is the central entity for the HR/HCM service.
type Employee struct {
	ID            string         `json:"id"`             // Unique identifier for the employee
	FirstName     string         `json:"first_name"`     // Employee's first name
	LastName      string         `json:"last_name"`      // Employee's last name
	Email         string         `json:"email"`          // Employee's primary email address (unique)
	PhoneNumber   string         `json:"phone_number"`   // Employee's phone number
	Address       string         `json:"address"`        // Employee's residential address
	DateOfBirth   time.Time      `json:"date_of_birth"`  // Employee's date of birth
	HireDate      time.Time      `json:"hire_date"`      // Date the employee was hired
	TerminationDate *time.Time     `json:"termination_date,omitempty"` // Date of termination, if applicable
	Status        EmployeeStatus `json:"status"`         // Current employment status
	JobID         string         `json:"job_id"`         // ID of the job/position held by the employee
	DepartmentID  string         `json:"department_id"`  // ID of the department the employee belongs to
	ManagerID     string         `json:"manager_id"`     // ID of the employee's manager (self-referencing)
	CreatedAt     time.Time      `json:"created_at"`     // Timestamp when the employee record was created
	UpdatedAt     time.Time      `json:"updated_at"`     // Timestamp when the employee record was last updated
}
