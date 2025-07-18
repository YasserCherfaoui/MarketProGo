package review

import (
	"regexp"
	"strings"

	"github.com/YasserCherfaoui/MarketProGo/models"
)

// ReviewValidator handles validation for review-related data
type ReviewValidator struct{}

// NewReviewValidator creates a new ReviewValidator
func NewReviewValidator() *ReviewValidator {
	return &ReviewValidator{}
}

// ValidateCreateReviewRequest validates a CreateReviewRequest
func (v *ReviewValidator) ValidateCreateReviewRequest(req *CreateReviewRequest) *ValidationError {
	validationErr := NewValidationError()

	// Validate rating
	if req.Rating < 1 || req.Rating > 5 {
		validationErr.AddError(ErrReviewInvalidRating, "rating")
	}

	// Validate content
	if strings.TrimSpace(req.Content) == "" {
		validationErr.AddError(ErrReviewInvalidContent, "content", "Review content is required")
	} else if len(req.Content) > 1000 {
		validationErr.AddError(ErrReviewInvalidContent, "content", "Review content must be less than 1000 characters")
	}

	// Validate title (optional but if provided, must be valid)
	if req.Title != "" && len(req.Title) > 100 {
		validationErr.AddError(ErrReviewInvalidTitle, "title")
	}

	// Validate images
	if len(req.Images) > 5 {
		validationErr.AddError(ErrReviewTooManyImages, "images")
	}

	// Validate image URLs if provided
	for i, imageURL := range req.Images {
		if imageURL != "" && !v.isValidImageURL(imageURL) {
			validationErr.AddError(ErrReviewInvalidImage, "images", "Invalid image URL at index "+string(rune(i)))
		}
	}

	// Validate product variant ID
	if req.ProductVariantID == 0 {
		validationErr.AddError(ErrProductVariantNotFound, "product_variant_id")
	}

	return validationErr
}

// ValidateUpdateReviewRequest validates an UpdateReviewRequest
func (v *ReviewValidator) ValidateUpdateReviewRequest(req *UpdateReviewRequest) *ValidationError {
	validationErr := NewValidationError()

	// Validate rating
	if req.Rating < 1 || req.Rating > 5 {
		validationErr.AddError(ErrReviewInvalidRating, "rating")
	}

	// Validate content
	if strings.TrimSpace(req.Content) == "" {
		validationErr.AddError(ErrReviewInvalidContent, "content", "Review content is required")
	} else if len(req.Content) > 1000 {
		validationErr.AddError(ErrReviewInvalidContent, "content", "Review content must be less than 1000 characters")
	}

	// Validate title
	if strings.TrimSpace(req.Title) == "" {
		validationErr.AddError(ErrReviewInvalidTitle, "title", "Review title is required")
	} else if len(req.Title) > 100 {
		validationErr.AddError(ErrReviewInvalidTitle, "title")
	}

	return validationErr
}

// ValidateSellerResponseRequest validates a SellerResponseRequest
func (v *ReviewValidator) ValidateSellerResponseRequest(req *SellerResponseRequest) *ValidationError {
	validationErr := NewValidationError()

	// Validate content
	if strings.TrimSpace(req.Content) == "" {
		validationErr.AddError(ErrSellerResponseTooLong, "content", "Seller response content is required")
	} else if len(req.Content) > 500 {
		validationErr.AddError(ErrSellerResponseTooLong, "content")
	}

	return validationErr
}

// ValidateHelpfulnessRequest validates a MarkReviewHelpfulRequest
func (v *ReviewValidator) ValidateHelpfulnessRequest(req *MarkReviewHelpfulRequest) *ValidationError {
	validationErr := NewValidationError()

	// No specific validation needed for boolean field
	return validationErr
}

// ValidateModerationRequest validates a ModerationRequest
func (v *ReviewValidator) ValidateModerationRequest(req *ModerationRequest) *ValidationError {
	validationErr := NewValidationError()

	// Validate status
	validStatuses := []models.ReviewStatus{
		models.ReviewStatusPending,
		models.ReviewStatusApproved,
		models.ReviewStatusRejected,
		models.ReviewStatusFlagged,
	}

	isValidStatus := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValidStatus = true
			break
		}
	}

	if !isValidStatus {
		validationErr.AddError(ErrInvalidModerationAction, "status")
	}

	// Validate reason if status is rejected or flagged
	if (req.Status == models.ReviewStatusRejected || req.Status == models.ReviewStatusFlagged) && strings.TrimSpace(req.Reason) == "" {
		validationErr.AddError(ErrInvalidModerationAction, "reason", "Reason is required for rejected or flagged reviews")
	}

	return validationErr
}

// ValidateFileUpload validates file upload parameters
func (v *ReviewValidator) ValidateFileUpload(files []interface{}, maxFiles int, maxSize int64) *ValidationError {
	validationErr := NewValidationError()

	// Validate number of files
	if len(files) > maxFiles {
		validationErr.AddError(ErrTooManyFiles, "files", "Too many files uploaded")
	}

	// Validate each file
	for i, file := range files {
		if fileHeader, ok := file.(interface {
			Size() int64
			Header() map[string][]string
		}); ok {
			// Validate file size
			if fileHeader.Size() > maxSize {
				validationErr.AddError(ErrFileTooLarge, "files", "File too large at index "+string(rune(i)))
			}

			// Validate file type
			headers := fileHeader.Header()
			contentTypes := headers["Content-Type"]
			contentType := ""
			if len(contentTypes) > 0 {
				contentType = contentTypes[0]
			}
			if !v.isValidImageType(contentType) {
				validationErr.AddError(ErrInvalidFileType, "files", "Invalid file type at index "+string(rune(i)))
			}
		}
	}

	return validationErr
}

// isValidImageURL checks if a URL is a valid image URL
func (v *ReviewValidator) isValidImageURL(url string) bool {
	// Basic URL validation
	urlPattern := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlPattern.MatchString(url) {
		return false
	}

	// Check for common image extensions
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp"}
	lowerURL := strings.ToLower(url)
	for _, ext := range imageExtensions {
		if strings.HasSuffix(lowerURL, ext) {
			return true
		}
	}

	return false
}

// isValidImageType checks if a content type is a valid image type
func (v *ReviewValidator) isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/bmp",
	}

	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}

	return false
}

// ValidateReviewOwnership validates if a user owns a review
func (v *ReviewValidator) ValidateReviewOwnership(review *models.ProductReview, userID uint, userType models.UserType) *ReviewError {
	// Users can only modify their own reviews
	if review.UserID != userID {
		return NewReviewError(ErrReviewNotAuthorized)
	}

	return nil
}

// ValidateSellerAuthorization validates if a seller can respond to a review
func (v *ReviewValidator) ValidateSellerAuthorization(review *models.ProductReview, userType models.UserType) *ReviewError {
	// Only vendors can respond to reviews
	if userType != models.Vendor {
		return NewReviewError(ErrSellerNotAuthorized)
	}

	return nil
}

// ValidateAdminAuthorization validates if a user has admin privileges
func (v *ReviewValidator) ValidateAdminAuthorization(userType models.UserType) *ReviewError {
	// Only admins can perform moderation actions
	if userType != models.Admin {
		return NewReviewError(ErrAdminNotAuthorized)
	}

	return nil
}

// ValidateReviewStatus validates if a review can be modified
func (v *ReviewValidator) ValidateReviewStatus(review *models.ProductReview) *ReviewError {
	// Reviews that are rejected or flagged cannot be modified
	if review.Status == models.ReviewStatusRejected || review.Status == models.ReviewStatusFlagged {
		return NewReviewError(ErrReviewInvalidStatus, "Review cannot be modified in current status")
	}

	return nil
}
