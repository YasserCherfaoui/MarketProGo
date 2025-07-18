package review

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGetReview(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDBWithReviewTables(t)
	handler := NewReviewHandler(db, nil)

	// Create test data
	user := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	productVariant := createTestProductVariant(db, product.ID)
	review := createTestReview(t, db, user.ID, productVariant.ID, 5, "Great product!", "Excellent quality")

	t.Run("Success - Get existing review", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}

		handler.GetReview(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		assert.Equal(t, float64(review.ID), data["id"])
		assert.Equal(t, float64(productVariant.ID), data["product_variant_id"])
		assert.Equal(t, float64(5), data["rating"])
		assert.Equal(t, "Great product!", data["title"])
		assert.Equal(t, "Excellent quality", data["content"])
		assert.True(t, data["is_verified_purchase"].(bool))
		assert.Equal(t, float64(0), data["helpful_count"])

		// Check user data
		userData := data["user"].(map[string]interface{})
		assert.Equal(t, float64(user.ID), userData["id"])
		// Note: createTestUser doesn't set first/last names, so they'll be empty
		assert.Equal(t, "", userData["first_name"])
		assert.Equal(t, "", userData["last_name"])
		assert.Equal(t, "Anonymous", userData["name"]) // Default name when no first/last name
		assert.Equal(t, user.Email, userData["email"])
		assert.Equal(t, user.Phone, userData["phone"])
		assert.Equal(t, user.Avatar, userData["avatar"])

		// Check that moderation history is NOT included for regular users
		assert.NotContains(t, data, "moderation_history")
	})

	t.Run("Error - Invalid review ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.GetReview(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, float64(http.StatusBadRequest), response["status"])
		assert.Equal(t, "INVALID_REVIEW_ID", response["error"].(map[string]interface{})["code"])
	})

	t.Run("Error - Review not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "99999"}}

		handler.GetReview(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, float64(http.StatusNotFound), response["status"])
		assert.Equal(t, "REVIEW_NOT_FOUND", response["error"].(map[string]interface{})["code"])
	})

	t.Run("Error - Pending review not accessible for regular users", func(t *testing.T) {
		// Create a pending review
		pendingReview := createTestReview(t, db, user.ID, productVariant.ID, 3, "Pending", "Pending content")
		pendingReview.Status = models.ReviewStatusPending
		db.Save(&pendingReview)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(pendingReview.ID), 10)}}

		handler.GetReview(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Success - Admin can access pending review", func(t *testing.T) {
		// Create a pending review
		pendingReview := createTestReview(t, db, user.ID, productVariant.ID, 3, "Pending", "Pending content")
		pendingReview.Status = models.ReviewStatusPending
		db.Save(&pendingReview)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(pendingReview.ID), 10)}}

		// Set admin user context
		c.Set("user_type", models.Admin)

		handler.GetReview(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		assert.Equal(t, float64(pendingReview.ID), data["id"])
		assert.Equal(t, "Pending", data["title"])
		assert.Equal(t, "Pending content", data["content"])

		// Check user data includes new fields
		userData := data["user"].(map[string]interface{})
		assert.Equal(t, user.Email, userData["email"])
		assert.Equal(t, user.Phone, userData["phone"])
		assert.Equal(t, user.Avatar, userData["avatar"])

		// Check that moderation history is included for admin
		assert.Contains(t, data, "moderation_history")
		moderationHistory, exists := data["moderation_history"]
		if exists && moderationHistory != nil {
			historyArray := moderationHistory.([]interface{})
			assert.Len(t, historyArray, 0) // No moderation history for new review
		}
	})
}

func TestGetProductReviews(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDBWithReviewTables(t)
	handler := NewReviewHandler(db, nil)

	// Create test data
	user1 := createTestUser(db, models.Customer)
	user2 := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	productVariant := createTestProductVariant(db, product.ID)

	// Create multiple reviews
	_ = createTestReview(t, db, user1.ID, productVariant.ID, 5, "Excellent", "Great product")
	_ = createTestReview(t, db, user2.ID, productVariant.ID, 4, "Good", "Nice product")
	_ = createTestReview(t, db, user1.ID, productVariant.ID, 3, "Average", "Okay product")
	_ = createTestReview(t, db, user2.ID, productVariant.ID, 2, "Poor", "Not great")
	review5 := createTestReview(t, db, user1.ID, productVariant.ID, 1, "Terrible", "Bad product")

	// Create a pending review (should not appear in results)
	pendingReview := createTestReview(t, db, user1.ID, productVariant.ID, 5, "Pending", "Pending review")
	pendingReview.Status = models.ReviewStatusPending
	db.Save(&pendingReview)

	t.Run("Success - Get all reviews with default pagination", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "productVariantId", Value: strconv.FormatUint(uint64(productVariant.ID), 10)}}

		handler.GetProductReviews(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})

		// Check pagination
		pagination := data["pagination"].(map[string]interface{})
		assert.Equal(t, float64(1), pagination["page"])
		assert.Equal(t, float64(10), pagination["limit"])
		assert.Equal(t, float64(5), pagination["total"]) // 5 approved reviews
		assert.Equal(t, float64(1), pagination["total_pages"])
		assert.False(t, pagination["has_next"].(bool))
		assert.False(t, pagination["has_prev"].(bool))

		// Check reviews
		reviews := data["reviews"].([]interface{})
		assert.Len(t, reviews, 5)

		// Check first review (should be most recent due to default sort)
		firstReview := reviews[0].(map[string]interface{})
		assert.Equal(t, float64(review5.ID), firstReview["id"])
		assert.Equal(t, float64(1), firstReview["rating"])
		assert.Equal(t, "Terrible", firstReview["title"])
	})

	t.Run("Success - Pagination with limit 2", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "productVariantId", Value: strconv.FormatUint(uint64(productVariant.ID), 10)}}
		c.Request = httptest.NewRequest("GET", "/?limit=2&page=1", nil)

		handler.GetProductReviews(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response["data"].(map[string]interface{})
		pagination := data["pagination"].(map[string]interface{})
		assert.Equal(t, float64(1), pagination["page"])
		assert.Equal(t, float64(2), pagination["limit"])
		assert.Equal(t, float64(5), pagination["total"])
		assert.Equal(t, float64(3), pagination["total_pages"]) // ceil(5/2) = 3
		assert.True(t, pagination["has_next"].(bool))
		assert.False(t, pagination["has_prev"].(bool))

		reviews := data["reviews"].([]interface{})
		assert.Len(t, reviews, 2)
	})

	t.Run("Success - Filter by rating", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "productVariantId", Value: strconv.FormatUint(uint64(productVariant.ID), 10)}}
		c.Request = httptest.NewRequest("GET", "/?rating=5", nil)

		handler.GetProductReviews(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response["data"].(map[string]interface{})
		pagination := data["pagination"].(map[string]interface{})
		assert.Equal(t, float64(1), pagination["total"]) // Only 1 review with rating 5

		reviews := data["reviews"].([]interface{})
		assert.Len(t, reviews, 1)
		assert.Equal(t, float64(5), reviews[0].(map[string]interface{})["rating"])
	})

	t.Run("Success - Sort by rating ascending", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "productVariantId", Value: strconv.FormatUint(uint64(productVariant.ID), 10)}}
		c.Request = httptest.NewRequest("GET", "/?sort=rating&order=asc", nil)

		handler.GetProductReviews(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response["data"].(map[string]interface{})
		reviews := data["reviews"].([]interface{})

		// Should be sorted by rating ascending (1, 2, 3, 4, 5)
		assert.Equal(t, float64(1), reviews[0].(map[string]interface{})["rating"])
		assert.Equal(t, float64(5), reviews[4].(map[string]interface{})["rating"])
	})

	t.Run("Error - Invalid product variant ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "productVariantId", Value: "invalid"}}

		handler.GetProductReviews(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, float64(http.StatusBadRequest), response["status"])
		assert.Equal(t, "INVALID_PRODUCT_VARIANT_ID", response["error"].(map[string]interface{})["code"])
	})

	t.Run("Error - Invalid rating filter", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "productVariantId", Value: strconv.FormatUint(uint64(productVariant.ID), 10)}}
		c.Request = httptest.NewRequest("GET", "/?rating=6", nil)

		handler.GetProductReviews(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, float64(http.StatusBadRequest), response["status"])
		assert.Equal(t, "INVALID_RATING_FILTER", response["error"].(map[string]interface{})["code"])
	})

	t.Run("Success - No reviews for product", func(t *testing.T) {
		// Create a new product variant with no reviews
		emptyProduct := createTestProduct(db)
		emptyProductVariant := createTestProductVariant(db, emptyProduct.ID)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "productVariantId", Value: strconv.FormatUint(uint64(emptyProductVariant.ID), 10)}}

		handler.GetProductReviews(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response["data"].(map[string]interface{})
		pagination := data["pagination"].(map[string]interface{})
		assert.Equal(t, float64(0), pagination["total"])
		assert.Equal(t, float64(0), pagination["total_pages"])

		reviews, ok := data["reviews"].([]interface{})
		if ok {
			assert.Len(t, reviews, 0)
		} else {
			assert.Nil(t, data["reviews"])
		}
	})

	t.Run("Success - Include seller response", func(t *testing.T) {
		// Create a review with seller response
		reviewWithResponse := createTestReview(t, db, user1.ID, productVariant.ID, 4, "With Response", "Good product")
		sellerResponse := models.SellerResponse{
			ProductReviewID: reviewWithResponse.ID,
			UserID:          user2.ID, // user2 as seller
			Content:         "Thank you for your review!",
		}
		db.Create(&sellerResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "productVariantId", Value: strconv.FormatUint(uint64(productVariant.ID), 10)}}

		handler.GetProductReviews(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response["data"].(map[string]interface{})
		reviews := data["reviews"].([]interface{})

		// Find the review with seller response
		var reviewWithSellerResponse map[string]interface{}
		for _, r := range reviews {
			review := r.(map[string]interface{})
			if review["title"] == "With Response" {
				reviewWithSellerResponse = review
				break
			}
		}

		assert.NotNil(t, reviewWithSellerResponse)
		assert.Contains(t, reviewWithSellerResponse, "seller_response")

		sellerResp := reviewWithSellerResponse["seller_response"].(map[string]interface{})
		assert.Equal(t, "Thank you for your review!", sellerResp["content"])
		assert.Equal(t, float64(user2.ID), sellerResp["user"].(map[string]interface{})["id"])
	})
}

// Helper function for creating test reviews
func createTestReview(t *testing.T, db *gorm.DB, userID, productVariantID uint, rating int, title, content string) *models.ProductReview {
	review := &models.ProductReview{
		ProductVariantID:   productVariantID,
		UserID:             userID,
		Rating:             rating,
		Title:              title,
		Content:            content,
		IsVerifiedPurchase: true,
		Status:             models.ReviewStatusApproved,
		HelpfulCount:       0,
	}

	err := db.Create(review).Error
	assert.NoError(t, err)

	return review
}
