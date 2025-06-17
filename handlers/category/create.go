package category

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	// Parse form fields
	name := c.PostForm("name")
	description := c.PostForm("description")
	parentIDStr := c.PostForm("parent_id")
	var parentID *uint
	if parentIDStr != "" {
		if pid64, err := strconv.ParseUint(parentIDStr, 10, 64); err == nil {
			pid := uint(pid64)
			parentID = &pid
		} else {
			response.GenerateBadRequestResponse(c, "category/create", "Invalid parent_id")
			return
		}
	}

	// Handle image file
	imageURL := ""
	fileHeader, err := c.FormFile("image")
	if err == nil && fileHeader != nil {
		// GCS logic commented out
		/*
			file, err := fileHeader.Open()
			if err != nil {
				response.GenerateBadRequestResponse(c, "category/create", "Failed to open uploaded image")
				return
			}
			defer file.Close()
			objectName := fmt.Sprintf("categories/%s_%d%s", strings.ReplaceAll(name, " ", "_"), time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
			attrs, err := h.gcsService.UploadFile(c.Request.Context(), file, objectName, fileHeader.Header.Get("Content-Type"))
			if err != nil {
				response.GenerateInternalServerErrorResponse(c, "category/create", fmt.Sprintf("Failed to upload image to GCS: %v", err))
				return
			}
			imageURL = fmt.Sprintf("https://storage.googleapis.com/%s/%s", attrs.Bucket, attrs.Name)
		*/

		// Appwrite logic
		fileId, err := h.appwriteService.UploadFile(fileHeader)
		if err != nil {
			response.GenerateInternalServerErrorResponse(c, "category/create", "Failed to upload image to Appwrite: "+err.Error())
			return
		}
		imageURL = h.appwriteService.GetFileURL(fileId)
	}

	category := models.Category{
		Name:        name,
		Description: description,
		Image:       imageURL,
		ParentID:    parentID,
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// Generate slug from name
	category.Slug = strings.ToLower(strings.ReplaceAll(name, " ", "_"))

	// Check if slug is unique
	if err := tx.Where("slug = ?", category.Slug).First(&models.Category{}).Error; err != nil {

		// If slug already exists, make it unique by appending a counter
		var count int64
		tx.Model(&models.Category{}).Where("slug LIKE ?", category.Slug+"%").Count(&count)
		if count > 0 {
			category.Slug = fmt.Sprintf("%s-%d", category.Slug, count+1)
		}
	}

	if err := tx.Create(&category).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "category/create", err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "category/commit", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Category created successfully", category)
}
