package product

import (
	"encoding/json"
	"strconv"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProductRatingSummary represents the aggregated rating data for a product
type ProductRatingSummary struct {
	AverageRating   float64                `json:"average_rating"`
	TotalReviews    int                    `json:"total_reviews"`
	RatingBreakdown map[string]int         `json:"rating_breakdown"`
	HasReviews      bool                   `json:"has_reviews"`
	VariantRatings  []ProductVariantRating `json:"variant_ratings,omitempty"`
}

// ProductVariantRating represents rating data for a specific variant
type ProductVariantRating struct {
	VariantID       uint                   `json:"variant_id"`
	VariantName     string                 `json:"variant_name"`
	AverageRating   float64                `json:"average_rating"`
	TotalReviews    int                    `json:"total_reviews"`
	RatingBreakdown map[string]int         `json:"rating_breakdown"`
	HasReviews      bool                   `json:"has_reviews"`
	RecentReviews   []models.ProductReview `json:"recent_reviews,omitempty"`
}

// ReviewIntegrationService handles review data integration with products
type ReviewIntegrationService struct {
	db *gorm.DB
}

// NewReviewIntegrationService creates a new review integration service
func NewReviewIntegrationService(db *gorm.DB) *ReviewIntegrationService {
	return &ReviewIntegrationService{db: db}
}

// GetProductRatingSummary retrieves aggregated rating data for a product
func (ris *ReviewIntegrationService) GetProductRatingSummary(productID uint) (*ProductRatingSummary, error) {
	var variants []models.ProductVariant
	err := ris.db.Where("product_id = ?", productID).Find(&variants).Error
	if err != nil {
		return nil, err
	}

	summary := &ProductRatingSummary{
		RatingBreakdown: make(map[string]int),
		VariantRatings:  make([]ProductVariantRating, 0),
	}

	var totalRating float64
	var totalReviews int
	allRatingBreakdown := make(map[string]int)

	// Get rating data for each variant
	for _, variant := range variants {
		variantRating, err := ris.GetVariantRatingSummary(variant.ID)
		if err != nil {
			continue // Skip variants with errors
		}

		summary.VariantRatings = append(summary.VariantRatings, *variantRating)
		totalRating += variantRating.AverageRating * float64(variantRating.TotalReviews)
		totalReviews += variantRating.TotalReviews

		// Aggregate rating breakdown
		for rating, count := range variantRating.RatingBreakdown {
			allRatingBreakdown[rating] += count
		}
	}

	// Calculate overall product rating
	if totalReviews > 0 {
		summary.AverageRating = totalRating / float64(totalReviews)
		summary.HasReviews = true
	} else {
		summary.AverageRating = 0.0
		summary.HasReviews = false
	}

	summary.TotalReviews = totalReviews
	summary.RatingBreakdown = allRatingBreakdown

	return summary, nil
}

// GetVariantRatingSummary retrieves rating data for a specific variant
func (ris *ReviewIntegrationService) GetVariantRatingSummary(variantID uint) (*ProductVariantRating, error) {
	var rating models.ProductRating
	err := ris.db.Where("product_variant_id = ?", variantID).First(&rating).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return empty rating data if no rating exists
			return &ProductVariantRating{
				VariantID:       variantID,
				AverageRating:   0.0,
				TotalReviews:    0,
				RatingBreakdown: make(map[string]int),
				HasReviews:      false,
			}, nil
		}
		return nil, err
	}

	// Parse rating breakdown JSON
	var ratingBreakdown map[string]int
	if rating.RatingBreakdown != "" {
		err = json.Unmarshal([]byte(rating.RatingBreakdown), &ratingBreakdown)
		if err != nil {
			ratingBreakdown = make(map[string]int)
		}
	} else {
		ratingBreakdown = make(map[string]int)
	}

	// Get recent reviews for this variant
	var recentReviews []models.ProductReview
	ris.db.Where("product_variant_id = ? AND status = ?", variantID, models.ReviewStatusApproved).
		Preload("User").
		Preload("Images").
		Preload("SellerResponse").
		Order("created_at DESC").
		Limit(3).
		Find(&recentReviews)

	return &ProductVariantRating{
		VariantID:       variantID,
		AverageRating:   rating.AverageRating,
		TotalReviews:    rating.TotalReviews,
		RatingBreakdown: ratingBreakdown,
		HasReviews:      rating.TotalReviews > 0,
		RecentReviews:   recentReviews,
	}, nil
}

// GetVariantRatingSummaryWithName retrieves rating data for a specific variant with name
func (ris *ReviewIntegrationService) GetVariantRatingSummaryWithName(variantID uint, variantName string) (*ProductVariantRating, error) {
	rating, err := ris.GetVariantRatingSummary(variantID)
	if err != nil {
		return nil, err
	}
	rating.VariantName = variantName
	return rating, nil
}

// AddReviewDataToProduct adds review data to a product response
func (ris *ReviewIntegrationService) AddReviewDataToProduct(product *models.Product) error {
	summary, err := ris.GetProductRatingSummary(product.ID)
	if err != nil {
		return err
	}

	// Add rating summary to product variants
	for i := range product.Variants {
		for j := range summary.VariantRatings {
			if product.Variants[i].ID == summary.VariantRatings[j].VariantID {
				// Add rating data to variant
				product.Variants[i].RatingSummary = &summary.VariantRatings[j]
				break
			}
		}
	}

	// Add overall product rating summary
	product.RatingSummary = summary
	return nil
}

// AddReviewDataToProducts adds review data to multiple products
func (ris *ReviewIntegrationService) AddReviewDataToProducts(products []models.Product) error {
	for i := range products {
		err := ris.AddReviewDataToProduct(&products[i])
		if err != nil {
			// Continue processing other products even if one fails
			continue
		}
	}
	return nil
}

// GetProductReviewStats retrieves review statistics for a product
func (ris *ReviewIntegrationService) GetProductReviewStats(productID uint) (map[string]interface{}, error) {
	var stats map[string]interface{}

	// Get total reviews by status
	var statusCounts []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	err := ris.db.Model(&models.ProductReview{}).
		Joins("JOIN product_variants ON product_variants.id = product_reviews.product_variant_id").
		Where("product_variants.product_id = ?", productID).
		Select("product_reviews.status, COUNT(*) as count").
		Group("product_reviews.status").
		Scan(&statusCounts).Error

	if err != nil {
		return nil, err
	}

	stats = make(map[string]interface{})
	stats["status_breakdown"] = statusCounts

	// Get average rating by variant
	var variantRatings []struct {
		VariantID     uint    `json:"variant_id"`
		VariantName   string  `json:"variant_name"`
		AverageRating float64 `json:"average_rating"`
		TotalReviews  int     `json:"total_reviews"`
	}

	err = ris.db.Model(&models.ProductRating{}).
		Joins("JOIN product_variants ON product_variants.id = product_ratings.product_variant_id").
		Where("product_variants.product_id = ?", productID).
		Select("product_variants.id as variant_id, product_variants.name as variant_name, product_ratings.average_rating, product_ratings.total_reviews").
		Scan(&variantRatings).Error

	if err != nil {
		return nil, err
	}

	stats["variant_ratings"] = variantRatings

	return stats, nil
}

// GetProductReviewStatsHandler handles the HTTP request for product review statistics
func (ris *ReviewIntegrationService) GetProductReviewStatsHandler(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid product ID"})
		return
	}

	stats, err := ris.GetProductReviewStats(uint(productID))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get review statistics"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    stats,
	})
}
