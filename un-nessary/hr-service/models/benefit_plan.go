// BenefitPlan represents a type of benefit offered by the company.
type BenefitPlan struct {
	ID          string    `json:"id"`           // Unique identifier for the benefit plan
	Name        string    `json:"name"`         // Name of the plan (e.g., "Medical PPO", "Dental HMO", "401k")
	Description string    `json:"description"`  // Description of the plan
	Provider    string    `json:"provider"`     // Provider of the benefit (e.g., "Blue Cross", "Fidelity")
	Cost        float64   `json:"cost"`         // Cost of the plan to the company (per employee or flat)
	Type        string    `json:"type"`         // Type of benefit (e.g., "Health", "Retirement", "Paid Time Off")
	CreatedAt   time.Time `json:"created_at"`   // Timestamp when the benefit plan was created
	UpdatedAt   time.Time `json:"updated_at"`   // Timestamp when the benefit plan was last updated
}
