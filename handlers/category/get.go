package category

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *CategoryHandler) GetCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category
	if err := h.db.Where("id = ?", id).First(&category).Error; err != nil {
		response.GenerateNotFoundResponse(c, "category/get", "Category not found")
		return
	}
	response.GenerateSuccessResponse(c, "Category fetched successfully", category)
}
