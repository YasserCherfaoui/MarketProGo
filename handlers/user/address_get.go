package user

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *UserHandler) GetAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "user/get_address", "User not authenticated")
		return
	}
	uid := userID.(uint)

	addressID := c.Param("id")
	if addressID == "" {
		response.GenerateBadRequestResponse(c, "user/get_address", "Address ID is required")
		return
	}

	var address models.Address
	if err := h.db.Where("id = ? AND user_id = ?", addressID, uid).
		First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateNotFoundResponse(c, "user/get_address", "Address not found")
		} else {
			response.GenerateInternalServerErrorResponse(c, "user/get_address", "Failed to get address")
		}
		return
	}

	response.GenerateSuccessResponse(c, "Address retrieved successfully", address)
}
