package brand

import (
	"fmt"
	"strings"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *BrandHandler) UpdateBrand(c *gin.Context) {
	id := c.Param("id")

	var brand models.Brand
	if err := h.db.First(&brand, id).Error; err != nil {
		response.GenerateBadRequestResponse(c, "brand/update", "Brand not found")
		return
	}

	name := c.PostForm("name")
	isDisplayedStr := c.DefaultPostForm("is_displayed", "true")
	isDisplayed := isDisplayedStr == "true" || isDisplayedStr == "1"

	// Handle parent brand update
	if pid := c.PostForm("parent_id"); pid != "" {
		var id uint
		_, err := fmt.Sscanf(pid, "%d", &id)
		if err == nil {
			brand.ParentID = &id
		}
	}

	// Handle image replacement
	fileHeader, err := c.FormFile("image")
	if err == nil && fileHeader != nil {
		fileId, err := h.appwriteService.UploadFile(fileHeader)
		if err != nil {
			response.GenerateInternalServerErrorResponse(c, "brand/update", "Failed to upload image to Appwrite: "+err.Error())
			return
		}
		brand.Image = fileId
	}

	brand.Name = name
	brand.IsDisplayed = isDisplayed
	brand.Slug = strings.ToLower(strings.ReplaceAll(name, " ", "_"))

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Save(&brand).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "brand/update", err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "brand/commit", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Brand updated successfully", brand)
}
