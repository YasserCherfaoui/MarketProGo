package review

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UpdateReviewRequest represents the request body for updating a review
type UpdateReviewRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Title   string `json:"title" binding:"required,min=1,max=100"`
	Content string `json:"content" binding:"required,min=10,max=1000"`
}

// UpdateReview handles PUT /api/v1/reviews/:id
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse review ID
	reviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REVIEW_ID", "Invalid review ID")
		return
	}

	// Parse request body
	var req UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Find the review and verify ownership
	var review models.ProductReview
	err = h.db.Preload("ProductVariant.Product").Where("id = ? AND user_id = ?", reviewID, userID.(uint)).First(&review).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateErrorResponse(c, http.StatusNotFound, "REVIEW_NOT_FOUND", "Review not found or you don't own this review")
			return
		}
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to retrieve review")
		return
	}

	// Check if review is approved (only allow updates to approved reviews)
	if review.Status != models.ReviewStatusApproved {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "REVIEW_NOT_APPROVED", "Can only update approved reviews")
		return
	}

	// Update review fields
	review.Rating = req.Rating
	review.Title = strings.TrimSpace(req.Title)
	review.Content = strings.TrimSpace(req.Content)

	// Save the updated review
	if err := h.db.Save(&review).Error; err != nil {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update review")
		return
	}

	// Update rating aggregation
	if err := h.UpdateProductRating(review.ProductVariant.ProductID); err != nil {
		// Log the error but don't fail the request
		// TODO: Add proper logging
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Review updated successfully",
		"data":    review,
	})
}

// DeleteReview handles DELETE /api/v1/reviews/:id
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse review ID
	reviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REVIEW_ID", "Invalid review ID")
		return
	}

	// Find the review and verify ownership
	var review models.ProductReview
	err = h.db.Preload("ProductVariant.Product").Where("id = ? AND user_id = ?", reviewID, userID.(uint)).First(&review).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateErrorResponse(c, http.StatusNotFound, "REVIEW_NOT_FOUND", "Review not found or you don't own this review")
			return
		}
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to retrieve review")
		return
	}

	// Check if review is approved (only allow deletion of approved reviews)
	if review.Status != models.ReviewStatusApproved {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "REVIEW_NOT_APPROVED", "Can only delete approved reviews")
		return
	}

	// Store product ID for aggregation update
	productID := review.ProductVariant.ProductID

	// Delete the review (soft delete)
	if err := h.db.Delete(&review).Error; err != nil {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete review")
		return
	}

	// Also delete related data (soft delete)
	h.db.Where("product_review_id = ?", review.ID).Delete(&models.ReviewImage{})
	h.db.Where("product_review_id = ?", review.ID).Delete(&models.ReviewHelpful{})
	h.db.Where("product_review_id = ?", review.ID).Delete(&models.SellerResponse{})

	// Update rating aggregation
	if err := h.UpdateProductRating(productID); err != nil {
		// Log the error but don't fail the request
		// TODO: Add proper logging
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Review deleted successfully",
	})
}

// GetUserReviews handles GET /api/v1/reviews/user/me
func (h *ReviewHandler) GetUserReviews(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build query
	query := h.db.Model(&models.ProductReview{}).
		Preload("ProductVariant.Product").
		Preload("Images").
		Preload("SellerResponse").
		Where("user_id = ?", userID.(uint))

	// Add status filter if provided
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get reviews with pagination
	var reviews []models.ProductReview
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&reviews).Error

	if err != nil {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to retrieve reviews")
		return
	}

	// Calculate pagination info
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"reviews": reviews,
			"pagination": gin.H{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"totalPages": totalPages,
				"hasNext":    hasNext,
				"hasPrev":    hasPrev,
			},
		},
	})
}
