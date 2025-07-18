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
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMarkReviewHelpful(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDBWithReviewTables(t)
	handler := NewReviewHandler(db, nil)

	// Create test data
	reviewer := createTestUser(db, models.Customer)
	voter := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	productVariant := createTestProductVariant(db, product.ID)
	review := createTestReview(t, db, reviewer.ID, productVariant.ID, 5, "Great product!", "Excellent quality")

	t.Run("Success - Mark review as helpful", func(t *testing.T) {
		requestBody := MarkReviewHelpfulRequest{
			IsHelpful: true,
		}
		body, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/reviews/"+strconv.FormatUint(uint64(review.ID), 10)+"/helpful", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", voter.ID)

		handler.MarkReviewHelpful(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Vote recorded successfully", response["message"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, float64(review.ID), data["review_id"])
		assert.True(t, data["is_helpful"].(bool))
		assert.Equal(t, float64(1), data["helpful_count"])

		// Verify vote was created in database
		var vote models.ReviewHelpful
		err = db.Where("product_review_id = ? AND user_id = ?", review.ID, voter.ID).First(&vote).Error
		require.NoError(t, err)
		assert.True(t, vote.IsHelpful)

		// Verify review helpful count was updated
		var updatedReview models.ProductReview
		err = db.First(&updatedReview, review.ID).Error
		require.NoError(t, err)
		assert.Equal(t, 1, updatedReview.HelpfulCount)
	})

	t.Run("Success - Mark review as unhelpful", func(t *testing.T) {
		// Create a fresh review and a new voter for this test
		freshReview := createTestReview(t, db, reviewer.ID, productVariant.ID, 4, "Another review", "Another content")
		newVoter := createTestUser(db, models.Customer)

		requestBody := MarkReviewHelpfulRequest{
			IsHelpful: false,
		}
		body, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("POST", "/api/v1/reviews/"+strconv.FormatUint(uint64(freshReview.ID), 10)+"/helpful", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(freshReview.ID), 10)}}
		c.Set("user_id", newVoter.ID)

		handler.MarkReviewHelpful(c)

		if w.Code != http.StatusOK {
			t.Logf("Response body: %s", w.Body.String())
		}
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify vote was created in database
		var vote models.ReviewHelpful
		err := db.Where("product_review_id = ? AND user_id = ?", freshReview.ID, newVoter.ID).First(&vote).Error
		require.NoError(t, err)
		assert.False(t, vote.IsHelpful)

		// Verify review helpful count was updated (should be 0 for unhelpful votes)
		var updatedReview models.ProductReview
		err = db.First(&updatedReview, freshReview.ID).Error
		require.NoError(t, err)
		assert.Equal(t, 0, updatedReview.HelpfulCount)
	})

	t.Run("Success - Remove vote by voting same way again", func(t *testing.T) {
		// Create a fresh review for this test
		freshReview := createTestReview(t, db, reviewer.ID, productVariant.ID, 3, "Third review", "Third content")

		// First, mark as helpful
		requestBody := MarkReviewHelpfulRequest{
			IsHelpful: true,
		}
		body, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/reviews/"+strconv.FormatUint(uint64(freshReview.ID), 10)+"/helpful", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(freshReview.ID), 10)}}
		c.Set("user_id", voter.ID)

		handler.MarkReviewHelpful(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// Now vote helpful again (should remove the vote)
		req2 := httptest.NewRequest("POST", "/api/v1/reviews/"+strconv.FormatUint(uint64(freshReview.ID), 10)+"/helpful", bytes.NewBuffer(body))
		req2.Header.Set("Content-Type", "application/json")
		c.Request = req2
		handler.MarkReviewHelpful(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify vote was removed from database
		var vote models.ReviewHelpful
		err := db.Where("product_review_id = ? AND user_id = ?", freshReview.ID, voter.ID).First(&vote).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)

		// Verify review helpful count was updated (should be 0)
		var updatedReview models.ProductReview
		err = db.First(&updatedReview, freshReview.ID).Error
		require.NoError(t, err)
		assert.Equal(t, 0, updatedReview.HelpfulCount)
	})

	t.Run("Error - Invalid review ID", func(t *testing.T) {
		requestBody := MarkReviewHelpfulRequest{
			IsHelpful: true,
		}
		body, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/reviews/invalid/helpful", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}
		c.Set("user_id", voter.ID)

		handler.MarkReviewHelpful(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(http.StatusBadRequest), response["status"])
		assert.Equal(t, "INVALID_REVIEW_ID", response["error"].(map[string]interface{})["code"])
	})

	t.Run("Error - Review not found", func(t *testing.T) {
		requestBody := MarkReviewHelpfulRequest{
			IsHelpful: true,
		}
		body, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/reviews/99999/helpful", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "99999"}}
		c.Set("user_id", voter.ID)

		handler.MarkReviewHelpful(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(http.StatusNotFound), response["status"])
		assert.Equal(t, "REVIEW_NOT_FOUND", response["error"].(map[string]interface{})["code"])
	})

	t.Run("Error - Pending review not accessible", func(t *testing.T) {
		// Create a pending review
		pendingReview := createTestReview(t, db, reviewer.ID, productVariant.ID, 3, "Pending", "Pending content")
		pendingReview.Status = models.ReviewStatusPending
		db.Save(&pendingReview)

		requestBody := MarkReviewHelpfulRequest{
			IsHelpful: true,
		}
		body, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/reviews/"+strconv.FormatUint(uint64(pendingReview.ID), 10)+"/helpful", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(pendingReview.ID), 10)}}
		c.Set("user_id", voter.ID)

		handler.MarkReviewHelpful(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Error - User cannot vote on own review", func(t *testing.T) {
		requestBody := MarkReviewHelpfulRequest{
			IsHelpful: true,
		}
		body, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/reviews/"+strconv.FormatUint(uint64(review.ID), 10)+"/helpful", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", reviewer.ID) // Reviewer trying to vote on their own review

		handler.MarkReviewHelpful(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(http.StatusBadRequest), response["status"])
		assert.Equal(t, "SELF_VOTE_NOT_ALLOWED", response["error"].(map[string]interface{})["code"])
	})

	t.Run("Error - Invalid request body", func(t *testing.T) {
		// Send invalid JSON
		body := []byte(`{"invalid": "json"`)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/reviews/"+strconv.FormatUint(uint64(review.ID), 10)+"/helpful", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", voter.ID)

		handler.MarkReviewHelpful(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(http.StatusBadRequest), response["status"])
		assert.Equal(t, "INVALID_REQUEST", response["error"].(map[string]interface{})["code"])
	})

	t.Run("Error - User not authenticated", func(t *testing.T) {
		requestBody := MarkReviewHelpfulRequest{
			IsHelpful: true,
		}
		body, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/reviews/"+strconv.FormatUint(uint64(review.ID), 10)+"/helpful", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		// No user_id set in context

		handler.MarkReviewHelpful(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(http.StatusUnauthorized), response["status"])
		assert.Equal(t, "UNAUTHORIZED", response["error"].(map[string]interface{})["code"])
	})
}

func TestGetUserVoteStatus(t *testing.T) {
	// Setup
	db := setupTestDBWithReviewTables(t)
	handler := NewReviewHandler(db, nil)

	// Create test data
	user := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	productVariant := createTestProductVariant(db, product.ID)
	review := createTestReview(t, db, user.ID, productVariant.ID, 5, "Test", "Test content")

	t.Run("No vote found", func(t *testing.T) {
		vote, err := handler.GetUserVoteStatus(user.ID, review.ID)
		assert.NoError(t, err)
		assert.Nil(t, vote)
	})

	t.Run("Vote found", func(t *testing.T) {
		// Create a vote
		vote := models.ReviewHelpful{
			ProductReviewID: review.ID,
			UserID:          user.ID,
			IsHelpful:       true,
		}
		err := db.Create(&vote).Error
		require.NoError(t, err)

		// Get vote status
		retrievedVote, err := handler.GetUserVoteStatus(user.ID, review.ID)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedVote)
		assert.Equal(t, review.ID, retrievedVote.ProductReviewID)
		assert.Equal(t, user.ID, retrievedVote.UserID)
		assert.True(t, retrievedVote.IsHelpful)
	})
}

func TestUpdateReviewHelpfulCount(t *testing.T) {
	// Setup
	db := setupTestDBWithReviewTables(t)
	handler := NewReviewHandler(db, nil)

	// Create test data
	user1 := createTestUser(db, models.Customer)
	user2 := createTestUser(db, models.Customer)
	user3 := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	productVariant := createTestProductVariant(db, product.ID)
	review := createTestReview(t, db, user1.ID, productVariant.ID, 5, "Test", "Test content")

	t.Run("Update helpful count with multiple votes", func(t *testing.T) {
		// Create helpful votes
		helpfulVote1 := models.ReviewHelpful{
			ProductReviewID: review.ID,
			UserID:          user1.ID,
			IsHelpful:       true,
		}
		helpfulVote2 := models.ReviewHelpful{
			ProductReviewID: review.ID,
			UserID:          user2.ID,
			IsHelpful:       true,
		}
		// Create unhelpful vote
		unhelpfulVote := models.ReviewHelpful{
			ProductReviewID: review.ID,
			UserID:          user3.ID,
			IsHelpful:       false,
		}

		err := db.Create(&helpfulVote1).Error
		require.NoError(t, err)
		err = db.Create(&helpfulVote2).Error
		require.NoError(t, err)
		err = db.Create(&unhelpfulVote).Error
		require.NoError(t, err)

		// Update helpful count
		err = handler.UpdateReviewHelpfulCount(review.ID)
		assert.NoError(t, err)

		// Verify count was updated
		var updatedReview models.ProductReview
		err = db.First(&updatedReview, review.ID).Error
		require.NoError(t, err)
		assert.Equal(t, 2, updatedReview.HelpfulCount) // Only helpful votes count
	})
}
