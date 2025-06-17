package brand

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *BrandHandler) DeleteBrand(c *gin.Context) {
	brandID := c.Param("id")

	if err := h.db.Delete(&models.Brand{}, brandID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "brand/delete", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Brand deleted successfully", nil)
}
