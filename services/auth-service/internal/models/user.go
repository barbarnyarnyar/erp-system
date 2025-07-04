// File: services/auth-service/internal/models/user.go
package models

import (
	"time"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"unique;not null;size:50"`
	Email     string         `json:"email" gorm:"unique;not null;size:100"`
	Password  string         `json:"-" gorm:"not null;size:255"` // Never return in JSON
	FirstName string         `json:"first_name" gorm:"size:50"`
	LastName  string         `json:"last_name" gorm:"size:50"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	LastLogin *time.Time     `json:"last_login,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Roles       []Role       `json:"roles" gorm:"many2many:user_roles;"`
	Sessions    []Session    `json:"-" gorm:"foreignKey:UserID"`
	Permissions []Permission `json:"permissions" gorm:"many2many:user_permissions;"`
}

type Role struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"unique;not null;size:50"`
	Description string       `json:"description" gorm:"size:255"`
	IsActive    bool         `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`

	// Relationships
	Users       []User       `json:"-" gorm:"many2many:user_roles;"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
}

type Permission struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"unique;not null;size:100"`
	Resource    string    `json:"resource" gorm:"not null;size:50"` // e.g., "invoices", "payments"
	Action      string    `json:"action" gorm:"not null;size:50"`   // e.g., "read", "write", "delete"
	Service     string    `json:"service" gorm:"not null;size:20"`  // e.g., "fm", "hr", "scm"
	Description string    `json:"description" gorm:"size:255"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Users []User `json:"-" gorm:"many2many:user_permissions;"`
	Roles []Role `json:"-" gorm:"many2many:role_permissions;"`
}

type Session struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Token     string    `json:"token" gorm:"unique;not null;size:255"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	IPAddress string    `json:"ip_address" gorm:"size:45"`
	UserAgent string    `json:"user_agent" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// User methods
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) HasPermission(service, resource, action string) bool {
	// Check direct permissions
	for _, perm := range u.Permissions {
		if perm.Service == service && perm.Resource == resource && perm.Action == action {
			return true
		}
	}
	
	// Check role permissions
	for _, role := range u.Roles {
		if !role.IsActive {
			continue
		}
		for _, perm := range role.Permissions {
			if perm.Service == service && perm.Resource == resource && perm.Action == action {
				return true
			}
		}
	}
	
	return false
}

func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName && role.IsActive {
			return true
		}
	}
	return false
}

func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// Table names
func (User) TableName() string       { return "users" }
func (Role) TableName() string       { return "roles" }
func (Permission) TableName() string { return "permissions" }
func (Session) TableName() string    { return "sessions" }