package wishlist

import (
	"strconv"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

// RemoveItem removes an item from the user's wishlist
func (h *WishlistHandler) RemoveItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "wishlist/remove_item", "User not authenticated")
		return
	}
	uid := userID.(uint)

	// Get item ID from URL parameter
	itemIDStr := c.Param("id")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "wishlist/remove_item", "Invalid item ID")
		return
	}

	// Get user's wishlist
	var wishlist models.Wishlist
	if err := h.db.Where("user_id = ?", uid).First(&wishlist).Error; err != nil {
		response.GenerateNotFoundResponse(c, "wishlist/remove_item", "Wishlist not found")
		return
	}

	// Find and delete the wishlist item
	var wishlistItem models.WishlistItem
	if err := h.db.Where("id = ? AND wishlist_id = ?", itemID, wishlist.ID).First(&wishlistItem).Error; err != nil {
		response.GenerateNotFoundResponse(c, "wishlist/remove_item", "Item not found in wishlist")
		return
	}

	if err := h.db.Delete(&wishlistItem).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "wishlist/remove_item", "Failed to remove item from wishlist")
		return
	}

	response.GenerateSuccessResponse(c, "Item removed from wishlist successfully", nil)
}
