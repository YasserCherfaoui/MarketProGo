package review

import (
	"fmt"
	"strconv"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateReviewRequest represents the request body for creating a review
type CreateReviewRequest struct {
	ProductVariantID uint     `json:"product_variant_id"`
	OrderItemID      *uint    `json:"order_item_id"` // Optional, for specific order verification
	Rating           int      `json:"rating"`
	Title            string   `json:"title"`
	Content          string   `json:"content"`
	Images           []string `json:"images"` // Array of image URLs (handled separately)
}

// CreateReviewResponse represents the response for a created review
type CreateReviewResponse struct {
	Review *models.ProductReview `json:"review"`
}

// CreateReview handles POST /api/v1/reviews
// Allows authenticated users to submit reviews for products they have purchased
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "review/create", "user not authenticated")
		return
	}
	userID := userIDInterface.(uint)

	// Parse request body
	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Handle JSON binding errors
		GenerateReviewBadRequestResponse(c, NewReviewError("VALIDATION_ERROR", "Invalid request format"))
		return
	}

	// Validate request data
	if validationErr := h.validator.ValidateCreateReviewRequest(&req); validationErr.HasErrors() {
		GenerateValidationErrorResponse(c, validationErr)
		return
	}

	// Validate product variant exists
	var productVariant models.ProductVariant
	if err := h.db.Preload("Product").First(&productVariant, req.ProductVariantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			GenerateReviewNotFoundResponse(c, NewReviewError(ErrProductVariantNotFound))
			return
		}
		GenerateReviewInternalServerErrorResponse(c, NewReviewError("DATABASE_ERROR", "Failed to verify product variant"))
		return
	}

	// Verify purchase
	var purchaseResult *PurchaseVerificationResult
	var err error

	if req.OrderItemID != nil {
		// Verify purchase using specific order item
		purchaseResult, err = h.VerifyPurchaseWithOrderItemID(userID, *req.OrderItemID)
	} else {
		// Verify purchase using product variant
		purchaseResult, err = h.VerifyPurchase(userID, req.ProductVariantID)
	}

	if err != nil {
		GenerateReviewInternalServerErrorResponse(c, NewReviewError("DATABASE_ERROR", "Failed to verify purchase"))
		return
	}

	if !purchaseResult.IsVerified {
		GenerateReviewForbiddenResponse(c, NewReviewError(ErrReviewPurchaseRequired, purchaseResult.ErrorMessage))
		return
	}

	// Check if user has already reviewed this product variant
	var existingReview models.ProductReview
	err = h.db.Where("user_id = ? AND product_variant_id = ?", userID, req.ProductVariantID).
		First(&existingReview).Error

	if err == nil {
		GenerateReviewConflictResponse(c, NewReviewError(ErrReviewAlreadyExists))
		return
	}

	if err != gorm.ErrRecordNotFound {
		GenerateReviewInternalServerErrorResponse(c, NewReviewError("DATABASE_ERROR", "Failed to check existing review"))
		return
	}

	// Handle image uploads if any
	var reviewImages []models.ReviewImage
	if len(req.Images) > 0 {
		// Validate image count (max 5 images per review)
		if len(req.Images) > 5 {
			GenerateReviewBadRequestResponse(c, NewReviewError(ErrReviewTooManyImages))
			return
		}

		// Process each image URL
		for i, imageURL := range req.Images {
			// Validate URL format (basic validation)
			if imageURL == "" {
				continue
			}

			// Create review image record
			reviewImage := models.ReviewImage{
				URL:     imageURL,
				AltText: fmt.Sprintf("Review image %d", i+1),
			}
			reviewImages = append(reviewImages, reviewImage)
		}
	}

	// Create the review
	review := &models.ProductReview{
		ProductVariantID:   req.ProductVariantID,
		UserID:             userID,
		OrderItemID:        req.OrderItemID,
		Rating:             req.Rating,
		Title:              req.Title,
		Content:            req.Content,
		IsVerifiedPurchase: true,
		Status:             models.ReviewStatusPending, // Default to pending for moderation
		Images:             reviewImages,
	}

	// Save review to database
	if err := h.db.Create(review).Error; err != nil {
		HandleDatabaseError(c, err, "create review")
		return
	}

	// Preload related data for response
	if err := h.db.Preload("User").
		Preload("ProductVariant").
		Preload("ProductVariant.Product").
		Preload("Images").
		First(review, review.ID).Error; err != nil {
		GenerateReviewInternalServerErrorResponse(c, NewReviewError("DATABASE_ERROR", "Failed to load review data"))
		return
	}

	// Return success response
	response.GenerateCreatedResponse(c, "Review submitted successfully", CreateReviewResponse{
		Review: review,
	})

	// Trigger rating aggregation (if review is auto-approved, otherwise will be triggered on moderation)
	_ = h.UpdateProductRating(req.ProductVariantID)
}

// UploadReviewImages handles POST /api/v1/reviews/upload-images
// Allows users to upload images for their reviews
func (h *ReviewHandler) UploadReviewImages(c *gin.Context) {
	// Get user from context (authentication check)
	_, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "review/upload-images", "user not authenticated")
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		response.GenerateBadRequestResponse(c, "review/upload-images", "Failed to parse form data")
		return
	}

	files := c.Request.MultipartForm.File["images"]
	if len(files) == 0 {
		response.GenerateBadRequestResponse(c, "review/upload-images", "No images provided")
		return
	}

	// Validate number of files (max 5)
	if len(files) > 5 {
		response.GenerateBadRequestResponse(c, "review/upload-images", "Maximum 5 images allowed")
		return
	}

	var uploadedImages []string

	// Process each uploaded file
	for _, fileHeader := range files {
		// Validate file size (max 5MB per file)
		if fileHeader.Size > 5<<20 { // 5MB
			response.GenerateBadRequestResponse(c, "review/upload-images", "File size too large (max 5MB per file)")
			return
		}

		// Validate file type
		contentType := fileHeader.Header.Get("Content-Type")
		if !isValidImageType(contentType) {
			response.GenerateBadRequestResponse(c, "review/upload-images", "Invalid file type. Only images are allowed")
			return
		}

		// Upload to Appwrite storage
		fileID, err := h.appwriteService.UploadFile(fileHeader)
		if err != nil {
			response.GenerateInternalServerErrorResponse(c, "review/upload-images", "Failed to upload image")
			return
		}

		// Construct the file URL
		fileURL := h.appwriteService.GetFileURL(fileID)
		uploadedImages = append(uploadedImages, fileURL)
	}

	// Return success response with uploaded image URLs
	response.GenerateSuccessResponse(c, "Images uploaded successfully", gin.H{
		"images": uploadedImages,
	})
}

// isValidImageType checks if the content type is a valid image type
func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

// GetReviewableProducts handles GET /api/v1/reviews/reviewable-products
// Returns a list of products the user can review (purchased but not yet reviewed)
func (h *ReviewHandler) GetReviewableProducts(c *gin.Context) {
	// Get user from context
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "review/reviewable-products", "user not authenticated")
		return
	}
	userID := userIDInterface.(uint)

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	// Get reviewable products
	reviewableItems, err := h.GetReviewableProductsForUser(userID, limit)
	if err != nil {
		response.GenerateInternalServerErrorResponse(c, "review/reviewable-products", "Failed to get reviewable products")
		return
	}

	// Format response data
	var responseData []gin.H
	for _, item := range reviewableItems {
		responseData = append(responseData, gin.H{
			"order_item_id":      item.ID,
			"product_variant_id": item.ProductVariantID,
			"product_variant":    item.ProductVariant,
			"product":            item.ProductVariant.Product,
			"order":              item.Order,
			"purchased_date":     item.Order.DeliveredDate,
			"quantity":           item.Quantity,
			"unit_price":         item.UnitPrice,
		})
	}

	response.GenerateSuccessResponse(c, "Reviewable products retrieved successfully", gin.H{
		"products": responseData,
		"count":    len(responseData),
	})
}
