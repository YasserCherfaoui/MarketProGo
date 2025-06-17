package category

import (
	"strconv"
	"strings"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")

	// Find the category
	var category models.Category
	if err := h.db.First(&category, id).Error; err != nil {
		response.GenerateBadRequestResponse(c, "category/update", "Category not found")
		return
	}

	// Handle images_to_delete (single image URL or comma-separated)
	imagesToDelete := c.PostFormArray("images_to_delete")
	if len(imagesToDelete) == 1 && strings.Contains(imagesToDelete[0], ",") {
		imagesToDelete = strings.Split(imagesToDelete[0], ",")
	}
	imagesToDeleteMap := make(map[string]bool)
	for _, url := range imagesToDelete {
		imagesToDeleteMap[strings.TrimSpace(url)] = true
	}

	// If the current image is in images_to_delete, delete it from GCS and clear the field
	if category.Image != "" && imagesToDeleteMap[category.Image] {
		parts := strings.SplitN(category.Image, "/", 5)
		if len(parts) == 5 {
			objectName := parts[4]
			h.gcsService.DeleteFile(c.Request.Context(), objectName)
		}
		category.Image = ""
	}

	// Parse form fields
	name := c.PostForm("name")
	description := c.PostForm("description")
	parentIDStr := c.PostForm("parent_id")
	isFeatureOne := c.PostForm("is_feature_one")
	isFeatureOneBool, err := strconv.ParseBool(isFeatureOne)
	if err != nil {
		response.GenerateBadRequestResponse(c, "category/update", "Invalid is_feature_one")
		return
	}
	category.IsFeatureOne = isFeatureOneBool

	var parentID *uint
	if parentIDStr != "" {
		if pid64, err := strconv.ParseUint(parentIDStr, 10, 64); err == nil {
			pid := uint(pid64)
			parentID = &pid
		} else {
			response.GenerateBadRequestResponse(c, "category/update", "Invalid parent_id")
			return
		}
	}

	category.Name = name
	category.Description = description
	category.ParentID = parentID

	// Handle new image file (replace old image if provided)
	fileHeader, err := c.FormFile("image")
	if err == nil && fileHeader != nil {
		// Appwrite logic
		fileId, err := h.appwriteService.UploadFile(fileHeader)
		if err != nil {
			response.GenerateInternalServerErrorResponse(c, "category/update", "Failed to upload image to Appwrite: "+err.Error())
			return
		}
		category.Image = fileId
	}

	// Update slug if name changed
	category.Slug = strings.ToLower(strings.ReplaceAll(name, " ", "_"))

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Save(&category).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "category/update", err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "category/commit", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Category updated successfully", category)
}
