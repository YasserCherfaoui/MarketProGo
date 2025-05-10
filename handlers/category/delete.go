package category

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	categoryID := c.Param("id")

	if err := h.db.Delete(&models.Category{}, categoryID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "category/delete", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Category deleted successfully", nil)
}
