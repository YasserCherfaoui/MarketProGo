package cart

import (
	"strconv"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

func (h *CartHandler) UpdateItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "cart/update_item", "Unauthorized")
		return
	}
	uid := userID.(uint)

	itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.GenerateBadRequestResponse(c, "cart/update_item", "Invalid item ID")
		return
	}

	var req UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "cart/update_item", err.Error())
		return
	}

	var item models.CartItem
	if err := h.db.First(&item, itemID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "cart/update_item", "Cart item not found")
		return
	}

	// Optionally, check that the item belongs to the user's cart
	var cart models.Cart
	if err := h.db.First(&cart, item.CartID).Error; err != nil || cart.UserID == nil || uid == 0 || *cart.UserID != uid {
		response.GenerateForbiddenResponse(c, "cart/update_item", "Forbidden")
		return
	}

	// Update quantity and recalculate total price
	item.Quantity = req.Quantity
	item.TotalPrice = float64(item.Quantity) * item.UnitPrice

	h.db.Save(&item)

	// Preload variant and product data for response
	h.db.Preload("ProductVariant.Product").Preload("ProductVariant.Images").First(&item, item.ID)

	response.GenerateSuccessResponse(c, "cart/update_item", item)
}
