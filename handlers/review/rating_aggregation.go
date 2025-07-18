package review

import (
	"encoding/json"
	"fmt"

	"github.com/YasserCherfaoui/MarketProGo/models"
)

// UpdateProductRating recalculates and updates the ProductRating for a product variant
func (h *ReviewHandler) UpdateProductRating(productVariantID uint) error {
	var reviews []models.ProductReview
	// Only count approved reviews that are not soft deleted
	err := h.db.Where("product_variant_id = ? AND status = ? AND deleted_at IS NULL", productVariantID, models.ReviewStatusApproved).Find(&reviews).Error
	if err != nil {
		return fmt.Errorf("failed to fetch reviews: %w", err)
	}

	total := len(reviews)
	if total == 0 {
		// No reviews: set rating to zero and clear breakdown
		rating := &models.ProductRating{
			ProductVariantID: productVariantID,
			AverageRating:    0,
			TotalReviews:     0,
			RatingBreakdown:  `{"1":0,"2":0,"3":0,"4":0,"5":0}`,
		}

		// Check if rating exists
		var existingRating models.ProductRating
		err := h.db.Where("product_variant_id = ?", productVariantID).First(&existingRating).Error
		if err != nil {
			if err.Error() == "record not found" {
				// Create new rating
				return h.db.Create(rating).Error
			}
			return fmt.Errorf("failed to check existing rating: %w", err)
		}

		// Update existing rating
		return h.db.Model(&existingRating).Updates(map[string]interface{}{
			"average_rating":   rating.AverageRating,
			"total_reviews":    rating.TotalReviews,
			"rating_breakdown": rating.RatingBreakdown,
		}).Error
	}

	breakdown := map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	sum := 0
	for _, r := range reviews {
		breakdown[r.Rating]++
		sum += r.Rating
	}
	avg := float64(sum) / float64(total)

	breakdownJSON, err := json.Marshal(breakdown)
	if err != nil {
		return fmt.Errorf("failed to marshal breakdown: %w", err)
	}

	rating := &models.ProductRating{
		ProductVariantID: productVariantID,
		AverageRating:    avg,
		TotalReviews:     total,
		RatingBreakdown:  string(breakdownJSON),
	}

	// Check if rating exists
	var existingRating models.ProductRating
	err = h.db.Where("product_variant_id = ?", productVariantID).First(&existingRating).Error
	if err != nil {
		if err.Error() == "record not found" {
			// Create new rating
			return h.db.Create(rating).Error
		}
		return fmt.Errorf("failed to check existing rating: %w", err)
	}

	// Update existing rating
	return h.db.Model(&existingRating).Updates(map[string]interface{}{
		"average_rating":   rating.AverageRating,
		"total_reviews":    rating.TotalReviews,
		"rating_breakdown": rating.RatingBreakdown,
	}).Error
}
