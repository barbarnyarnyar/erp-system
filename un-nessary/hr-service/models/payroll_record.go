package models

import (
	"time"
)


// PayrollRecord represents a single payroll entry for an employee for a specific period.
type PayrollRecord struct {
	ID          string    `json:"id"`           // Unique identifier for the payroll record
	EmployeeID  string    `json:"employee_id"`  // ID of the employee
	PayPeriodStart time.Time `json:"pay_period_start"` // Start date of the pay period
	PayPeriodEnd time.Time `json:"pay_period_end"`   // End date of the pay period
	GrossPay    float64   `json:"gross_pay"`    // Total earnings before deductions
	NetPay      float64   `json:"net_pay"`      // Earnings after all deductions
	Deductions  float64   `json:"deductions"`   // Total deductions (taxes, benefits, etc.)
	Taxes       float64   `json:"taxes"`        // Total taxes withheld
	Status      string    `json:"status"`       // Status (e.g., "Calculated", "Paid", "Pending")
	PaymentDate time.Time `json:"payment_date"` // Date the payment was made
	CreatedAt   time.Time `json:"created_at"`   // Timestamp when the payroll record was created
	UpdatedAt   time.Time `json:"updated_at"`   // Timestamp when the payroll record was last updated
}
