package user

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *UserHandler) GetAddresses(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "user/get_addresses", "User not authenticated")
		return
	}
	uid := userID.(uint)

	var addresses []models.Address
	if err := h.db.Where("user_id = ?", uid).
		Order("is_default DESC, created_at DESC").
		Find(&addresses).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "user/get_addresses", "Failed to get addresses")
		return
	}

	response.GenerateSuccessResponse(c, "Addresses retrieved successfully", addresses)
}
