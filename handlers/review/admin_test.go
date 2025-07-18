package review

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetAllReviews(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDBWithReviewTables(t)
	handler := NewReviewHandler(db, nil)

	admin := createTestUser(db, models.Admin)
	customer1 := createTestUser(db, models.Customer)
	customer2 := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	productVariant := createTestProductVariant(db, product.ID)

	// Create reviews with different statuses
	review1 := createTestReview(t, db, customer1.ID, productVariant.ID, 5, "Great!", "Excellent product")
	review2 := createTestReview(t, db, customer2.ID, productVariant.ID, 3, "Okay", "Average product")
	review3 := createTestReview(t, db, customer1.ID, productVariant.ID, 1, "Bad", "Poor product")

	// Set different statuses
	db.Model(&review1).Update("status", models.ReviewStatusApproved)
	db.Model(&review2).Update("status", models.ReviewStatusPending)
	db.Model(&review3).Update("status", models.ReviewStatusRejected)

	t.Run("Success - Get all reviews", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/api/v1/admin/reviews", nil)
		c.Request = req
		c.Set("user_id", admin.ID)

		handler.GetAllReviews(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		data := response["data"].(map[string]interface{})
		reviews := data["reviews"].([]interface{})
		pagination := data["pagination"].(map[string]interface{})

		// Should return all 3 reviews
		assert.Len(t, reviews, 3)
		assert.Equal(t, float64(3), pagination["total"])
	})

	t.Run("Success - Filter by status", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/api/v1/admin/reviews?status=APPROVED", nil)
		c.Request = req
		c.Set("user_id", admin.ID)

		handler.GetAllReviews(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		data := response["data"].(map[string]interface{})
		reviews := data["reviews"].([]interface{})
		pagination := data["pagination"].(map[string]interface{})

		// Should return only approved reviews
		assert.Len(t, reviews, 1)
		assert.Equal(t, float64(1), pagination["total"])
	})

	t.Run("Success - Filter by rating", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/api/v1/admin/reviews?rating=5", nil)
		c.Request = req
		c.Set("user_id", admin.ID)

		handler.GetAllReviews(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		data := response["data"].(map[string]interface{})
		reviews := data["reviews"].([]interface{})
		pagination := data["pagination"].(map[string]interface{})

		// Should return only 5-star reviews
		assert.Len(t, reviews, 1)
		assert.Equal(t, float64(1), pagination["total"])
	})

	t.Run("Success - Filter by user", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/api/v1/admin/reviews?user_id="+strconv.FormatUint(uint64(customer1.ID), 10), nil)
		c.Request = req
		c.Set("user_id", admin.ID)

		handler.GetAllReviews(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		data := response["data"].(map[string]interface{})
		reviews := data["reviews"].([]interface{})
		pagination := data["pagination"].(map[string]interface{})

		// Should return only customer1's reviews
		assert.Len(t, reviews, 2)
		assert.Equal(t, float64(2), pagination["total"])
	})
}

func TestModerateReview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDBWithReviewTables(t)
	handler := NewReviewHandler(db, nil)

	admin := createTestUser(db, models.Admin)
	customer := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	productVariant := createTestProductVariant(db, product.ID)
	review := createTestReview(t, db, customer.ID, productVariant.ID, 5, "Great!", "Nice product")

	t.Run("Success - Approve review", func(t *testing.T) {
		requestBody := ModerationRequest{
			Status: models.ReviewStatusApproved,
			Reason: "Review meets community guidelines",
		}
		body, _ := json.Marshal(requestBody)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("PUT", "/api/v1/admin/reviews/"+strconv.FormatUint(uint64(review.ID), 10)+"/moderate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", admin.ID)

		handler.ModerateReview(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Review moderated successfully", response["message"])

		// Verify the review was updated
		var updatedReview models.ProductReview
		db.First(&updatedReview, review.ID)
		assert.Equal(t, models.ReviewStatusApproved, updatedReview.Status)
		assert.Equal(t, "Review meets community guidelines", updatedReview.ModerationReason)
		assert.NotNil(t, updatedReview.ModeratedBy)
		assert.NotNil(t, updatedReview.ModeratedAt)
	})

	t.Run("Success - Reject review", func(t *testing.T) {
		// Create a new review for rejection
		reviewToReject := createTestReview(t, db, customer.ID, productVariant.ID, 1, "Bad", "Poor product")

		requestBody := ModerationRequest{
			Status: models.ReviewStatusRejected,
			Reason: "Review violates community guidelines",
		}
		body, _ := json.Marshal(requestBody)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("PUT", "/api/v1/admin/reviews/"+strconv.FormatUint(uint64(reviewToReject.ID), 10)+"/moderate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(reviewToReject.ID), 10)}}
		c.Set("user_id", admin.ID)

		handler.ModerateReview(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		// Verify the review was rejected
		var updatedReview models.ProductReview
		db.First(&updatedReview, reviewToReject.ID)
		assert.Equal(t, models.ReviewStatusRejected, updatedReview.Status)
	})

	t.Run("Error - Invalid status", func(t *testing.T) {
		requestBody := ModerationRequest{
			Status: "INVALID_STATUS",
			Reason: "Invalid status test",
		}
		body, _ := json.Marshal(requestBody)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("PUT", "/api/v1/admin/reviews/"+strconv.FormatUint(uint64(review.ID), 10)+"/moderate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", admin.ID)

		handler.ModerateReview(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INVALID_STATUS", response["error"].(map[string]interface{})["code"])
	})

	t.Run("Error - Review not found", func(t *testing.T) {
		requestBody := ModerationRequest{
			Status: models.ReviewStatusApproved,
			Reason: "Test reason",
		}
		body, _ := json.Marshal(requestBody)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("PUT", "/api/v1/admin/reviews/999/moderate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "999"}}
		c.Set("user_id", admin.ID)

		handler.ModerateReview(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "REVIEW_NOT_FOUND", response["error"].(map[string]interface{})["code"])
	})
}

func TestAdminDeleteReview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDBWithReviewTables(t)
	handler := NewReviewHandler(db, nil)

	admin := createTestUser(db, models.Admin)
	customer := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	productVariant := createTestProductVariant(db, product.ID)
	review := createTestReview(t, db, customer.ID, productVariant.ID, 5, "Great!", "Nice product")

	t.Run("Success - Delete review", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("DELETE", "/api/v1/admin/reviews/"+strconv.FormatUint(uint64(review.ID), 10), nil)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", admin.ID)

		handler.AdminDeleteReview(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Review permanently deleted", response["message"])

		// Verify the review was permanently deleted
		var deletedReview models.ProductReview
		err = db.Unscoped().First(&deletedReview, review.ID).Error
		assert.Error(t, err) // Should not find the review
	})

	t.Run("Error - Review not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("DELETE", "/api/v1/admin/reviews/999", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "999"}}
		c.Set("user_id", admin.ID)

		handler.AdminDeleteReview(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "REVIEW_NOT_FOUND", response["error"].(map[string]interface{})["code"])
	})
}

func TestGetModerationStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDBWithReviewTables(t)
	handler := NewReviewHandler(db, nil)

	admin := createTestUser(db, models.Admin)
	customer1 := createTestUser(db, models.Customer)
	customer2 := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	productVariant := createTestProductVariant(db, product.ID)

	// Create reviews with different statuses
	review1 := createTestReview(t, db, customer1.ID, productVariant.ID, 5, "Great!", "Excellent product")
	review2 := createTestReview(t, db, customer2.ID, productVariant.ID, 3, "Okay", "Average product")
	review3 := createTestReview(t, db, customer1.ID, productVariant.ID, 1, "Bad", "Poor product")

	// Set different statuses
	db.Model(&review1).Update("status", models.ReviewStatusApproved)
	db.Model(&review2).Update("status", models.ReviewStatusPending)
	db.Model(&review3).Update("status", models.ReviewStatusRejected)

	t.Run("Success - Get moderation stats", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/api/v1/admin/reviews/stats", nil)
		c.Request = req
		c.Set("user_id", admin.ID)

		handler.GetModerationStats(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		data := response["data"].(map[string]interface{})
		stats := data["stats"].(map[string]interface{})

		// Should have correct counts
		assert.Equal(t, float64(3), stats["total"])
		assert.Equal(t, float64(1), stats["pending"])
		assert.Equal(t, float64(1), stats["approved"])
		assert.Equal(t, float64(1), stats["rejected"])
		assert.Equal(t, float64(0), stats["flagged"])
		assert.Equal(t, float64(0), stats["deleted"])
	})
}
