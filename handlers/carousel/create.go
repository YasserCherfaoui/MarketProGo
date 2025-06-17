package carousel

import (
	"strconv"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *CarouselHandler) CreateCarousel(c *gin.Context) {

	// Parse form fields
	title := c.PostForm("title")
	rankStr := c.PostForm("rank")
	link := c.PostForm("link")
	caption := c.PostForm("caption")

	// Validate and parse rank
	var rank int
	if rankStr == "" {
		response.GenerateBadRequestResponse(c, "carousel/create", "Rank is required")
		return
	}
	if r, err := strconv.Atoi(rankStr); err == nil {
		rank = r
	} else {
		response.GenerateBadRequestResponse(c, "carousel/create", "Invalid rank")
		return
	}

	// Check for unique rank
	var count int64
	h.db.Model(&models.Carousel{}).Where("rank = ?", rank).Count(&count)
	if count > 0 {
		response.GenerateBadRequestResponse(c, "carousel/create", "Rank must be unique")
		return
	}

	// Handle image file
	imageURL := ""
	fileHeader, err := c.FormFile("image")
	if err == nil && fileHeader != nil {
		fileId, err := h.appwriteService.UploadFile(fileHeader)
		if err != nil {
			response.GenerateInternalServerErrorResponse(c, "carousel/create", "Failed to upload image to Appwrite: "+err.Error())
			return
		}
		imageURL = fileId
	} else {
		response.GenerateBadRequestResponse(c, "carousel/create", "Image is required")
		return
	}

	carousel := models.Carousel{
		Title:    title,
		ImageURL: imageURL,
		Rank:     rank,
		Link:     link,
		Caption:  caption,
	}

	if err := h.db.Create(&carousel).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "carousel/create", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Carousel created successfully", carousel)
}
