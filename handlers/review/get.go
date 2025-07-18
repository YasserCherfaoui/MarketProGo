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

// GetReview handles GET /api/v1/reviews/:id
// Returns a single review by ID with all related data
func (h *ReviewHandler) GetReview(c *gin.Context) {
	reviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REVIEW_ID", "Invalid review ID")
		return
	}

	var review models.ProductReview
	err = h.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, first_name, last_name, email")
	}).
		Preload("ProductVariant").
		Preload("Images").
		Preload("SellerResponse.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, first_name, last_name, email")
		}).
		Where("id = ? AND status = ?", reviewID, models.ReviewStatusApproved).
		First(&review).Error

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
		"id":                 review.ID,
		"product_variant_id": review.ProductVariantID,
		"product_variant":    review.ProductVariant,
		"user": gin.H{
			"id":         review.User.ID,
			"first_name": review.User.FirstName,
			"last_name":  review.User.LastName,
			"name":       review.GetReviewerName(),
		},
		"rating":               review.Rating,
		"title":                review.Title,
		"content":              review.Content,
		"is_verified_purchase": review.IsVerifiedPurchase,
		"helpful_count":        review.HelpfulCount,
		"images":               review.Images,
		"created_at":           review.CreatedAt,
		"updated_at":           review.UpdatedAt,
	}

	// Include seller response if exists
	if review.SellerResponse != nil {
		responseData["seller_response"] = gin.H{
			"id":      review.SellerResponse.ID,
			"content": review.SellerResponse.Content,
			"user": gin.H{
				"id":         review.SellerResponse.User.ID,
				"first_name": review.SellerResponse.User.FirstName,
				"last_name":  review.SellerResponse.User.LastName,
				"name":       review.SellerResponse.User.FirstName + " " + review.SellerResponse.User.LastName,
			},
			"created_at": review.SellerResponse.CreatedAt,
			"updated_at": review.SellerResponse.UpdatedAt,
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
	sortBy := c.DefaultQuery("sort", "created_at")
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
		"created_at":    true,
		"rating":        true,
		"helpful_count": true,
		"updated_at":    true,
	}
	if !allowedSortFields[sortBy] {
		sortBy = "created_at"
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
		return db.Select("id, first_name, last_name, email")
	}).
		Preload("Images").
		Preload("SellerResponse.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, first_name, last_name, email")
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
			"id": review.ID,
			"user": gin.H{
				"id":         review.User.ID,
				"first_name": review.User.FirstName,
				"last_name":  review.User.LastName,
				"name":       review.GetReviewerName(),
			},
			"rating":               review.Rating,
			"title":                review.Title,
			"content":              review.Content,
			"is_verified_purchase": review.IsVerifiedPurchase,
			"helpful_count":        review.HelpfulCount,
			"images":               review.Images,
			"created_at":           review.CreatedAt,
			"updated_at":           review.UpdatedAt,
		}

		// Include seller response if exists
		if review.SellerResponse != nil {
			reviewData["seller_response"] = gin.H{
				"id":      review.SellerResponse.ID,
				"content": review.SellerResponse.Content,
				"user": gin.H{
					"id":         review.SellerResponse.User.ID,
					"first_name": review.SellerResponse.User.FirstName,
					"last_name":  review.SellerResponse.User.LastName,
					"name":       review.SellerResponse.User.FirstName + " " + review.SellerResponse.User.LastName,
				},
				"created_at": review.SellerResponse.CreatedAt,
				"updated_at": review.SellerResponse.UpdatedAt,
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
				"average_rating": ratingStats.AverageRating,
				"total_reviews":  ratingStats.TotalReviews,
				"rating_breakdown": func() gin.H {
					if ratingStats.RatingBreakdown != "" {
						// Parse JSON rating breakdown if available
						// For now, return empty breakdown
						return gin.H{
							"1": 0,
							"2": 0,
							"3": 0,
							"4": 0,
							"5": 0,
						}
					}
					return gin.H{
						"1": 0,
						"2": 0,
						"3": 0,
						"4": 0,
						"5": 0,
					}
				}(),
			},
		},
	})
}
