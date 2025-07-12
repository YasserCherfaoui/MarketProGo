package models

import (
	"time"

	"gorm.io/gorm"
)

type Promotion struct {
	gorm.Model
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	Image       string    `gorm:"not null" json:"image"` // File ID or URL
	ButtonText  string    `json:"button_text"`
	ButtonLink  string    `json:"button_link"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`

	// Optional links
	ProductID  *uint     `json:"product_id"`
	Product    *Product  `json:"product" gorm:"foreignKey:ProductID"`
	CategoryID *uint     `json:"category_id"`
	Category   *Category `json:"category" gorm:"foreignKey:CategoryID"`
	BrandID    *uint     `json:"brand_id"`
	Brand      *Brand    `json:"brand" gorm:"foreignKey:BrandID"`
}
