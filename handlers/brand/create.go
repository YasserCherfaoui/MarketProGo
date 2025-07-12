package brand

import (
	"fmt"
	"strings"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *BrandHandler) CreateBrand(c *gin.Context) {
	name := c.PostForm("name")
	isDisplayedStr := c.DefaultPostForm("is_displayed", "true")
	isDisplayed := isDisplayedStr == "true" || isDisplayedStr == "1"

	// Handle parent brand
	var parentID *uint
	if pid := c.PostForm("parent_id"); pid != "" {
		var id uint
		_, err := fmt.Sscanf(pid, "%d", &id)
		if err == nil {
			parentID = &id
		}
	}

	// Handle image file
	imageURL := ""
	fileHeader, err := c.FormFile("image")
	if err == nil && fileHeader != nil {
		fileId, err := h.appwriteService.UploadFile(fileHeader)
		if err != nil {
			response.GenerateInternalServerErrorResponse(c, "brand/create", "Failed to upload image to Appwrite: "+err.Error())
			return
		}
		imageURL = fileId
	}

	slug := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Ensure slug is unique
	if err := tx.Where("slug = ?", slug).First(&models.Brand{}).Error; err == nil {
		var count int64
		tx.Model(&models.Brand{}).Where("slug LIKE ?", slug+"%").Count(&count)
		if count > 0 {
			slug = fmt.Sprintf("%s-%d", slug, count+1)
		}
	}

	brand := models.Brand{
		Name:        name,
		Image:       imageURL,
		Slug:        slug,
		IsDisplayed: isDisplayed,
		ParentID:    parentID,
	}

	if err := tx.Create(&brand).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "brand/create", err.Error())
		tx.Rollback()
		return
	}

	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "brand/commit", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Brand created successfully", brand)
}
