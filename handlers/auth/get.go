package auth

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) GetUser(c *gin.Context) {
	userID := c.GetUint("user_id")
	var user models.User

	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		response.GenerateNotFoundResponse(c, "auth/get-user", "User not found")
		return
	}

	response.GenerateSuccessResponse(c, "User retrieved successfully", user)
}
