package category

import (
	"fmt"
	"strings"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Image       string `json:"image"`
	ParentID    *uint  `json:"parent_id"`
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "category/create", err.Error())
		return
	}

	category := models.Category{
		Name:        req.Name,
		Description: req.Description,
		Image:       req.Image,
		ParentID:    req.ParentID,
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// Generate slug from name
	category.Slug = strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))

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
