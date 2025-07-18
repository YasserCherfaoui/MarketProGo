package review

import (
	"net/http"
	"strconv"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ModerationRequest represents the request body for moderating a review
type ModerationRequest struct {
	Status models.ReviewStatus `json:"status" binding:"required"`
	Reason string              `json:"reason" binding:"required,max=500"`
}

// GetAllReviews handles GET /api/v1/admin/reviews
func (h *ReviewHandler) GetAllReviews(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	rating, _ := strconv.Atoi(c.Query("rating"))
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
	productVariantID, _ := strconv.ParseUint(c.Query("product_variant_id"), 10, 32)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query
	query := h.db.Model(&models.ProductReview{}).
		Preload("User").
		Preload("ProductVariant.Product").
		Preload("Images").
		Preload("SellerResponse")

	// Add filters
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if rating > 0 && rating <= 5 {
		query = query.Where("rating = ?", rating)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if productVariantID > 0 {
		query = query.Where("product_variant_id = ?", productVariantID)
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

// ModerateReview handles PUT /api/v1/admin/reviews/:id/moderate
func (h *ReviewHandler) ModerateReview(c *gin.Context) {
	// Get admin user ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		response.GenerateErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Admin not authenticated")
		return
	}

	// Parse review ID
	reviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REVIEW_ID", "Invalid review ID")
		return
	}

	// Parse request body
	var req ModerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate status
	if req.Status != models.ReviewStatusApproved &&
		req.Status != models.ReviewStatusRejected &&
		req.Status != models.ReviewStatusFlagged {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_STATUS", "Invalid review status")
		return
	}

	// Find the review
	var review models.ProductReview
	err = h.db.Preload("ProductVariant.Product").Where("id = ?", reviewID).First(&review).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateErrorResponse(c, http.StatusNotFound, "REVIEW_NOT_FOUND", "Review not found")
			return
		}
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to retrieve review")
		return
	}

	// Store old status for logging
	oldStatus := review.Status

	// Update review status
	review.Status = req.Status
	review.ModerationReason = req.Reason
	adminIDUint := adminID.(uint)
	review.ModeratedBy = &adminIDUint
	now := time.Now()
	review.ModeratedAt = &now

	// Save the updated review
	if err := h.db.Save(&review).Error; err != nil {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update review")
		return
	}

	// Create moderation log entry
	moderationLog := models.ReviewModerationLog{
		ReviewID:    review.ID,
		AdminID:     adminID.(uint),
		OldStatus:   oldStatus,
		NewStatus:   req.Status,
		Reason:      req.Reason,
		ModeratedAt: time.Now(),
	}
	h.db.Create(&moderationLog)

	// Update rating aggregation if status changed to/from approved
	if (oldStatus == models.ReviewStatusApproved && req.Status != models.ReviewStatusApproved) ||
		(oldStatus != models.ReviewStatusApproved && req.Status == models.ReviewStatusApproved) {
		if err := h.UpdateProductRating(review.ProductVariantID); err != nil {
			// Log the error but don't fail the request
			// TODO: Add proper logging
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Review moderated successfully",
		"data": gin.H{
			"review_id":    review.ID,
			"old_status":   oldStatus,
			"new_status":   req.Status,
			"moderated_by": adminID,
			"moderated_at": review.ModeratedAt,
		},
	})
}

// AdminDeleteReview handles DELETE /api/v1/admin/reviews/:id
func (h *ReviewHandler) AdminDeleteReview(c *gin.Context) {
	// Get admin user ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		response.GenerateErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Admin not authenticated")
		return
	}

	// Parse review ID
	reviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REVIEW_ID", "Invalid review ID")
		return
	}

	// Find the review
	var review models.ProductReview
	err = h.db.Preload("ProductVariant.Product").Where("id = ?", reviewID).First(&review).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateErrorResponse(c, http.StatusNotFound, "REVIEW_NOT_FOUND", "Review not found")
			return
		}
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to retrieve review")
		return
	}

	// Store product variant ID for aggregation update
	productVariantID := review.ProductVariantID

	// Permanently delete the review and related data
	if err := h.db.Unscoped().Delete(&review).Error; err != nil {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete review")
		return
	}

	// Permanently delete related data
	h.db.Unscoped().Where("product_review_id = ?", review.ID).Delete(&models.ReviewImage{})
	h.db.Unscoped().Where("product_review_id = ?", review.ID).Delete(&models.ReviewHelpful{})
	h.db.Unscoped().Where("product_review_id = ?", review.ID).Delete(&models.SellerResponse{})
	h.db.Unscoped().Where("review_id = ?", review.ID).Delete(&models.ReviewModerationLog{})

	// Create deletion log entry
	deletionLog := models.ReviewModerationLog{
		ReviewID:    review.ID,
		AdminID:     adminID.(uint),
		OldStatus:   review.Status,
		NewStatus:   "DELETED",
		Reason:      "Permanently deleted by admin",
		ModeratedAt: time.Now(),
	}
	h.db.Create(&deletionLog)

	// Update rating aggregation
	if err := h.UpdateProductRating(productVariantID); err != nil {
		// Log the error but don't fail the request
		// TODO: Add proper logging
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Review permanently deleted",
		"data": gin.H{
			"review_id":  review.ID,
			"deleted_by": adminID,
			"deleted_at": time.Now(),
		},
	})
}

// GetModerationStats handles GET /api/v1/admin/reviews/stats
func (h *ReviewHandler) GetModerationStats(c *gin.Context) {
	// Get review counts by status
	var stats struct {
		Total    int64 `json:"total"`
		Pending  int64 `json:"pending"`
		Approved int64 `json:"approved"`
		Rejected int64 `json:"rejected"`
		Flagged  int64 `json:"flagged"`
		Deleted  int64 `json:"deleted"`
	}

	h.db.Model(&models.ProductReview{}).Count(&stats.Total)
	h.db.Model(&models.ProductReview{}).Where("status = ?", models.ReviewStatusPending).Count(&stats.Pending)
	h.db.Model(&models.ProductReview{}).Where("status = ?", models.ReviewStatusApproved).Count(&stats.Approved)
	h.db.Model(&models.ProductReview{}).Where("status = ?", models.ReviewStatusRejected).Count(&stats.Rejected)
	h.db.Model(&models.ProductReview{}).Where("status = ?", models.ReviewStatusFlagged).Count(&stats.Flagged)
	h.db.Unscoped().Model(&models.ProductReview{}).Where("deleted_at IS NOT NULL").Count(&stats.Deleted)

	// Get recent moderation activity (last 7 days)
	var recentModerations []models.ReviewModerationLog
	h.db.Preload("Review").
		Preload("Admin").
		Where("moderated_at >= ?", time.Now().AddDate(0, 0, -7)).
		Order("moderated_at DESC").
		Limit(10).
		Find(&recentModerations)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"stats":              stats,
			"recent_moderations": recentModerations,
		},
	})
}
