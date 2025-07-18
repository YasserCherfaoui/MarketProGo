package review

import (
	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ReviewHandler handles all review-related HTTP requests
type ReviewHandler struct {
	db              *gorm.DB
	appwriteService *aw.AppwriteService
	validator       *ReviewValidator
}

// NewReviewHandler creates a new instance of ReviewHandler
func NewReviewHandler(db *gorm.DB, appwriteService *aw.AppwriteService) *ReviewHandler {
	return &ReviewHandler{
		db:              db,
		appwriteService: appwriteService,
		validator:       NewReviewValidator(),
	}
}

// GetReview and GetProductReviews are implemented in get.go

// CreateReview is implemented in create.go

// UpdateReview is implemented in manage.go

// DeleteReview is implemented in manage.go

// MarkReviewHelpful is implemented in helpful.go

// CreateSellerResponse and UpdateSellerResponse are implemented in response.go

// GetAllReviews is implemented in admin.go

// ModerateReview is implemented in admin.go

// AdminDeleteReview is implemented in admin.go

// GetModerationStats is implemented in admin.go

// GetSellerReviews handles GET /api/v1/seller/reviews
func (h *ReviewHandler) GetSellerReviews(c *gin.Context) {
	c.JSON(501, gin.H{"message": "Not implemented yet"})
}
