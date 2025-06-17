package carousel

import (
	"strconv"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *CarouselHandler) UpdateCarousel(c *gin.Context) {
	id := c.Param("id")

	// Find the carousel row
	var carousel models.Carousel
	if err := h.db.First(&carousel, id).Error; err != nil {
		response.GenerateBadRequestResponse(c, "carousel/update", "Carousel not found")
		return
	}

	// Parse form fields
	title := c.PostForm("title")
	rankStr := c.PostForm("rank")
	link := c.PostForm("link")
	caption := c.PostForm("caption")

	// Validate and parse rank
	var rank int
	if rankStr == "" {
		rank = carousel.Rank // keep old rank if not provided
	} else {
		if r, err := strconv.Atoi(rankStr); err == nil {
			rank = r
		} else {
			response.GenerateBadRequestResponse(c, "carousel/update", "Invalid rank")
			return
		}
	}

	// Check for unique rank (if changed)
	if rank != carousel.Rank {
		var count int64
		h.db.Model(&models.Carousel{}).Where("rank = ? AND id != ?", rank, id).Count(&count)
		if count > 0 {
			response.GenerateBadRequestResponse(c, "carousel/update", "Rank must be unique")
			return
		}
	}

	// Handle image file (optional replacement)
	fileHeader, err := c.FormFile("image")
	if err == nil && fileHeader != nil {
		fileId, err := h.appwriteService.UploadFile(fileHeader)
		if err != nil {
			response.GenerateInternalServerErrorResponse(c, "carousel/update", "Failed to upload image to Appwrite: "+err.Error())
			return
		}
		carousel.ImageURL = fileId
	}

	// Update fields
	if title != "" {
		carousel.Title = title
	}
	carousel.Rank = rank
	carousel.Link = link
	carousel.Caption = caption

	if err := h.db.Save(&carousel).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "carousel/update", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Carousel updated successfully", carousel)
}
