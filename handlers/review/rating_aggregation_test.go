package review

import (
	"encoding/json"
	"testing"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRatingTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.ProductVariant{},
		&models.ProductReview{},
		&models.ProductRating{},
	)
	require.NoError(t, err)
	return db
}

func TestUpdateProductRating(t *testing.T) {
	db := setupRatingTestDB(t)
	h := &ReviewHandler{db: db}

	// Use a unique SKU to ensure isolation
	const uniqueSKU = "SKU-RATING-TEST-UNIQUE-99999"

	// Create product and variant with unique SKU
	product := models.Product{Name: "Test Product Rating", IsActive: true}
	require.NoError(t, db.Create(&product).Error)
	variant := models.ProductVariant{ProductID: product.ID, Name: "Rating Test Variant", SKU: uniqueSKU, BasePrice: 10, IsActive: true}
	require.NoError(t, db.Create(&variant).Error)

	// Fetch the variant by SKU to get its ID
	var fetchedVariant models.ProductVariant
	require.NoError(t, db.Where("sku = ?", uniqueSKU).First(&fetchedVariant).Error)
	variantID := fetchedVariant.ID

	// Clear any existing ProductRating records for this variant
	db.Unscoped().Where("product_variant_id = ?", variantID).Delete(&models.ProductRating{})
	// Clear any existing reviews for this variant
	db.Unscoped().Where("product_variant_id = ?", variantID).Delete(&models.ProductReview{})

	// Verify no reviews exist for this variant
	var count int64
	db.Model(&models.ProductReview{}).Where("product_variant_id = ? AND deleted_at IS NULL", variantID).Count(&count)
	require.Equal(t, int64(0), count, "No reviews should exist for this variant initially")

	// No reviews
	require.NoError(t, h.UpdateProductRating(variantID))
	var rating models.ProductRating
	require.NoError(t, db.Where("product_variant_id = ?", variantID).First(&rating).Error)
	assert.Equal(t, 0.0, rating.AverageRating)
	assert.Equal(t, 0, rating.TotalReviews)
	var breakdown map[string]int
	require.NoError(t, json.Unmarshal([]byte(rating.RatingBreakdown), &breakdown))
	assert.Equal(t, map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0}, breakdown)

	// Add reviews: 1x5, 2x4, 1x3, 1x2, 1x1
	reviews := []models.ProductReview{
		{ProductVariantID: variantID, UserID: 1, Rating: 5, Status: models.ReviewStatusApproved},
		{ProductVariantID: variantID, UserID: 2, Rating: 4, Status: models.ReviewStatusApproved},
		{ProductVariantID: variantID, UserID: 3, Rating: 4, Status: models.ReviewStatusApproved},
		{ProductVariantID: variantID, UserID: 4, Rating: 3, Status: models.ReviewStatusApproved},
		{ProductVariantID: variantID, UserID: 5, Rating: 2, Status: models.ReviewStatusApproved},
		{ProductVariantID: variantID, UserID: 6, Rating: 1, Status: models.ReviewStatusApproved},
	}
	for _, r := range reviews {
		require.NoError(t, db.Create(&r).Error)
	}

	require.NoError(t, h.UpdateProductRating(variantID))
	require.NoError(t, db.Where("product_variant_id = ?", variantID).First(&rating).Error)
	assert.InDelta(t, 3.166, rating.AverageRating, 0.01)
	assert.Equal(t, 6, rating.TotalReviews)
	require.NoError(t, json.Unmarshal([]byte(rating.RatingBreakdown), &breakdown))
	assert.Equal(t, map[string]int{"1": 1, "2": 1, "3": 1, "4": 2, "5": 1}, breakdown)

	// Add a pending review (should not count)
	pending := models.ProductReview{ProductVariantID: variantID, UserID: 7, Rating: 5, Status: models.ReviewStatusPending}
	require.NoError(t, db.Create(&pending).Error)
	require.NoError(t, h.UpdateProductRating(variantID))
	require.NoError(t, db.Where("product_variant_id = ?", variantID).First(&rating).Error)
	assert.Equal(t, 6, rating.TotalReviews)

	// Delete a review and update
	require.NoError(t, db.Unscoped().Where("user_id = ? AND product_variant_id = ?", 1, variantID).Delete(&models.ProductReview{}).Error)
	require.NoError(t, h.UpdateProductRating(variantID))
	require.NoError(t, db.Where("product_variant_id = ?", variantID).First(&rating).Error)
	assert.Equal(t, 5, rating.TotalReviews)

	// All reviews deleted
	require.NoError(t, db.Unscoped().Where("product_variant_id = ?", variantID).Delete(&models.ProductReview{}).Error)

	require.NoError(t, h.UpdateProductRating(variantID))
	require.NoError(t, db.Where("product_variant_id = ?", variantID).First(&rating).Error)
	assert.Equal(t, 0, rating.TotalReviews)
	assert.Equal(t, 0.0, rating.AverageRating)
}
