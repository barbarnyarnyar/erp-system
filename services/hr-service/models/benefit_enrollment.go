// BenefitEnrollment represents an employee's enrollment in a specific benefit plan.
type BenefitEnrollment struct {
	ID         string    `json:"id"`           // Unique identifier for the benefit enrollment
	EmployeeID string    `json:"employee_id"`  // ID of the employee
	BenefitPlanID string `json:"benefit_plan_id"` // ID of the benefit plan (e.g., health insurance, 401k)
	EnrollmentDate time.Time `json:"enrollment_date"` // Date of enrollment
	Status     string    `json:"status"`       // Status (e.g., "Active", "Inactive", "Pending")
	ContributionAmount float64 `json:"contribution_amount"` // Employee's contribution
	CreatedAt  time.Time `json:"created_at"`   // Timestamp when the enrollment record was created
	UpdatedAt  time.Time `json:"updated_at"`   // Timestamp when the enrollment record was last updated
}