package user

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if err := h.db.Delete(&models.User{}, userID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "user/delete", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "User deleted successfully", nil)
}
