package review

import (
	"fmt"
	"strings"

	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

// ReviewError represents a review-specific error
type ReviewError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description,omitempty"`
	Field       string `json:"field,omitempty"`
}

// Error constants for review operations
const (
	// General review errors
	ErrReviewNotFound         = "REVIEW_NOT_FOUND"
	ErrReviewAlreadyExists    = "REVIEW_ALREADY_EXISTS"
	ErrReviewNotAuthorized    = "REVIEW_NOT_AUTHORIZED"
	ErrReviewPurchaseRequired = "PURCHASE_REQUIRED"
	ErrReviewModerationFailed = "MODERATION_FAILED"
	ErrReviewInvalidStatus    = "INVALID_REVIEW_STATUS"

	// Validation errors
	ErrReviewInvalidRating  = "INVALID_RATING"
	ErrReviewInvalidContent = "INVALID_CONTENT"
	ErrReviewInvalidTitle   = "INVALID_TITLE"
	ErrReviewTooManyImages  = "TOO_MANY_IMAGES"
	ErrReviewInvalidImage   = "INVALID_IMAGE"

	// Purchase verification errors
	ErrPurchaseNotFound     = "PURCHASE_NOT_FOUND"
	ErrPurchaseNotDelivered = "PURCHASE_NOT_DELIVERED"
	ErrPurchaseTooOld       = "PURCHASE_TOO_OLD"
	ErrPurchaseNotPaid      = "PURCHASE_NOT_PAID"

	// Product errors
	ErrProductVariantNotFound = "PRODUCT_VARIANT_NOT_FOUND"
	ErrProductNotActive       = "PRODUCT_NOT_ACTIVE"

	// Seller response errors
	ErrSellerResponseNotFound = "SELLER_RESPONSE_NOT_FOUND"
	ErrSellerNotAuthorized    = "SELLER_NOT_AUTHORIZED"
	ErrSellerResponseTooLong  = "SELLER_RESPONSE_TOO_LONG"

	// Helpfulness errors
	ErrHelpfulnessAlreadyVoted = "ALREADY_VOTED"
	ErrHelpfulnessInvalidVote  = "INVALID_VOTE_TYPE"

	// Admin errors
	ErrAdminNotAuthorized      = "ADMIN_NOT_AUTHORIZED"
	ErrReviewAlreadyModerated  = "REVIEW_ALREADY_MODERATED"
	ErrInvalidModerationAction = "INVALID_MODERATION_ACTION"

	// File upload errors
	ErrFileUploadFailed = "FILE_UPLOAD_FAILED"
	ErrFileTooLarge     = "FILE_TOO_LARGE"
	ErrInvalidFileType  = "INVALID_FILE_TYPE"
	ErrTooManyFiles     = "TOO_MANY_FILES"
)

// Error messages
var errorMessages = map[string]string{
	ErrReviewNotFound:         "Review not found",
	ErrReviewAlreadyExists:    "You have already reviewed this product",
	ErrReviewNotAuthorized:    "You are not authorized to perform this action",
	ErrReviewPurchaseRequired: "You must purchase this product before reviewing it",
	ErrReviewModerationFailed: "Failed to moderate review",
	ErrReviewInvalidStatus:    "Invalid review status",

	ErrReviewInvalidRating:  "Rating must be between 1 and 5",
	ErrReviewInvalidContent: "Review content is required and must be less than 1000 characters",
	ErrReviewInvalidTitle:   "Review title must be less than 100 characters",
	ErrReviewTooManyImages:  "Maximum 5 images allowed per review",
	ErrReviewInvalidImage:   "Invalid image format or URL",

	ErrPurchaseNotFound:     "No purchase found for this product",
	ErrPurchaseNotDelivered: "Product must be delivered before reviewing",
	ErrPurchaseTooOld:       "Product must be purchased within the last 2 years to review",
	ErrPurchaseNotPaid:      "Product must be paid for before reviewing",

	ErrProductVariantNotFound: "Product variant not found",
	ErrProductNotActive:       "Product is not available for review",

	ErrSellerResponseNotFound: "Seller response not found",
	ErrSellerNotAuthorized:    "You are not authorized to respond to this review",
	ErrSellerResponseTooLong:  "Seller response must be less than 500 characters",

	ErrHelpfulnessAlreadyVoted: "You have already voted on this review",
	ErrHelpfulnessInvalidVote:  "Invalid vote type. Must be 'helpful' or 'unhelpful'",

	ErrAdminNotAuthorized:      "Admin access required",
	ErrReviewAlreadyModerated:  "Review has already been moderated",
	ErrInvalidModerationAction: "Invalid moderation action",

	ErrFileUploadFailed: "Failed to upload file",
	ErrFileTooLarge:     "File size too large (max 5MB per file)",
	ErrInvalidFileType:  "Invalid file type. Only images are allowed",
	ErrTooManyFiles:     "Too many files uploaded",
}

// NewReviewError creates a new ReviewError
func NewReviewError(code string, description ...string) *ReviewError {
	msg := errorMessages[code]
	if msg == "" {
		msg = "Unknown error occurred"
	}

	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}

	return &ReviewError{
		Code:        code,
		Message:     msg,
		Description: desc,
	}
}

// NewReviewErrorWithField creates a new ReviewError with a specific field
func NewReviewErrorWithField(code string, field string, description ...string) *ReviewError {
	err := NewReviewError(code, description...)
	err.Field = field
	return err
}

// ValidationError represents multiple validation errors
type ValidationError struct {
	Errors []*ReviewError `json:"errors"`
}

// NewValidationError creates a new ValidationError
func NewValidationError() *ValidationError {
	return &ValidationError{
		Errors: make([]*ReviewError, 0),
	}
}

// AddError adds an error to the validation error collection
func (ve *ValidationError) AddError(code string, field string, description ...string) {
	ve.Errors = append(ve.Errors, NewReviewErrorWithField(code, field, description...))
}

// HasErrors returns true if there are validation errors
func (ve *ValidationError) HasErrors() bool {
	return len(ve.Errors) > 0
}

// Error returns a string representation of the validation error
func (ve *ValidationError) Error() string {
	if len(ve.Errors) == 0 {
		return "No validation errors"
	}

	var messages []string
	for _, err := range ve.Errors {
		if err.Field != "" {
			messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
		} else {
			messages = append(messages, err.Message)
		}
	}

	return strings.Join(messages, "; ")
}

// Response helpers for review errors

// GenerateReviewErrorResponse sends a review error response
func GenerateReviewErrorResponse(c *gin.Context, status int, err *ReviewError) {
	response.GenerateErrorResponse(c, status, err.Code, err.Message)
}

// GenerateReviewBadRequestResponse sends a bad request response for review errors
func GenerateReviewBadRequestResponse(c *gin.Context, err *ReviewError) {
	GenerateReviewErrorResponse(c, 400, err)
}

// GenerateReviewUnauthorizedResponse sends an unauthorized response for review errors
func GenerateReviewUnauthorizedResponse(c *gin.Context, err *ReviewError) {
	GenerateReviewErrorResponse(c, 401, err)
}

// GenerateReviewForbiddenResponse sends a forbidden response for review errors
func GenerateReviewForbiddenResponse(c *gin.Context, err *ReviewError) {
	GenerateReviewErrorResponse(c, 403, err)
}

// GenerateReviewNotFoundResponse sends a not found response for review errors
func GenerateReviewNotFoundResponse(c *gin.Context, err *ReviewError) {
	GenerateReviewErrorResponse(c, 404, err)
}

// GenerateReviewInternalServerErrorResponse sends an internal server error response for review errors
func GenerateReviewInternalServerErrorResponse(c *gin.Context, err *ReviewError) {
	GenerateReviewErrorResponse(c, 500, err)
}

// GenerateValidationErrorResponse sends a validation error response
func GenerateValidationErrorResponse(c *gin.Context, validationErr *ValidationError) {
	c.JSON(400, gin.H{
		"status":  400,
		"message": "Validation failed",
		"error": gin.H{
			"code":    "VALIDATION_ERROR",
			"message": "One or more validation errors occurred",
			"details": validationErr.Errors,
		},
	})
}

// GenerateReviewConflictResponse sends a conflict response for review errors
func GenerateReviewConflictResponse(c *gin.Context, err *ReviewError) {
	GenerateReviewErrorResponse(c, 409, err)
}

// GenerateReviewUnprocessableEntityResponse sends an unprocessable entity response for review errors
func GenerateReviewUnprocessableEntityResponse(c *gin.Context, err *ReviewError) {
	GenerateReviewErrorResponse(c, 422, err)
}

// Convenience functions for common error scenarios

// HandleDatabaseError handles database errors and returns appropriate responses
func HandleDatabaseError(c *gin.Context, err error, operation string) {
	if err == nil {
		return
	}

	// Check for specific database errors
	switch {
	case strings.Contains(err.Error(), "duplicate key"):
		GenerateReviewConflictResponse(c, NewReviewError(ErrReviewAlreadyExists))
	case strings.Contains(err.Error(), "foreign key constraint"):
		GenerateReviewBadRequestResponse(c, NewReviewError(ErrProductVariantNotFound))
	case strings.Contains(err.Error(), "record not found"):
		GenerateReviewNotFoundResponse(c, NewReviewError(ErrReviewNotFound))
	default:
		GenerateReviewInternalServerErrorResponse(c, NewReviewError("DATABASE_ERROR", fmt.Sprintf("Database operation failed: %v", err)))
	}
}

// HandleValidationError handles validation errors and returns appropriate responses
func HandleValidationError(c *gin.Context, err error, operation string) {
	if err == nil {
		return
	}

	// Check if it's a validation error
	if validationErr, ok := err.(*ValidationError); ok {
		GenerateValidationErrorResponse(c, validationErr)
		return
	}

	// Handle other validation errors
	GenerateReviewBadRequestResponse(c, NewReviewError("VALIDATION_ERROR", err.Error()))
}

// HandleFileUploadError handles file upload errors and returns appropriate responses
func HandleFileUploadError(c *gin.Context, err error, operation string) {
	if err == nil {
		return
	}

	// Check for specific file upload errors
	switch {
	case strings.Contains(err.Error(), "file too large"):
		GenerateReviewBadRequestResponse(c, NewReviewError(ErrFileTooLarge))
	case strings.Contains(err.Error(), "invalid file type"):
		GenerateReviewBadRequestResponse(c, NewReviewError(ErrInvalidFileType))
	case strings.Contains(err.Error(), "too many files"):
		GenerateReviewBadRequestResponse(c, NewReviewError(ErrTooManyFiles))
	default:
		GenerateReviewInternalServerErrorResponse(c, NewReviewError(ErrFileUploadFailed, err.Error()))
	}
}
