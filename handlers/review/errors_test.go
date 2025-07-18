package review

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupErrorTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.ProductVariant{},
		&models.Order{},
		&models.OrderItem{},
		&models.ProductReview{},
		&models.ReviewImage{},
		&models.SellerResponse{},
		&models.ReviewHelpful{},
		&models.ProductRating{},
		&models.ReviewModerationLog{},
	)
	require.NoError(t, err)

	return db
}

func TestNewReviewError(t *testing.T) {
	// Test creating a review error with code only
	err := NewReviewError(ErrReviewNotFound)
	assert.Equal(t, ErrReviewNotFound, err.Code)
	assert.Equal(t, "Review not found", err.Message)
	assert.Equal(t, "", err.Description)

	// Test creating a review error with description
	err = NewReviewError(ErrReviewPurchaseRequired, "No verified purchase found")
	assert.Equal(t, ErrReviewPurchaseRequired, err.Code)
	assert.Equal(t, "You must purchase this product before reviewing it", err.Message)
	assert.Equal(t, "No verified purchase found", err.Description)
}

func TestNewReviewErrorWithField(t *testing.T) {
	err := NewReviewErrorWithField(ErrReviewInvalidRating, "rating", "Rating must be between 1 and 5")
	assert.Equal(t, ErrReviewInvalidRating, err.Code)
	assert.Equal(t, "Rating must be between 1 and 5", err.Message)
	assert.Equal(t, "rating", err.Field)
	assert.Equal(t, "Rating must be between 1 and 5", err.Description)
}

func TestValidationError(t *testing.T) {
	validationErr := NewValidationError()
	assert.False(t, validationErr.HasErrors())
	assert.Equal(t, "No validation errors", validationErr.Error())

	// Add errors
	validationErr.AddError(ErrReviewInvalidRating, "rating")
	validationErr.AddError(ErrReviewInvalidContent, "content", "Content is required")

	assert.True(t, validationErr.HasErrors())
	assert.Equal(t, 2, len(validationErr.Errors))
	assert.Contains(t, validationErr.Error(), "rating: Rating must be between 1 and 5")
	assert.Contains(t, validationErr.Error(), "content: Review content is required and must be less than 1000 characters")
}

func TestReviewValidator_ValidateCreateReviewRequest(t *testing.T) {
	validator := NewReviewValidator()

	tests := []struct {
		name     string
		request  CreateReviewRequest
		hasError bool
	}{
		{
			name: "valid request",
			request: CreateReviewRequest{
				ProductVariantID: 1,
				Rating:           5,
				Title:            "Great product",
				Content:          "This is an excellent product",
			},
			hasError: false,
		},
		{
			name: "invalid rating - too low",
			request: CreateReviewRequest{
				ProductVariantID: 1,
				Rating:           0,
				Title:            "Great product",
				Content:          "This is an excellent product",
			},
			hasError: true,
		},
		{
			name: "invalid rating - too high",
			request: CreateReviewRequest{
				ProductVariantID: 1,
				Rating:           6,
				Title:            "Great product",
				Content:          "This is an excellent product",
			},
			hasError: true,
		},
		{
			name: "missing content",
			request: CreateReviewRequest{
				ProductVariantID: 1,
				Rating:           5,
				Title:            "Great product",
				Content:          "",
			},
			hasError: true,
		},
		{
			name: "content too long",
			request: CreateReviewRequest{
				ProductVariantID: 1,
				Rating:           5,
				Title:            "Great product",
				Content:          string(make([]byte, 1001)),
			},
			hasError: true,
		},
		{
			name: "title too long",
			request: CreateReviewRequest{
				ProductVariantID: 1,
				Rating:           5,
				Title:            string(make([]byte, 101)),
				Content:          "This is an excellent product",
			},
			hasError: true,
		},
		{
			name: "too many images",
			request: CreateReviewRequest{
				ProductVariantID: 1,
				Rating:           5,
				Title:            "Great product",
				Content:          "This is an excellent product",
				Images:           []string{"img1.jpg", "img2.jpg", "img3.jpg", "img4.jpg", "img5.jpg", "img6.jpg"},
			},
			hasError: true,
		},
		{
			name: "invalid product variant ID",
			request: CreateReviewRequest{
				ProductVariantID: 0,
				Rating:           5,
				Title:            "Great product",
				Content:          "This is an excellent product",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErr := validator.ValidateCreateReviewRequest(&tt.request)
			if tt.hasError {
				assert.True(t, validationErr.HasErrors())
			} else {
				assert.False(t, validationErr.HasErrors())
			}
		})
	}
}

func TestReviewValidator_ValidateUpdateReviewRequest(t *testing.T) {
	validator := NewReviewValidator()

	tests := []struct {
		name     string
		request  UpdateReviewRequest
		hasError bool
	}{
		{
			name: "valid request",
			request: UpdateReviewRequest{
				Rating:  5,
				Title:   "Great product",
				Content: "This is an excellent product",
			},
			hasError: false,
		},
		{
			name: "invalid rating",
			request: UpdateReviewRequest{
				Rating:  0,
				Title:   "Great product",
				Content: "This is an excellent product",
			},
			hasError: true,
		},
		{
			name: "missing content",
			request: UpdateReviewRequest{
				Rating:  5,
				Title:   "Great product",
				Content: "",
			},
			hasError: true,
		},
		{
			name: "missing title",
			request: UpdateReviewRequest{
				Rating:  5,
				Title:   "",
				Content: "This is an excellent product",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErr := validator.ValidateUpdateReviewRequest(&tt.request)
			if tt.hasError {
				assert.True(t, validationErr.HasErrors())
			} else {
				assert.False(t, validationErr.HasErrors())
			}
		})
	}
}

func TestReviewValidator_ValidateSellerResponseRequest(t *testing.T) {
	validator := NewReviewValidator()

	tests := []struct {
		name     string
		request  SellerResponseRequest
		hasError bool
	}{
		{
			name: "valid request",
			request: SellerResponseRequest{
				Content: "Thank you for your feedback",
			},
			hasError: false,
		},
		{
			name: "missing content",
			request: SellerResponseRequest{
				Content: "",
			},
			hasError: true,
		},
		{
			name: "content too long",
			request: SellerResponseRequest{
				Content: string(make([]byte, 501)),
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErr := validator.ValidateSellerResponseRequest(&tt.request)
			if tt.hasError {
				assert.True(t, validationErr.HasErrors())
			} else {
				assert.False(t, validationErr.HasErrors())
			}
		})
	}
}

func TestReviewValidator_ValidateModerationRequest(t *testing.T) {
	validator := NewReviewValidator()

	tests := []struct {
		name     string
		request  ModerationRequest
		hasError bool
	}{
		{
			name: "valid approval",
			request: ModerationRequest{
				Status: models.ReviewStatusApproved,
				Reason: "Good review",
			},
			hasError: false,
		},
		{
			name: "valid rejection with reason",
			request: ModerationRequest{
				Status: models.ReviewStatusRejected,
				Reason: "Inappropriate content",
			},
			hasError: false,
		},
		{
			name: "rejection without reason",
			request: ModerationRequest{
				Status: models.ReviewStatusRejected,
				Reason: "",
			},
			hasError: true,
		},
		{
			name: "invalid status",
			request: ModerationRequest{
				Status: "INVALID_STATUS",
				Reason: "Test",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErr := validator.ValidateModerationRequest(&tt.request)
			if tt.hasError {
				assert.True(t, validationErr.HasErrors())
			} else {
				assert.False(t, validationErr.HasErrors())
			}
		})
	}
}

func TestReviewValidator_ValidateReviewOwnership(t *testing.T) {
	validator := NewReviewValidator()

	review := &models.ProductReview{
		UserID: 1,
	}

	// Test valid ownership
	err := validator.ValidateReviewOwnership(review, 1, models.Customer)
	assert.Nil(t, err)

	// Test invalid ownership
	err = validator.ValidateReviewOwnership(review, 2, models.Customer)
	assert.NotNil(t, err)
	assert.Equal(t, ErrReviewNotAuthorized, err.Code)
}

func TestReviewValidator_ValidateSellerAuthorization(t *testing.T) {
	validator := NewReviewValidator()

	review := &models.ProductReview{}

	// Test valid seller authorization
	err := validator.ValidateSellerAuthorization(review, models.Vendor)
	assert.Nil(t, err)

	// Test invalid seller authorization
	err = validator.ValidateSellerAuthorization(review, models.Customer)
	assert.NotNil(t, err)
	assert.Equal(t, ErrSellerNotAuthorized, err.Code)
}

func TestReviewValidator_ValidateAdminAuthorization(t *testing.T) {
	validator := NewReviewValidator()

	// Test valid admin authorization
	err := validator.ValidateAdminAuthorization(models.Admin)
	assert.Nil(t, err)

	// Test invalid admin authorization
	err = validator.ValidateAdminAuthorization(models.Customer)
	assert.NotNil(t, err)
	assert.Equal(t, ErrAdminNotAuthorized, err.Code)
}

func TestReviewValidator_ValidateReviewStatus(t *testing.T) {
	validator := NewReviewValidator()

	// Test approved review
	review := &models.ProductReview{
		Status: models.ReviewStatusApproved,
	}
	err := validator.ValidateReviewStatus(review)
	assert.Nil(t, err)

	// Test rejected review
	review.Status = models.ReviewStatusRejected
	err = validator.ValidateReviewStatus(review)
	assert.NotNil(t, err)
	assert.Equal(t, ErrReviewInvalidStatus, err.Code)
}

func TestGenerateValidationErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/test", func(c *gin.Context) {
		validationErr := NewValidationError()
		validationErr.AddError(ErrReviewInvalidRating, "rating")
		validationErr.AddError(ErrReviewInvalidContent, "content", "Content is required")
		GenerateValidationErrorResponse(c, validationErr)
	})

	req, err := http.NewRequest("POST", "/test", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, float64(400), response["status"])
	assert.Equal(t, "Validation failed", response["message"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "VALIDATION_ERROR", errorData["code"])
	assert.Equal(t, "One or more validation errors occurred", errorData["message"])

	details := errorData["details"].([]interface{})
	assert.Equal(t, 2, len(details))
}

func TestHandleDatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/test", func(c *gin.Context) {
		// Simulate a database error
		err := gorm.ErrRecordNotFound
		HandleDatabaseError(c, err, "test operation")
	})

	req, err := http.NewRequest("POST", "/test", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestErrorMessages(t *testing.T) {
	// Test that all error codes have corresponding messages
	errorCodes := []string{
		ErrReviewNotFound,
		ErrReviewAlreadyExists,
		ErrReviewNotAuthorized,
		ErrReviewPurchaseRequired,
		ErrReviewModerationFailed,
		ErrReviewInvalidStatus,
		ErrReviewInvalidRating,
		ErrReviewInvalidContent,
		ErrReviewInvalidTitle,
		ErrReviewTooManyImages,
		ErrReviewInvalidImage,
		ErrPurchaseNotFound,
		ErrPurchaseNotDelivered,
		ErrPurchaseTooOld,
		ErrPurchaseNotPaid,
		ErrProductVariantNotFound,
		ErrProductNotActive,
		ErrSellerResponseNotFound,
		ErrSellerNotAuthorized,
		ErrSellerResponseTooLong,
		ErrHelpfulnessAlreadyVoted,
		ErrHelpfulnessInvalidVote,
		ErrAdminNotAuthorized,
		ErrReviewAlreadyModerated,
		ErrInvalidModerationAction,
		ErrFileUploadFailed,
		ErrFileTooLarge,
		ErrInvalidFileType,
		ErrTooManyFiles,
	}

	for _, code := range errorCodes {
		t.Run(code, func(t *testing.T) {
			err := NewReviewError(code)
			assert.NotEqual(t, "Unknown error occurred", err.Message, "Error code %s should have a message", code)
		})
	}
}
