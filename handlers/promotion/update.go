package promotion

import (
	"fmt"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *PromotionHandler) UpdatePromotion(c *gin.Context) {
	id := c.Param("id")

	var promotion models.Promotion
	if err := h.db.First(&promotion, id).Error; err != nil {
		response.GenerateBadRequestResponse(c, "promotion/update", "Promotion not found")
		return
	}

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
		if err == nil {
			promotion.StartDate = startDate
		}
	}
	if endDateStr != "" {
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err == nil {
			promotion.EndDate = endDate
		}
	}

	// Optional links
	if pid := c.PostForm("product_id"); pid != "" {
		var id uint
		_, err := fmt.Sscanf(pid, "%d", &id)
		if err == nil {
			promotion.ProductID = &id
		}
	}
	if cid := c.PostForm("category_id"); cid != "" {
		var id uint
		_, err := fmt.Sscanf(cid, "%d", &id)
		if err == nil {
			promotion.CategoryID = &id
		}
	}
	if bid := c.PostForm("brand_id"); bid != "" {
		var id uint
		_, err := fmt.Sscanf(bid, "%d", &id)
		if err == nil {
			promotion.BrandID = &id
		}
	}

	// Handle image replacement
	fileHeader, err := c.FormFile("image")
	if err == nil && fileHeader != nil {
		fileId, err := h.appwriteService.UploadFile(fileHeader)
		if err != nil {
			response.GenerateInternalServerErrorResponse(c, "promotion/update", "Failed to upload image to Appwrite: "+err.Error())
			return
		}
		promotion.Image = fileId
	}

	promotion.Title = title
	promotion.Description = description
	promotion.ButtonText = buttonText
	promotion.ButtonLink = buttonLink
	promotion.IsActive = isActive

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Save(&promotion).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "promotion/update", err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "promotion/commit", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Promotion updated successfully", promotion)
}
