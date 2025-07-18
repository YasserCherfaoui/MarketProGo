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

// SellerResponseRequest represents the request body for creating/updating a seller response
type SellerResponseRequest struct {
	Content string `json:"content" binding:"required,max=500"`
}

// CreateSellerResponse handles POST /api/v1/reviews/:id/response
func (h *ReviewHandler) CreateSellerResponse(c *gin.Context) {
	// Get seller user ID from context
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
	var req SellerResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body or content too long")
		return
	}

	// Check if review exists and is approved
	var review models.ProductReview
	err = h.db.Preload("ProductVariant.Product").Where("id = ? AND status = ?", reviewID, models.ReviewStatusApproved).First(&review).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateErrorResponse(c, http.StatusNotFound, "REVIEW_NOT_FOUND", "Review not found or not approved")
			return
		}
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to retrieve review")
		return
	}

	// Verify seller owns the product being reviewed
	if !h.isSellerOwnerOfProduct(userID.(uint), review.ProductVariant.ProductID) {
		response.GenerateErrorResponse(c, http.StatusForbidden, "NOT_PRODUCT_OWNER", "You do not own the product for this review")
		return
	}

	// Check if a response already exists
	var existing models.SellerResponse
	err = h.db.Where("product_review_id = ?", review.ID).First(&existing).Error
	if err == nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "RESPONSE_EXISTS", "A response already exists for this review. Use update endpoint.")
		return
	}
	if err != gorm.ErrRecordNotFound {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to check existing response")
		return
	}

	// Create seller response
	sellerResponse := models.SellerResponse{
		ProductReviewID: review.ID,
		UserID:          userID.(uint),
		Content:         strings.TrimSpace(req.Content),
	}
	if err := h.db.Create(&sellerResponse).Error; err != nil {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create seller response")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Seller response created successfully",
		"data":    sellerResponse,
	})
}

// UpdateSellerResponse handles PUT /api/v1/reviews/:id/response
func (h *ReviewHandler) UpdateSellerResponse(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	reviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REVIEW_ID", "Invalid review ID")
		return
	}

	var req SellerResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body or content too long")
		return
	}

	// Check if review exists and is approved
	var review models.ProductReview
	err = h.db.Preload("ProductVariant.Product").Where("id = ? AND status = ?", reviewID, models.ReviewStatusApproved).First(&review).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateErrorResponse(c, http.StatusNotFound, "REVIEW_NOT_FOUND", "Review not found or not approved")
			return
		}
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to retrieve review")
		return
	}

	// Verify seller owns the product being reviewed
	if !h.isSellerOwnerOfProduct(userID.(uint), review.ProductVariant.ProductID) {
		response.GenerateErrorResponse(c, http.StatusForbidden, "NOT_PRODUCT_OWNER", "You do not own the product for this review")
		return
	}

	// Find existing response
	var sellerResponse models.SellerResponse
	err = h.db.Where("product_review_id = ? AND user_id = ?", review.ID, userID.(uint)).First(&sellerResponse).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateErrorResponse(c, http.StatusNotFound, "RESPONSE_NOT_FOUND", "No existing response to update")
			return
		}
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to retrieve seller response")
		return
	}

	sellerResponse.Content = strings.TrimSpace(req.Content)
	if err := h.db.Save(&sellerResponse).Error; err != nil {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update seller response")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Seller response updated successfully",
		"data":    sellerResponse,
	})
}

// isSellerOwnerOfProduct checks if the seller owns the product
func (h *ReviewHandler) isSellerOwnerOfProduct(userID, productID uint) bool {
	// Check if the user is a vendor/seller
	var user models.User
	if err := h.db.Where("id = ? AND user_type IN (?)", userID, []models.UserType{models.Vendor, models.Wholesaler}).First(&user).Error; err != nil {
		return false
	}

	// For now, allow all vendors/wholesalers to respond to all reviews
	// TODO: Implement proper product ownership when seller-product relationship is established
	return true
}
