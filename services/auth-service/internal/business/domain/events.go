package domain

import (
	"context"
	"time"
)

type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload interface{}) error
}

type UserEventPayload struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	IsActive  bool      `json:"is_active"`
	Timestamp time.Time `json:"timestamp"`
}

type UserRoleEventPayload struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	RoleID     string    `json:"role_id"`
	AssignedBy string    `json:"assigned_by"`
	Timestamp  time.Time `json:"timestamp"`
}

type UserStoreEventPayload struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	StoreID   string    `json:"store_id"`
	Timestamp time.Time `json:"timestamp"`
}

type PasswordChangedEventPayload struct {
	UserID    string    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

// HREmployeeTerminatedEvent is the cross-service payload published by HR when
// an employee is terminated. Per the cross-service @reference convention
// (see master PRD 2.10), EmployeeID is treated as the Auth User ID for
// deactivation lookup.
type HREmployeeTerminatedEvent struct {
	EmployeeID string    `json:"employee_id"`
	TermDate   time.Time `json:"term_date"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"timestamp"`
}
