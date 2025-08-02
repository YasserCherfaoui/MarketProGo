package models

import (
	"gorm.io/gorm"
)

// Wishlist represents a user's wishlist
type Wishlist struct {
	gorm.Model
	UserID *uint          `json:"user_id"`
	User   *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Items  []WishlistItem `json:"items"`
}

// WishlistItem represents an item in a user's wishlist
type WishlistItem struct {
	gorm.Model
	WishlistID uint      `json:"wishlist_id"`
	Wishlist   *Wishlist `json:"-" gorm:"foreignKey:WishlistID"`

	// Product variant reference
	ProductVariantID uint            `json:"product_variant_id"`
	ProductVariant   *ProductVariant `json:"product_variant" gorm:"foreignKey:ProductVariantID"`

	// Legacy field for backward compatibility
	ProductID *uint    `json:"product_id,omitempty"`
	Product   *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`

	// Additional metadata
	Notes    string `json:"notes"`     // User notes about the item
	Priority int    `json:"priority"`  // Priority level (1-5, 5 being highest)
	IsPublic bool   `json:"is_public"` // Whether the item is visible to others
}
