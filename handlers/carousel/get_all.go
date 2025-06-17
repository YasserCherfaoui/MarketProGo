package carousel

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *CarouselHandler) GetCarousel(c *gin.Context) {
	var carousel []models.Carousel
	if err := h.db.Find(&carousel).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "carousel/get_all", err.Error())
		return
	}
	// Add Appwrite URLs to carousel images
	for i := range carousel {
		carousel[i].ImageURL = h.appwriteService.GetFileURL(carousel[i].ImageURL)
	}
	response.GenerateSuccessResponse(c, "carousel/get_all", carousel)
}
