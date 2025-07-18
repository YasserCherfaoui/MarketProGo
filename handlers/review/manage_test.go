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

func TestUpdateReview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDBWithReviewTables(t)
	handler := NewReviewHandler(db, nil)

	customer := createTestUser(db, models.Customer)
	otherCustomer := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	productVariant := createTestProductVariant(db, product.ID)
	review := createTestReview(t, db, customer.ID, productVariant.ID, 5, "Great!", "Nice product")

	t.Run("Success - Update review", func(t *testing.T) {
		requestBody := UpdateReviewRequest{
			Rating:  4,
			Title:   "Updated Title",
			Content: "Updated content with more details about the product experience.",
		}
		body, _ := json.Marshal(requestBody)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("PUT", "/api/v1/reviews/"+strconv.FormatUint(uint64(review.ID), 10), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", customer.ID)

		handler.UpdateReview(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Review updated successfully", response["message"])

		// Verify the review was actually updated
		var updatedReview models.ProductReview
		db.First(&updatedReview, review.ID)
		assert.Equal(t, 4, updatedReview.Rating)
		assert.Equal(t, "Updated Title", updatedReview.Title)
		assert.Equal(t, "Updated content with more details about the product experience.", updatedReview.Content)
	})

	t.Run("Error - Not review owner", func(t *testing.T) {
		requestBody := UpdateReviewRequest{
			Rating:  3,
			Title:   "Not my review",
			Content: "Trying to update someone else's review.",
		}
		body, _ := json.Marshal(requestBody)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("PUT", "/api/v1/reviews/"+strconv.FormatUint(uint64(review.ID), 10), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", otherCustomer.ID)

		handler.UpdateReview(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "REVIEW_NOT_FOUND", response["error"].(map[string]interface{})["code"])
	})

	t.Run("Error - Invalid rating", func(t *testing.T) {
		requestBody := UpdateReviewRequest{
			Rating:  6, // Invalid rating
			Title:   "Invalid Rating",
			Content: "Content with invalid rating.",
		}
		body, _ := json.Marshal(requestBody)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("PUT", "/api/v1/reviews/"+strconv.FormatUint(uint64(review.ID), 10), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", customer.ID)

		handler.UpdateReview(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INVALID_REQUEST", response["error"].(map[string]interface{})["code"])
	})

	t.Run("Error - Content too short", func(t *testing.T) {
		requestBody := UpdateReviewRequest{
			Rating:  4,
			Title:   "Short Content",
			Content: "Short", // Too short
		}
		body, _ := json.Marshal(requestBody)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("PUT", "/api/v1/reviews/"+strconv.FormatUint(uint64(review.ID), 10), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", customer.ID)

		handler.UpdateReview(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INVALID_REQUEST", response["error"].(map[string]interface{})["code"])
	})
}

func TestDeleteReview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDBWithReviewTables(t)
	handler := NewReviewHandler(db, nil)

	customer := createTestUser(db, models.Customer)
	otherCustomer := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	productVariant := createTestProductVariant(db, product.ID)
	review := createTestReview(t, db, customer.ID, productVariant.ID, 5, "Great!", "Nice product")

	t.Run("Success - Delete review", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("DELETE", "/api/v1/reviews/"+strconv.FormatUint(uint64(review.ID), 10), nil)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", customer.ID)

		handler.DeleteReview(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Review deleted successfully", response["message"])

		// Verify the review was soft deleted
		var deletedReview models.ProductReview
		err = db.Unscoped().First(&deletedReview, review.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedReview.DeletedAt)
	})

	t.Run("Error - Not review owner", func(t *testing.T) {
		// Create a new review for the other customer
		otherReview := createTestReview(t, db, otherCustomer.ID, productVariant.ID, 4, "Other Review", "Other content")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("DELETE", "/api/v1/reviews/"+strconv.FormatUint(uint64(otherReview.ID), 10), nil)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(otherReview.ID), 10)}}
		c.Set("user_id", customer.ID)

		handler.DeleteReview(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "REVIEW_NOT_FOUND", response["error"].(map[string]interface{})["code"])
	})

	t.Run("Error - Review not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("DELETE", "/api/v1/reviews/999", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "999"}}
		c.Set("user_id", customer.ID)

		handler.DeleteReview(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "REVIEW_NOT_FOUND", response["error"].(map[string]interface{})["code"])
	})
}

func TestGetUserReviews(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDBWithReviewTables(t)
	handler := NewReviewHandler(db, nil)

	customer := createTestUser(db, models.Customer)
	otherCustomer := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	productVariant := createTestProductVariant(db, product.ID)

	// Create multiple reviews for the customer
	createTestReview(t, db, customer.ID, productVariant.ID, 5, "First Review", "First review content")
	createTestReview(t, db, customer.ID, productVariant.ID, 4, "Second Review", "Second review content")
	createTestReview(t, db, customer.ID, productVariant.ID, 3, "Third Review", "Third review content")

	// Create a review for another customer
	createTestReview(t, db, otherCustomer.ID, productVariant.ID, 5, "Other Review", "Other review content")

	t.Run("Success - Get user reviews", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/api/v1/reviews/user/me", nil)
		c.Request = req
		c.Set("user_id", customer.ID)

		handler.GetUserReviews(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		data := response["data"].(map[string]interface{})
		reviews := data["reviews"].([]interface{})
		pagination := data["pagination"].(map[string]interface{})

		// Should return 3 reviews for the customer
		assert.Len(t, reviews, 3)
		assert.Equal(t, float64(3), pagination["total"])
		assert.Equal(t, float64(1), pagination["page"])
		assert.Equal(t, float64(10), pagination["limit"])
	})

	t.Run("Success - Get user reviews with pagination", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/api/v1/reviews/user/me?page=1&limit=2", nil)
		c.Request = req
		c.Set("user_id", customer.ID)

		handler.GetUserReviews(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		data := response["data"].(map[string]interface{})
		reviews := data["reviews"].([]interface{})
		pagination := data["pagination"].(map[string]interface{})

		// Should return 2 reviews due to limit
		assert.Len(t, reviews, 2)
		assert.Equal(t, float64(3), pagination["total"])
		assert.Equal(t, float64(2), pagination["totalPages"])
		assert.True(t, pagination["hasNext"].(bool))
		assert.False(t, pagination["hasPrev"].(bool))
	})

	t.Run("Success - Get user reviews with status filter", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/api/v1/reviews/user/me?status=APPROVED", nil)
		c.Request = req
		c.Set("user_id", customer.ID)

		handler.GetUserReviews(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		data := response["data"].(map[string]interface{})
		reviews := data["reviews"].([]interface{})
		pagination := data["pagination"].(map[string]interface{})

		// Should return 3 approved reviews
		assert.Len(t, reviews, 3)
		assert.Equal(t, float64(3), pagination["total"])
	})

	t.Run("Error - Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/api/v1/reviews/user/me", nil)
		c.Request = req
		// No user_id set

		handler.GetUserReviews(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", response["error"].(map[string]interface{})["code"])
	})
}
