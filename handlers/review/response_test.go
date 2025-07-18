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

func TestCreateSellerResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDBWithReviewTables(t)
	handler := NewReviewHandler(db, nil)

	seller := createTestUser(db, models.Vendor)
	otherSeller := createTestUser(db, models.Vendor)
	customer := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	productVariant := createTestProductVariant(db, product.ID)
	review := createTestReview(t, db, customer.ID, productVariant.ID, 5, "Great!", "Nice product")

	t.Run("Success - Create seller response", func(t *testing.T) {
		requestBody := SellerResponseRequest{Content: "Thank you for your feedback!"}
		body, _ := json.Marshal(requestBody)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("POST", "/api/v1/reviews/"+strconv.FormatUint(uint64(review.ID), 10)+"/response", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", seller.ID)

		handler.CreateSellerResponse(c)
		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Seller response created successfully", response["message"])
	})

	t.Run("Error - Duplicate response", func(t *testing.T) {
		requestBody := SellerResponseRequest{Content: "Duplicate response"}
		body, _ := json.Marshal(requestBody)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("POST", "/api/v1/reviews/"+strconv.FormatUint(uint64(review.ID), 10)+"/response", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", seller.ID)

		handler.CreateSellerResponse(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "RESPONSE_EXISTS", response["error"].(map[string]interface{})["code"])
	})

	t.Run("Success - Other seller can also respond", func(t *testing.T) {
		// Create a different review for the other seller to respond to
		otherProduct := createTestProduct(db)
		otherProductVariant := createTestProductVariant(db, otherProduct.ID)
		otherReview := createTestReview(t, db, customer.ID, otherProductVariant.ID, 4, "Good product", "Nice quality")

		requestBody := SellerResponseRequest{Content: "Thank you from other seller!"}
		body, _ := json.Marshal(requestBody)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("POST", "/api/v1/reviews/"+strconv.FormatUint(uint64(otherReview.ID), 10)+"/response", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(otherReview.ID), 10)}}
		c.Set("user_id", otherSeller.ID)

		handler.CreateSellerResponse(c)
		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Seller response created successfully", response["message"])
	})

	t.Run("Error - Content too long", func(t *testing.T) {
		longContent := make([]byte, 501)
		for i := range longContent {
			longContent[i] = 'a'
		}
		requestBody := SellerResponseRequest{Content: string(longContent)}
		body, _ := json.Marshal(requestBody)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("POST", "/api/v1/reviews/"+strconv.FormatUint(uint64(review.ID), 10)+"/response", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", seller.ID)

		handler.CreateSellerResponse(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INVALID_REQUEST", response["error"].(map[string]interface{})["code"])
	})
}

func TestUpdateSellerResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDBWithReviewTables(t)
	handler := NewReviewHandler(db, nil)

	seller := createTestUser(db, models.Vendor)
	customer := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	productVariant := createTestProductVariant(db, product.ID)
	review := createTestReview(t, db, customer.ID, productVariant.ID, 5, "Great!", "Nice product")

	// Create initial response
	initialResponse := models.SellerResponse{
		ProductReviewID: review.ID,
		UserID:          seller.ID,
		Content:         "Initial response",
	}
	db.Create(&initialResponse)

	t.Run("Success - Update seller response", func(t *testing.T) {
		requestBody := SellerResponseRequest{Content: "Updated response"}
		body, _ := json.Marshal(requestBody)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("PUT", "/api/v1/reviews/"+strconv.FormatUint(uint64(review.ID), 10)+"/response", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", seller.ID)

		handler.UpdateSellerResponse(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Seller response updated successfully", response["message"])
	})

	t.Run("Error - No existing response to update", func(t *testing.T) {
		otherSeller := createTestUser(db, models.Vendor)
		requestBody := SellerResponseRequest{Content: "No response"}
		body, _ := json.Marshal(requestBody)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("PUT", "/api/v1/reviews/"+strconv.FormatUint(uint64(review.ID), 10)+"/response", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", otherSeller.ID)

		handler.UpdateSellerResponse(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "RESPONSE_NOT_FOUND", response["error"].(map[string]interface{})["code"])
	})

	t.Run("Error - Content too long", func(t *testing.T) {
		longContent := make([]byte, 501)
		for i := range longContent {
			longContent[i] = 'a'
		}
		requestBody := SellerResponseRequest{Content: string(longContent)}
		body, _ := json.Marshal(requestBody)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("PUT", "/api/v1/reviews/"+strconv.FormatUint(uint64(review.ID), 10)+"/response", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(review.ID), 10)}}
		c.Set("user_id", seller.ID)

		handler.UpdateSellerResponse(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INVALID_REQUEST", response["error"].(map[string]interface{})["code"])
	})
}
