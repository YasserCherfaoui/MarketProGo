package promotion

import (
	"fmt"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *PromotionHandler) CreatePromotion(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	buttonText := c.PostForm("button_text")
	buttonLink := c.PostForm("button_link")
	isActiveStr := c.DefaultPostForm("is_active", "true")
	isActive := isActiveStr == "true" || isActiveStr == "1"

	// Parse scheduling
	startDateStr := c.PostForm("start_date")
	endDateStr := c.PostForm("end_date")
	var startDate, endDate time.Time
	var err error
	if startDateStr != "" {
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			response.GenerateBadRequestResponse(c, "promotion/create", "Invalid start_date format. Use RFC3339.")
			return
		}
	}
	if endDateStr != "" {
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			response.GenerateBadRequestResponse(c, "promotion/create", "Invalid end_date format. Use RFC3339.")
			return
		}
	}

	// Optional links
	var productID, categoryID, brandID *uint
	if pid := c.PostForm("product_id"); pid != "" {
		var id uint
		_, err := fmt.Sscanf(pid, "%d", &id)
		if err == nil {
			productID = &id
		}
	}
	if cid := c.PostForm("category_id"); cid != "" {
		var id uint
		_, err := fmt.Sscanf(cid, "%d", &id)
		if err == nil {
			categoryID = &id
		}
	}
	if bid := c.PostForm("brand_id"); bid != "" {
		var id uint
		_, err := fmt.Sscanf(bid, "%d", &id)
		if err == nil {
			brandID = &id
		}
	}

	// Handle image file
	imageURL := ""
	fileHeader, err := c.FormFile("image")
	if err == nil && fileHeader != nil {
		fileId, err := h.appwriteService.UploadFile(fileHeader)
		if err != nil {
			response.GenerateInternalServerErrorResponse(c, "promotion/create", "Failed to upload image to Appwrite: "+err.Error())
			return
		}
		imageURL = fileId
	} else {
		response.GenerateBadRequestResponse(c, "promotion/create", "Image is required.")
		return
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	promotion := models.Promotion{
		Title:       title,
		Description: description,
		Image:       imageURL,
		ButtonText:  buttonText,
		ButtonLink:  buttonLink,
		StartDate:   startDate,
		EndDate:     endDate,
		IsActive:    isActive,
		ProductID:   productID,
		CategoryID:  categoryID,
		BrandID:     brandID,
	}

	if err := tx.Create(&promotion).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "promotion/create", err.Error())
		tx.Rollback()
		return
	}

	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "promotion/commit", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Promotion created successfully", promotion)
}
