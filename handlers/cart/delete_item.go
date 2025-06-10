package cart

import (
	"strconv"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *CartHandler) DeleteItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "cart/delete_item", "Unauthorized")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		response.GenerateUnauthorizedResponse(c, "cart/delete_item", "Unauthorized")
		return
	}

	itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.GenerateBadRequestResponse(c, "cart/delete_item", "Invalid item ID")
		return
	}

	var item models.CartItem
	if err := h.db.First(&item, itemID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "cart/delete_item", "Cart item not found")
		return
	}

	var cart models.Cart
	if err := h.db.First(&cart, item.CartID).Error; err != nil || cart.UserID == nil || *cart.UserID != userIDUint {
		response.GenerateForbiddenResponse(c, "cart/delete_item", "Forbidden")
		return
	}

	h.db.Delete(&item)
	response.GenerateSuccessResponse(c, "cart/delete_item", "Item removed from cart")
}
