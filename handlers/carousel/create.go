package carousel

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
		file, err := fileHeader.Open()
		if err != nil {
			response.GenerateBadRequestResponse(c, "carousel/create", "Failed to open uploaded image")
			return
		}
		defer file.Close()
		objectName := fmt.Sprintf("carousel/%s_%d%s", strings.ReplaceAll(title, " ", "_"), time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
		attrs, err := h.gcsService.UploadFile(c.Request.Context(), file, objectName, fileHeader.Header.Get("Content-Type"))
		if err != nil {
			response.GenerateInternalServerErrorResponse(c, "carousel/create", fmt.Sprintf("Failed to upload image to GCS: %v", err))
			return
		}
		imageURL = fmt.Sprintf("https://storage.googleapis.com/%s/%s", attrs.Bucket, attrs.Name)
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
