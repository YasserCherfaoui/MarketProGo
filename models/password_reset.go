package models

import (
	"time"

	"gorm.io/gorm"
)

// PasswordResetToken stores password reset requests securely
type PasswordResetToken struct {
	gorm.Model
	UserID    uint       `json:"user_id"`
	User      *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	TokenHash string     `json:"-" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"index;not null"`
	UsedAt    *time.Time `json:"used_at"`
}
