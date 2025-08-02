package wishlist

import (
	"gorm.io/gorm"
)

type WishlistHandler struct {
	db *gorm.DB
}

func NewWishlistHandler(db *gorm.DB) *WishlistHandler {
	return &WishlistHandler{db: db}
}
