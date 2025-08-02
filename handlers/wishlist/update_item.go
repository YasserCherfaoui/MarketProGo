package wishlist

import (
	"strconv"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type UpdateWishlistItemRequest struct {
	Notes    string `json:"notes"`
	Priority int    `json:"priority"` // 1-5, 5 being highest
	IsPublic bool   `json:"is_public"`
}

// UpdateItem updates a wishlist item's metadata
func (h *WishlistHandler) UpdateItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "wishlist/update_item", "User not authenticated")
		return
	}
	uid := userID.(uint)

	// Get item ID from URL parameter
	itemIDStr := c.Param("id")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "wishlist/update_item", "Invalid item ID")
		return
	}

	var req UpdateWishlistItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "wishlist/update_item", "Invalid request data: "+err.Error())
		return
	}

	// Validate priority range
	if req.Priority < 1 || req.Priority > 5 {
		req.Priority = 3 // Default priority
	}

	// Get user's wishlist
	var wishlist models.Wishlist
	if err := h.db.Where("user_id = ?", uid).First(&wishlist).Error; err != nil {
		response.GenerateNotFoundResponse(c, "wishlist/update_item", "Wishlist not found")
		return
	}

	// Find the wishlist item
	var wishlistItem models.WishlistItem
	if err := h.db.Where("id = ? AND wishlist_id = ?", itemID, wishlist.ID).First(&wishlistItem).Error; err != nil {
		response.GenerateNotFoundResponse(c, "wishlist/update_item", "Item not found in wishlist")
		return
	}

	// Update the item
	wishlistItem.Notes = req.Notes
	wishlistItem.Priority = req.Priority
	wishlistItem.IsPublic = req.IsPublic

	if err := h.db.Save(&wishlistItem).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "wishlist/update_item", "Failed to update wishlist item")
		return
	}

	// Load the product variant data for response
	h.db.Preload("ProductVariant.Product").First(&wishlistItem, wishlistItem.ID)

	response.GenerateSuccessResponse(c, "Wishlist item updated successfully", wishlistItem)
}
