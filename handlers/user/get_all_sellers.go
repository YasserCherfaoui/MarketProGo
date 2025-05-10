package user

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *UserHandler) GetAllSellers(c *gin.Context) {
	var companies []models.Company
	if err := h.db.
		Preload("Users").
		Find(&companies).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "user/get_all", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Sellers fetched successfully", companies)

}
