package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents the user model
type User struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	Username  string         `json:"username" gorm:"uniqueIndex;size:32"`
	Email     string         `json:"email" gorm:"size:128"`
	Password  string         `json:"-" gorm:"size:128"` // "-" means this field won't be included in JSON
	Status    int            `json:"status" gorm:"default:1"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}
