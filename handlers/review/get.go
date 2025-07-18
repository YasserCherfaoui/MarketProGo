package review

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetReview handles GET /api/v1/reviews/:id
// Returns a single review by ID with all related data
// Admins can see all reviews regardless of status, others only see approved reviews
func (h *ReviewHandler) GetReview(c *gin.Context) {
	reviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REVIEW_ID", "Invalid review ID")
		return
	}

	// Check if user is authenticated and is admin
	userType, exists := c.Get("user_type")
	isAdmin := exists && userType == models.Admin

	// Build query based on user permissions
	query := h.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, first_name, last_name, email, phone, avatar")
	}).
		Preload("ProductVariant").
		Preload("Images").
		Preload("SellerResponse.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, first_name, last_name, email, phone, avatar")
		}).
		Where("id = ?", reviewID)

	// Only show approved reviews unless user is admin
	if !isAdmin {
		query = query.Where("status = ?", models.ReviewStatusApproved)
	}

	var review models.ProductReview
	err = query.First(&review).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateErrorResponse(c, http.StatusNotFound, "REVIEW_NOT_FOUND", "Review not found")
			return
		}
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "RETRIEVE_REVIEW_ERROR", "Failed to retrieve review")
		return
	}

	// Format response
	responseData := gin.H{
		"ID":                 review.ID,
		"product_variant_id": review.ProductVariantID,
		"product_variant":    review.ProductVariant,
		"user": gin.H{
			"ID":         review.User.ID,
			"first_name": review.User.FirstName,
			"last_name":  review.User.LastName,
			"name":       review.GetReviewerName(),
			"email":      review.User.Email,
			"phone":      review.User.Phone,
			"avatar":     review.User.Avatar,
		},
		"rating":               review.Rating,
		"title":                review.Title,
		"content":              review.Content,
		"is_verified_purchase": review.IsVerifiedPurchase,
		"helpful_count":        review.HelpfulCount,
		"images":               review.Images,
		"CreatedAt":            review.CreatedAt,
		"UpdatedAt":            review.UpdatedAt,
		"status":               review.Status,
		"moderated_at":         review.ModeratedAt,
		"moderation_reason":    review.ModerationReason,
	}

	// Fetch moderation history if user is admin
	if isAdmin {
		var moderationLogs []models.ReviewModerationLog
		err = h.db.Where("review_id = ?", review.ID).
			Preload("Admin", func(db *gorm.DB) *gorm.DB {
				return db.Select("id, first_name, last_name, email")
			}).
			Order("created_at DESC").
			Find(&moderationLogs).Error

		if err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to fetch moderation logs: %v\n", err)
		} else {
			var moderationHistory []gin.H
			for _, log := range moderationLogs {
				moderationHistory = append(moderationHistory, gin.H{
					"ID":           log.ID,
					"old_status":   log.OldStatus,
					"new_status":   log.NewStatus,
					"reason":       log.Reason,
					"moderated_at": log.ModeratedAt,
					"admin": gin.H{
						"ID":         log.Admin.ID,
						"first_name": log.Admin.FirstName,
						"last_name":  log.Admin.LastName,
						"name":       log.Admin.FirstName + " " + log.Admin.LastName,
						"email":      log.Admin.Email,
					},
				})
			}
			responseData["moderation_history"] = moderationHistory
		}
	}

	// Include seller response if exists
	if review.SellerResponse != nil {
		responseData["seller_response"] = gin.H{
			"ID":      review.SellerResponse.ID,
			"content": review.SellerResponse.Content,
			"user": gin.H{
				"ID":         review.SellerResponse.User.ID,
				"first_name": review.SellerResponse.User.FirstName,
				"last_name":  review.SellerResponse.User.LastName,
				"name":       review.SellerResponse.User.FirstName + " " + review.SellerResponse.User.LastName,
				"email":      review.SellerResponse.User.Email,
				"phone":      review.SellerResponse.User.Phone,
				"avatar":     review.SellerResponse.User.Avatar,
			},
			"CreatedAt": review.SellerResponse.CreatedAt,
			"UpdatedAt": review.SellerResponse.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responseData,
	})
}

// GetProductReviews handles GET /api/v1/reviews/product/:productVariantId
// Returns paginated reviews for a product with filtering options
func (h *ReviewHandler) GetProductReviews(c *gin.Context) {
	// Parse product variant ID
	productVariantID, err := strconv.ParseUint(c.Param("productVariantId"), 10, 32)
	if err != nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_PRODUCT_VARIANT_ID", "Invalid product variant ID")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	sortBy := c.DefaultQuery("sort", "CreatedAt")
	sortOrder := c.DefaultQuery("order", "desc")

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Parse rating filter
	ratingFilter := c.Query("rating")
	var rating int
	if ratingFilter != "" {
		rating, err = strconv.Atoi(ratingFilter)
		if err != nil || rating < 1 || rating > 5 {
			response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_RATING_FILTER", "Invalid rating filter. Must be 1-5")
			return
		}
	}

	// Validate sort parameters
	allowedSortFields := map[string]bool{
		"CreatedAt":     true,
		"rating":        true,
		"helpful_count": true,
		"UpdatedAt":     true,
	}
	if !allowedSortFields[sortBy] {
		sortBy = "CreatedAt"
	}

	allowedSortOrders := map[string]bool{
		"asc":  true,
		"desc": true,
	}
	if !allowedSortOrders[sortOrder] {
		sortOrder = "desc"
	}

	// Build query
	query := h.db.Model(&models.ProductReview{}).
		Where("product_variant_id = ? AND status = ?", productVariantID, models.ReviewStatusApproved)

	// Apply rating filter
	if rating > 0 {
		query = query.Where("rating = ?", rating)
	}

	// Get total count for pagination
	var total int64
	err = query.Count(&total).Error
	if err != nil {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "COUNT_REVIEWS_ERROR", "Failed to count reviews")
		return
	}

	// Get reviews with pagination and sorting
	var reviews []models.ProductReview
	err = query.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, first_name, last_name, email, phone, avatar")
	}).
		Preload("Images").
		Preload("SellerResponse.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, first_name, last_name, email, phone, avatar")
		}).
		Order(sortBy + " " + strings.ToUpper(sortOrder)).
		Offset(offset).
		Limit(limit).
		Find(&reviews).Error

	if err != nil {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "RETRIEVE_REVIEWS_ERROR", "Failed to retrieve reviews")
		return
	}

	// Format reviews for response
	var formattedReviews []gin.H
	for _, review := range reviews {
		reviewData := gin.H{
			"ID": review.ID,
			"user": gin.H{
				"ID":         review.User.ID,
				"first_name": review.User.FirstName,
				"last_name":  review.User.LastName,
				"name":       review.GetReviewerName(),
				"email":      review.User.Email,
				"phone":      review.User.Phone,
				"avatar":     review.User.Avatar,
			},
			"rating":               review.Rating,
			"title":                review.Title,
			"content":              review.Content,
			"is_verified_purchase": review.IsVerifiedPurchase,
			"helpful_count":        review.HelpfulCount,
			"images":               review.Images,
			"CreatedAt":            review.CreatedAt,
			"UpdatedAt":            review.UpdatedAt,
		}

		// Include seller response if exists
		if review.SellerResponse != nil {
			reviewData["seller_response"] = gin.H{
				"ID":      review.SellerResponse.ID,
				"content": review.SellerResponse.Content,
				"user": gin.H{
					"ID":         review.SellerResponse.User.ID,
					"first_name": review.SellerResponse.User.FirstName,
					"last_name":  review.SellerResponse.User.LastName,
					"name":       review.SellerResponse.User.FirstName + " " + review.SellerResponse.User.LastName,
					"email":      review.SellerResponse.User.Email,
					"phone":      review.SellerResponse.User.Phone,
					"avatar":     review.SellerResponse.User.Avatar,
				},
				"CreatedAt": review.SellerResponse.CreatedAt,
				"UpdatedAt": review.SellerResponse.UpdatedAt,
			}
		}

		formattedReviews = append(formattedReviews, reviewData)
	}

	// Calculate pagination info
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrev := page > 1

	// Get rating statistics for this product
	var ratingStats models.ProductRating
	err = h.db.Where("product_variant_id = ?", productVariantID).First(&ratingStats).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "RETRIEVE_RATING_STATS_ERROR", "Failed to retrieve rating statistics")
		return
	}

	// Parse rating breakdown JSON
	var ratingBreakdown gin.H
	if ratingStats.RatingBreakdown != "" {
		var breakdown map[string]int
		if err := json.Unmarshal([]byte(ratingStats.RatingBreakdown), &breakdown); err != nil {
			// If parsing fails, use empty breakdown
			ratingBreakdown = gin.H{
				"1": 0,
				"2": 0,
				"3": 0,
				"4": 0,
				"5": 0,
			}
		} else {
			// Convert to gin.H format
			ratingBreakdown = gin.H{
				"1": breakdown["1"],
				"2": breakdown["2"],
				"3": breakdown["3"],
				"4": breakdown["4"],
				"5": breakdown["5"],
			}
		}
	} else {
		// No rating breakdown available
		ratingBreakdown = gin.H{
			"1": 0,
			"2": 0,
			"3": 0,
			"4": 0,
			"5": 0,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"reviews": formattedReviews,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
				"has_next":    hasNext,
				"has_prev":    hasPrev,
			},
			"filters": gin.H{
				"rating": ratingFilter,
				"sort":   sortBy,
				"order":  sortOrder,
			},
			"rating_stats": gin.H{
				"average_rating":   ratingStats.AverageRating,
				"total_reviews":    ratingStats.TotalReviews,
				"rating_breakdown": ratingBreakdown,
			},
		},
	})
}
