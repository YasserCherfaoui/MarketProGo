package category

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *CategoryHandler) GetAllCategories(c *gin.Context) {
	var categories []models.Category
	if err := h.db.Preload("Parent").Preload("Children").Preload("Products").Find(&categories).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "category/get_all", "Failed to get all categories")
		return
	}
	response.GenerateSuccessResponse(c, "Categories fetched successfully", categories)
}
