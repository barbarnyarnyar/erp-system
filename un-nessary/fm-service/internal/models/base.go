// File: services/financial-management/models/base.go
package models

import (
	"time"
	"gorm.io/gorm"
)

// BaseModel provides common fields for all entities
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Auditable adds audit trail fields
type Auditable struct {
	CreatedBy string `json:"created_by" gorm:"size:50"`
	UpdatedBy string `json:"updated_by" gorm:"size:50"`
}

// Trackable adds source tracking for integrations
type Trackable struct {
	SourceService string `json:"source_service" gorm:"size:20"` // HR, SCM, CRM, PM, M
	SourceID      string `json:"source_id" gorm:"size:50"`      // External service ID
	Reference     string `json:"reference" gorm:"size:100"`     // External reference
}

// Helper functions for pointer types
func StringPtr(s string) *string {
	return &s
}

func UintPtr(u uint) *uint {
	return &u
}

func Float64Ptr(f float64) *float64 {
	return &f
}

func IntPtr(i int) *int {
	return &i
}

// Common interfaces for extensibility
type Validator interface {
	Validate() error
}

type Calculator interface {
	Calculate()
}

type Balancer interface {
	UpdateBalance()
}

// Common constants
const (
	DefaultCurrency = "USD"
	DefaultExchangeRate = 1.0
)