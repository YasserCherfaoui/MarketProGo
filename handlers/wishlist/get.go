package wishlist

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

// GetWishlist retrieves the user's wishlist with all items
func (h *WishlistHandler) GetWishlist(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "wishlist/get", "User not authenticated")
		return
	}
	uid := userID.(uint)

	var wishlist models.Wishlist
	err := h.db.Where("user_id = ?", uid).
		Preload("Items.ProductVariant.Product").
		Preload("Items.ProductVariant.Product.Brand").
		Preload("Items.ProductVariant.Product.Categories").
		Preload("Items.ProductVariant.Product.Images").
		Preload("Items.ProductVariant.InventoryItems").
		First(&wishlist).Error

	if err != nil {
		// Create empty wishlist if it doesn't exist
		wishlist = models.Wishlist{
			UserID: &uid,
			Items:  []models.WishlistItem{},
		}
		if err := h.db.Create(&wishlist).Error; err != nil {
			response.GenerateInternalServerErrorResponse(c, "wishlist/get", "Failed to create wishlist")
			return
		}
	}

	response.GenerateSuccessResponse(c, "Wishlist retrieved successfully", wishlist)
}
