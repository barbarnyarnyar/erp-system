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
