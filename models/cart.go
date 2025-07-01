package models

import (
	"gorm.io/gorm"
)

type Cart struct {
	gorm.Model
	UserID *uint      `json:"user_id"` // nullable for guests
	User   *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Items  []CartItem `json:"items"`
}

type CartItem struct {
	gorm.Model
	CartID uint  `json:"cart_id"`
	Cart   *Cart `json:"-" gorm:"foreignKey:CartID"`

	// New variant-based structure
	ProductVariantID uint            `json:"product_variant_id"`
	ProductVariant   *ProductVariant `json:"product_variant" gorm:"foreignKey:ProductVariantID"`

	// Legacy field for backward compatibility (will be removed later)
	ProductID *uint    `json:"product_id,omitempty"`
	Product   *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`

	Quantity   int     `json:"quantity"`
	PriceType  string  `json:"price_type"`  // "customer" or "b2b"
	UnitPrice  float64 `json:"unit_price"`  // Price at time of adding to cart
	TotalPrice float64 `json:"total_price"` // UnitPrice * Quantity
}
