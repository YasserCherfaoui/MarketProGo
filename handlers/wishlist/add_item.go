package wishlist

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type AddWishlistItemRequest struct {
	ProductVariantID uint   `json:"product_variant_id" binding:"required"`
	ProductID        *uint  `json:"product_id,omitempty"` // Legacy support
	Notes            string `json:"notes"`
	Priority         int    `json:"priority"` // 1-5, 5 being highest
	IsPublic         bool   `json:"is_public"`
}

// AddItem adds an item to the user's wishlist
func (h *WishlistHandler) AddItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "wishlist/add_item", "User not authenticated")
		return
	}
	uid := userID.(uint)

	var req AddWishlistItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "wishlist/add_item", "Invalid request data: "+err.Error())
		return
	}

	// Validate priority range
	if req.Priority < 1 || req.Priority > 5 {
		req.Priority = 3 // Default priority
	}

	// Check if product variant exists
	var productVariant models.ProductVariant
	if err := h.db.First(&productVariant, req.ProductVariantID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "wishlist/add_item", "Product variant not found")
		return
	}

	// Get or create user's wishlist
	var wishlist models.Wishlist
	err := h.db.Where("user_id = ?", uid).First(&wishlist).Error
	if err != nil {
		// Create new wishlist if it doesn't exist
		wishlist = models.Wishlist{
			UserID: &uid,
		}
		if err := h.db.Create(&wishlist).Error; err != nil {
			response.GenerateInternalServerErrorResponse(c, "wishlist/add_item", "Failed to create wishlist")
			return
		}
	}

	// Check if item already exists in wishlist
	var existingItem models.WishlistItem
	err = h.db.Where("wishlist_id = ? AND product_variant_id = ?", wishlist.ID, req.ProductVariantID).First(&existingItem).Error
	if err == nil {
		response.GenerateBadRequestResponse(c, "wishlist/add_item", "Item already exists in wishlist")
		return
	}

	// Create wishlist item
	wishlistItem := models.WishlistItem{
		WishlistID:       wishlist.ID,
		ProductVariantID: req.ProductVariantID,
		ProductID:        req.ProductID,
		Notes:            req.Notes,
		Priority:         req.Priority,
		IsPublic:         req.IsPublic,
	}

	if err := h.db.Create(&wishlistItem).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "wishlist/add_item", "Failed to add item to wishlist")
		return
	}

	// Load the product variant data for response
	h.db.Preload("ProductVariant.Product").First(&wishlistItem, wishlistItem.ID)

	response.GenerateCreatedResponse(c, "Item added to wishlist successfully", wishlistItem)
}
