package cart

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

// GetCart returns the current user's cart, or creates one if it doesn't exist
func (h *CartHandler) GetCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "cart/get", "Unauthorized")
		return
	}

	uid := userID.(uint)
	var cart models.Cart

	if err := h.db.Preload("Items.Product.Images").Where("user_id = ?", uid).First(&cart).Error; err != nil {
		// If not found, create a new cart
		cart = models.Cart{UserID: &uid}
		if err := h.db.Create(&cart).Error; err != nil {
			response.GenerateInternalServerErrorResponse(c, "cart/get", err.Error())
			return
		}
	}

	response.GenerateSuccessResponse(c, "cart/get", cart)
}
