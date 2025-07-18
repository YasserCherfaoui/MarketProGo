package product

import (
	"encoding/json"
	"fmt"
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

func setupReviewIntegrationTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Brand{},
		&models.Category{},
		&models.Product{},
		&models.ProductVariant{},
		&models.ProductRating{},
		&models.ProductReview{},
		&models.ReviewImage{},
		&models.SellerResponse{},
		&models.ReviewHelpful{},
		&models.ReviewModerationLog{},
	)
	require.NoError(t, err)

	return db
}

func createTestProductWithVariants(t *testing.T, db *gorm.DB, suffix string) (*models.Product, []models.ProductVariant) {
	// Create test brand
	brand := models.Brand{
		Name: "Test Brand " + suffix,
		Slug: "test-brand-" + suffix,
	}
	err := db.Create(&brand).Error
	assert.NoError(t, err)

	// Create test product
	product := models.Product{
		Name:        "Test Product " + suffix,
		Description: "Test Description " + suffix,
		BrandID:     &brand.ID,
		IsActive:    true,
	}
	err = db.Create(&product).Error
	assert.NoError(t, err)

	// Create test variants
	variants := []models.ProductVariant{
		{
			ProductID: product.ID,
			Name:      "Small",
			SKU:       "TEST-SMALL-" + suffix,
			BasePrice: 10.0,
		},
		{
			ProductID: product.ID,
			Name:      "Large",
			SKU:       "TEST-LARGE-" + suffix,
			BasePrice: 20.0,
		},
	}

	for i := range variants {
		err = db.Create(&variants[i]).Error
		assert.NoError(t, err)
	}

	return &product, variants
}

func createTestUser(t *testing.T, db *gorm.DB) *models.User {
	user := models.User{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		UserType:  models.Customer,
	}
	err := db.Create(&user).Error
	assert.NoError(t, err)
	return &user
}

func createTestReviews(t *testing.T, db *gorm.DB, variantID uint, userID uint) {
	reviews := []models.ProductReview{
		{
			ProductVariantID: variantID,
			UserID:           userID,
			Rating:           5,
			Title:            "Great Product",
			Content:          "Excellent quality",
			Status:           models.ReviewStatusApproved,
		},
		{
			ProductVariantID: variantID,
			UserID:           userID,
			Rating:           4,
			Title:            "Good Product",
			Content:          "Good quality",
			Status:           models.ReviewStatusApproved,
		},
		{
			ProductVariantID: variantID,
			UserID:           userID,
			Rating:           3,
			Title:            "Average Product",
			Content:          "Average quality",
			Status:           models.ReviewStatusApproved,
		},
	}

	for i := range reviews {
		err := db.Create(&reviews[i]).Error
		assert.NoError(t, err)
	}

	// Create ProductRating record for the variant
	rating := models.ProductRating{
		ProductVariantID: variantID,
		AverageRating:    4.0, // (5+4+3)/3
		TotalReviews:     3,
		RatingBreakdown:  `{"1":0,"2":0,"3":1,"4":1,"5":1}`,
	}
	err := db.Create(&rating).Error
	assert.NoError(t, err)
}

func TestReviewIntegrationService_GetProductRatingSummary(t *testing.T) {
	db := setupReviewIntegrationTestDB(t)
	ris := NewReviewIntegrationService(db)

	// Create test data
	product, variants := createTestProductWithVariants(t, db, "1")
	user := createTestUser(t, db)

	// Create reviews for the first variant
	createTestReviews(t, db, variants[0].ID, user.ID)

	// Test getting rating summary
	summary, err := ris.GetProductRatingSummary(product.ID)
	assert.NoError(t, err)
	assert.NotNil(t, summary)

	// Verify summary data
	assert.Equal(t, 3, summary.TotalReviews)
	assert.True(t, summary.HasReviews)
	assert.Len(t, summary.VariantRatings, 2)

	// Check first variant (has reviews)
	variant1Rating := summary.VariantRatings[0]
	assert.Equal(t, variants[0].ID, variant1Rating.VariantID)
	assert.Equal(t, 3, variant1Rating.TotalReviews)
	assert.True(t, variant1Rating.HasReviews)
	assert.Len(t, variant1Rating.RecentReviews, 3)

	// Check second variant (no reviews)
	variant2Rating := summary.VariantRatings[1]
	assert.Equal(t, variants[1].ID, variant2Rating.VariantID)
	assert.Equal(t, 0, variant2Rating.TotalReviews)
	assert.False(t, variant2Rating.HasReviews)
	assert.Len(t, variant2Rating.RecentReviews, 0)
}

func TestReviewIntegrationService_GetVariantRatingSummary(t *testing.T) {
	db := setupReviewIntegrationTestDB(t)
	ris := NewReviewIntegrationService(db)

	// Create test data
	_, variants := createTestProductWithVariants(t, db, "2")
	user := createTestUser(t, db)

	// Create reviews for the first variant
	createTestReviews(t, db, variants[0].ID, user.ID)

	// Test getting variant rating summary
	summary, err := ris.GetVariantRatingSummary(variants[0].ID)
	assert.NoError(t, err)
	assert.NotNil(t, summary)

	// Verify summary data
	assert.Equal(t, variants[0].ID, summary.VariantID)
	assert.Equal(t, 3, summary.TotalReviews)
	assert.True(t, summary.HasReviews)
	assert.Len(t, summary.RecentReviews, 3)

	// Test variant with no reviews
	summary2, err := ris.GetVariantRatingSummary(variants[1].ID)
	assert.NoError(t, err)
	assert.NotNil(t, summary2)
	assert.Equal(t, variants[1].ID, summary2.VariantID)
	assert.Equal(t, 0, summary2.TotalReviews)
	assert.False(t, summary2.HasReviews)
	assert.Len(t, summary2.RecentReviews, 0)
}

func TestReviewIntegrationService_AddReviewDataToProduct(t *testing.T) {
	db := setupReviewIntegrationTestDB(t)
	ris := NewReviewIntegrationService(db)

	// Create test data
	product, variants := createTestProductWithVariants(t, db, "3")
	user := createTestUser(t, db)

	// Create reviews for the first variant
	createTestReviews(t, db, variants[0].ID, user.ID)

	// Load product with variants
	err := db.Preload("Variants").First(product, product.ID).Error
	assert.NoError(t, err)

	// Add review data
	err = ris.AddReviewDataToProduct(product)
	assert.NoError(t, err)

	// Verify product has rating summary
	assert.NotNil(t, product.RatingSummary)
	ratingSummary, ok := product.RatingSummary.(*ProductRatingSummary)
	assert.True(t, ok)
	assert.Equal(t, 3, ratingSummary.TotalReviews)

	// Verify variants have rating summaries
	assert.NotNil(t, product.Variants[0].RatingSummary)
	variantRating, ok := product.Variants[0].RatingSummary.(*ProductVariantRating)
	assert.True(t, ok)
	assert.Equal(t, 3, variantRating.TotalReviews)
}

func TestReviewIntegrationService_AddReviewDataToProducts(t *testing.T) {
	db := setupReviewIntegrationTestDB(t)
	ris := NewReviewIntegrationService(db)

	// Create test data
	_, variants1 := createTestProductWithVariants(t, db, "4")
	_, _ = createTestProductWithVariants(t, db, "5") // Create second product
	user := createTestUser(t, db)

	// Create reviews for first product
	createTestReviews(t, db, variants1[0].ID, user.ID)

	// Load products with variants
	var products []models.Product
	err := db.Preload("Variants").Find(&products).Error
	assert.NoError(t, err)

	// Add review data
	err = ris.AddReviewDataToProducts(products)
	assert.NoError(t, err)

	// Verify products have rating summaries
	assert.NotNil(t, products[0].RatingSummary)
	assert.NotNil(t, products[1].RatingSummary)

	// First product should have reviews
	ratingSummary1, ok := products[0].RatingSummary.(*ProductRatingSummary)
	assert.True(t, ok)
	assert.Equal(t, 3, ratingSummary1.TotalReviews)

	// Second product should have no reviews
	ratingSummary2, ok := products[1].RatingSummary.(*ProductRatingSummary)
	assert.True(t, ok)
	assert.Equal(t, 0, ratingSummary2.TotalReviews)
}

func TestReviewIntegrationService_GetProductReviewStats(t *testing.T) {
	db := setupReviewIntegrationTestDB(t)
	ris := NewReviewIntegrationService(db)

	// Create test data
	product, variants := createTestProductWithVariants(t, db, "6")
	user := createTestUser(t, db)

	// Create reviews with different statuses
	reviews := []models.ProductReview{
		{
			ProductVariantID: variants[0].ID,
			UserID:           user.ID,
			Rating:           5,
			Title:            "Great Product",
			Content:          "Excellent quality",
			Status:           models.ReviewStatusApproved,
		},
		{
			ProductVariantID: variants[0].ID,
			UserID:           user.ID,
			Rating:           4,
			Title:            "Good Product",
			Content:          "Good quality",
			Status:           models.ReviewStatusPending,
		},
		{
			ProductVariantID: variants[1].ID,
			UserID:           user.ID,
			Rating:           3,
			Title:            "Average Product",
			Content:          "Average quality",
			Status:           models.ReviewStatusApproved,
		},
	}

	for i := range reviews {
		err := db.Create(&reviews[i]).Error
		assert.NoError(t, err)
	}

	// Create ProductRating records for the variants
	rating1 := models.ProductRating{
		ProductVariantID: variants[0].ID,
		AverageRating:    4.5, // (5+4)/2
		TotalReviews:     2,
		RatingBreakdown:  `{"1":0,"2":0,"3":0,"4":1,"5":1}`,
	}
	err := db.Create(&rating1).Error
	assert.NoError(t, err)

	rating2 := models.ProductRating{
		ProductVariantID: variants[1].ID,
		AverageRating:    3.0,
		TotalReviews:     1,
		RatingBreakdown:  `{"1":0,"2":0,"3":1,"4":0,"5":0}`,
	}
	err = db.Create(&rating2).Error
	assert.NoError(t, err)

	// Test getting review stats
	stats, err := ris.GetProductReviewStats(product.ID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	// Verify stats structure
	assert.Contains(t, stats, "status_breakdown")
	assert.Contains(t, stats, "variant_ratings")

	// Verify status breakdown
	statusBreakdown := stats["status_breakdown"].([]struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	})
	assert.Len(t, statusBreakdown, 2) // approved and pending

	// Verify variant ratings
	variantRatings := stats["variant_ratings"].([]struct {
		VariantID     uint    `json:"variant_id"`
		VariantName   string  `json:"variant_name"`
		AverageRating float64 `json:"average_rating"`
		TotalReviews  int     `json:"total_reviews"`
	})
	assert.Len(t, variantRatings, 2)
}

func TestProductHandler_GetProductReviewStats(t *testing.T) {
	db := setupReviewIntegrationTestDB(t)
	handler := &ProductHandler{
		db:            db,
		reviewService: NewReviewIntegrationService(db),
	}

	// Create test data
	product, variants := createTestProductWithVariants(t, db, "7")
	user := createTestUser(t, db)

	// Create reviews
	createTestReviews(t, db, variants[0].ID, user.ID)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/products/:id/review-stats", handler.GetProductReviewStats)

	// Create request
	req, err := http.NewRequest("GET", fmt.Sprintf("/products/%d/review-stats", product.ID), nil)
	assert.NoError(t, err)

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Contains(t, response, "data")
}

func TestReviewIntegrationService_GetProductRatingSummary_NoReviews(t *testing.T) {
	db := setupReviewIntegrationTestDB(t)
	ris := NewReviewIntegrationService(db)

	// Create test data without reviews
	product, _ := createTestProductWithVariants(t, db, "8")

	// Test getting rating summary
	summary, err := ris.GetProductRatingSummary(product.ID)
	assert.NoError(t, err)
	assert.NotNil(t, summary)

	// Verify summary data for product with no reviews
	assert.Equal(t, 0, summary.TotalReviews)
	assert.False(t, summary.HasReviews)
	assert.Equal(t, 0.0, summary.AverageRating)
	assert.Len(t, summary.VariantRatings, 2)

	// Check that all variants have no reviews
	for _, variantRating := range summary.VariantRatings {
		assert.Equal(t, 0, variantRating.TotalReviews)
		assert.False(t, variantRating.HasReviews)
		assert.Equal(t, 0.0, variantRating.AverageRating)
	}
}

func TestReviewIntegrationService_GetVariantRatingSummary_NonExistentVariant(t *testing.T) {
	db := setupReviewIntegrationTestDB(t)
	ris := NewReviewIntegrationService(db)

	// Test getting rating summary for non-existent variant
	summary, err := ris.GetVariantRatingSummary(99999)
	assert.NoError(t, err)
	assert.NotNil(t, summary)

	// Verify empty summary data
	assert.Equal(t, uint(99999), summary.VariantID)
	assert.Equal(t, 0, summary.TotalReviews)
	assert.False(t, summary.HasReviews)
	assert.Equal(t, 0.0, summary.AverageRating)
	assert.Len(t, summary.RecentReviews, 0)
}
