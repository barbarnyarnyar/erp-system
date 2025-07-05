// shared/models/common.go
package models

import (
    "time"
    
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type BaseModel struct {
    ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
    if base.ID == "" {
        base.ID = uuid.New().String()
    }
    return nil
}