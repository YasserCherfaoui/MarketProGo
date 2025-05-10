package user

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	var users []models.User
	if err := h.db.Find(&users).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "user/get_all", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Users fetched successfully", users)
}
