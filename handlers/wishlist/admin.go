package wishlist

import (
	"strconv"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

// GetAllWishlists retrieves all wishlists in the system (admin only)
func (h *WishlistHandler) GetAllWishlists(c *gin.Context) {
	var wishlists []models.Wishlist

	// Get query parameters for pagination and filtering
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	userID := c.Query("user_id")

	offset := (page - 1) * limit

	query := h.db.Preload("User").
		Preload("Items.ProductVariant.Product").
		Preload("Items.ProductVariant.Product.Brand").
		Preload("Items.ProductVariant.Product.Categories").
		Preload("Items.ProductVariant.Product.Images").
		Preload("Items.ProductVariant.InventoryItems")

	// Filter by user_id if provided
	if userID != "" {
		if userIDUint, err := strconv.ParseUint(userID, 10, 32); err == nil {
			query = query.Where("user_id = ?", userIDUint)
		}
	}

	// Get total count
	var total int64
	query.Model(&models.Wishlist{}).Count(&total)

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Find(&wishlists).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "wishlist/admin/get_all", "Failed to retrieve wishlists")
		return
	}

	response.GenerateSuccessResponse(c, "Wishlists retrieved successfully", gin.H{
		"wishlists": wishlists,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetWishlistByID retrieves a specific wishlist by ID (admin only)
func (h *WishlistHandler) GetWishlistByID(c *gin.Context) {
	wishlistIDStr := c.Param("id")
	wishlistID, err := strconv.ParseUint(wishlistIDStr, 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "wishlist/admin/get_by_id", "Invalid wishlist ID")
		return
	}

	var wishlist models.Wishlist
	if err := h.db.Preload("User").
		Preload("Items.ProductVariant.Product").
		Preload("Items.ProductVariant.Product.Brand").
		Preload("Items.ProductVariant.Product.Categories").
		Preload("Items.ProductVariant.Product.Images").
		Preload("Items.ProductVariant.InventoryItems").
		First(&wishlist, wishlistID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "wishlist/admin/get_by_id", "Wishlist not found")
		return
	}

	response.GenerateSuccessResponse(c, "Wishlist retrieved successfully", wishlist)
}

// GetWishlistStats retrieves wishlist statistics (admin only)
func (h *WishlistHandler) GetWishlistStats(c *gin.Context) {
	var stats struct {
		TotalWishlists     int64 `json:"total_wishlists"`
		TotalWishlistItems int64 `json:"total_wishlist_items"`
		ActiveUsers        int64 `json:"active_users"`
		MostWishedItems    []struct {
			ProductVariantID uint   `json:"product_variant_id"`
			ProductName      string `json:"product_name"`
			VariantName      string `json:"variant_name"`
			Count            int64  `json:"count"`
		} `json:"most_wished_items"`
	}

	// Get total wishlists
	h.db.Model(&models.Wishlist{}).Count(&stats.TotalWishlists)

	// Get total wishlist items
	h.db.Model(&models.WishlistItem{}).Count(&stats.TotalWishlistItems)

	// Get active users (users with wishlists)
	h.db.Model(&models.Wishlist{}).Distinct("user_id").Count(&stats.ActiveUsers)

	// Get most wished items
	rows, err := h.db.Model(&models.WishlistItem{}).
		Select("product_variant_id, COUNT(*) as count").
		Group("product_variant_id").
		Order("count DESC").
		Limit(10).
		Rows()
	if err != nil {
		response.GenerateInternalServerErrorResponse(c, "wishlist/admin/stats", "Failed to retrieve wishlist statistics")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item struct {
			ProductVariantID uint   `json:"product_variant_id"`
			ProductName      string `json:"product_name"`
			VariantName      string `json:"variant_name"`
			Count            int64  `json:"count"`
		}

		var productVariantID uint
		var count int64
		if err := rows.Scan(&productVariantID, &count); err != nil {
			continue
		}

		// Get product variant details
		var productVariant models.ProductVariant
		if err := h.db.Preload("Product").First(&productVariant, productVariantID).Error; err != nil {
			continue
		}

		item.ProductVariantID = productVariantID
		item.ProductName = productVariant.Product.Name
		item.VariantName = productVariant.Name
		item.Count = count

		stats.MostWishedItems = append(stats.MostWishedItems, item)
	}

	response.GenerateSuccessResponse(c, "Wishlist statistics retrieved successfully", stats)
}

// DeleteWishlistItem deletes a wishlist item (admin only)
func (h *WishlistHandler) DeleteWishlistItem(c *gin.Context) {
	itemIDStr := c.Param("id")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "wishlist/admin/delete_item", "Invalid item ID")
		return
	}

	var wishlistItem models.WishlistItem
	if err := h.db.First(&wishlistItem, itemID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "wishlist/admin/delete_item", "Wishlist item not found")
		return
	}

	if err := h.db.Delete(&wishlistItem).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "wishlist/admin/delete_item", "Failed to delete wishlist item")
		return
	}

	response.GenerateSuccessResponse(c, "Wishlist item deleted successfully", nil)
}

// GetUserWishlist retrieves a specific user's wishlist (admin only)
func (h *WishlistHandler) GetUserWishlist(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "wishlist/admin/get_user_wishlist", "Invalid user ID")
		return
	}

	var wishlist models.Wishlist
	if err := h.db.Where("user_id = ?", userID).
		Preload("User").
		Preload("Items.ProductVariant.Product").
		Preload("Items.ProductVariant.Product.Brand").
		Preload("Items.ProductVariant.Product.Categories").
		Preload("Items.ProductVariant.Product.Images").
		Preload("Items.ProductVariant.InventoryItems").
		First(&wishlist).Error; err != nil {
		response.GenerateNotFoundResponse(c, "wishlist/admin/get_user_wishlist", "User wishlist not found")
		return
	}

	response.GenerateSuccessResponse(c, "User wishlist retrieved successfully", wishlist)
}
