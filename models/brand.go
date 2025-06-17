package models

import (
	"gorm.io/gorm"
)

type Brand struct {
	gorm.Model
	Name        string `gorm:"not null;unique" json:"name"`
	Image       string `gorm:"not null" json:"image"`
	Slug        string `gorm:"not null;unique" json:"slug"`
	IsDisplayed bool   `gorm:"default:true" json:"is_displayed"`
}
