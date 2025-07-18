package review

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// createTestPurchaseData creates test data for purchase verification
func createTestPurchaseData(t *testing.T, db *gorm.DB) (*models.User, *models.ProductVariant, *models.OrderItem) {
	// Create user
	user := createTestUser(db, models.Customer)

	// Create product and variant
	product := createTestProduct(db)
	productVariant := createTestProductVariant(db, product.ID)

	// Create order
	deliveredDate := time.Now()
	order := createTestOrder(db, user.ID, models.OrderStatusDelivered, &deliveredDate)

	// Create order item
	orderItem := createTestOrderItem(db, order.ID, productVariant.ID)

	return &user, &productVariant, &orderItem
}

// setupTestDBWithReviewTables creates a test database with all review-related tables
func setupTestDBWithReviewTables(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate all models including review tables
	err = db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.ProductVariant{},
		&models.Order{},
		&models.OrderItem{},
		&models.ProductReview{},
		&models.ReviewImage{},
		&models.SellerResponse{},
		&models.ReviewHelpful{},
		&models.ProductRating{},
		&models.ReviewModerationLog{},
	)
	require.NoError(t, err)

	// Reset counters
	userCounter = 0
	orderCounter = 0
	skuCounter = 0

	return db
}

func TestCreateReview(t *testing.T) {
	// Setup
	db := setupTestDBWithReviewTables(t)
	handler := NewReviewHandler(db, nil)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupData      func() (uint, uint, uint) // userID, productVariantID, orderItemID
		requestBody    CreateReviewRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful review creation",
			setupData: func() (uint, uint, uint) {
				user, productVariant, orderItem := createTestPurchaseData(t, db)
				return user.ID, productVariant.ID, orderItem.ID
			},
			requestBody: CreateReviewRequest{
				Rating:  5,
				Title:   "Great product!",
				Content: "This product exceeded my expectations. Highly recommended!",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "successful review with order item ID",
			setupData: func() (uint, uint, uint) {
				user, productVariant, orderItem := createTestPurchaseData(t, db)
				return user.ID, productVariant.ID, orderItem.ID
			},
			requestBody: CreateReviewRequest{
				Rating:  4,
				Title:   "Good product",
				Content: "This is a good product with some minor issues.",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid rating - too low",
			setupData: func() (uint, uint, uint) {
				user, productVariant, orderItem := createTestPurchaseData(t, db)
				return user.ID, productVariant.ID, orderItem.ID
			},
			requestBody: CreateReviewRequest{
				Rating:  0,
				Title:   "Test",
				Content: "Test content",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation failed",
		},
		{
			name: "invalid rating - too high",
			setupData: func() (uint, uint, uint) {
				user, productVariant, orderItem := createTestPurchaseData(t, db)
				return user.ID, productVariant.ID, orderItem.ID
			},
			requestBody: CreateReviewRequest{
				Rating:  6,
				Title:   "Test",
				Content: "Test content",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation failed",
		},
		{
			name: "missing content",
			setupData: func() (uint, uint, uint) {
				user, productVariant, orderItem := createTestPurchaseData(t, db)
				return user.ID, productVariant.ID, orderItem.ID
			},
			requestBody: CreateReviewRequest{
				Rating: 5,
				Title:  "Test",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation failed",
		},
		{
			name: "content too long",
			setupData: func() (uint, uint, uint) {
				user, productVariant, orderItem := createTestPurchaseData(t, db)
				return user.ID, productVariant.ID, orderItem.ID
			},
			requestBody: CreateReviewRequest{
				Rating:  5,
				Title:   "Test",
				Content: string(make([]byte, 1001)), // 1001 characters
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation failed",
		},
		{
			name: "title too long",
			setupData: func() (uint, uint, uint) {
				user, productVariant, orderItem := createTestPurchaseData(t, db)
				return user.ID, productVariant.ID, orderItem.ID
			},
			requestBody: CreateReviewRequest{
				Rating:  5,
				Title:   string(make([]byte, 101)), // 101 characters
				Content: "Test content",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation failed",
		},
		{
			name: "too many images",
			setupData: func() (uint, uint, uint) {
				user, productVariant, orderItem := createTestPurchaseData(t, db)
				return user.ID, productVariant.ID, orderItem.ID
			},
			requestBody: CreateReviewRequest{
				Rating:  5,
				Title:   "Test",
				Content: "Test content",
				Images:  []string{"img1.jpg", "img2.jpg", "img3.jpg", "img4.jpg", "img5.jpg", "img6.jpg"},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation failed",
		},
		{
			name: "product variant not found",
			setupData: func() (uint, uint, uint) {
				user, _, orderItem := createTestPurchaseData(t, db)
				return user.ID, 99999, orderItem.ID // Non-existent product variant
			},
			requestBody: CreateReviewRequest{
				Rating:  5,
				Title:   "Test",
				Content: "Test content",
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Product variant not found",
		},
		{
			name: "no purchase verification",
			setupData: func() (uint, uint, uint) {
				user := createTestUser(db, models.Customer)
				product := createTestProduct(db)
				productVariant := createTestProductVariant(db, product.ID)
				return user.ID, productVariant.ID, 0
			},
			requestBody: CreateReviewRequest{
				Rating:  5,
				Title:   "Test",
				Content: "Test content",
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "You must purchase this product before reviewing it",
		},
		{
			name: "duplicate review",
			setupData: func() (uint, uint, uint) {
				user, productVariant, orderItem := createTestPurchaseData(t, db)
				// Create an existing review
				existingReview := models.ProductReview{
					ProductVariantID:   productVariant.ID,
					UserID:             user.ID,
					OrderItemID:        &orderItem.ID,
					Rating:             4,
					Title:              "Existing review",
					Content:            "Existing content",
					IsVerifiedPurchase: true,
					Status:             models.ReviewStatusApproved,
				}
				require.NoError(t, db.Create(&existingReview).Error)
				return user.ID, productVariant.ID, orderItem.ID
			},
			requestBody: CreateReviewRequest{
				Rating:  5,
				Title:   "Test",
				Content: "Test content",
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "You have already reviewed this product",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test data
			userID, productVariantID, orderItemID := tt.setupData()

			// Set product variant ID in request
			tt.requestBody.ProductVariantID = productVariantID
			if orderItemID > 0 {
				tt.requestBody.OrderItemID = &orderItemID
			}

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/reviews", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-token")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("user_id", userID)
			c.Set("user_type", models.Customer)

			// Call handler
			handler.CreateReview(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["message"], tt.expectedError)
			} else {
				// Verify review was created in database
				var review models.ProductReview
				err := db.Where("user_id = ? AND product_variant_id = ?", userID, productVariantID).
					First(&review).Error
				require.NoError(t, err)
				assert.Equal(t, tt.requestBody.Rating, review.Rating)
				assert.Equal(t, tt.requestBody.Title, review.Title)
				assert.Equal(t, tt.requestBody.Content, review.Content)
				assert.Equal(t, models.ReviewStatusPending, review.Status)
				assert.True(t, review.IsVerifiedPurchase)
			}
		})
	}
}

func TestUploadReviewImages(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	// Create a mock Appwrite service for testing
	mockAppwriteService := &aw.AppwriteService{}
	handler := NewReviewHandler(db, mockAppwriteService)
	gin.SetMode(gin.TestMode)

	t.Run("successful image upload", func(t *testing.T) {
		t.Skip("Skipping image upload test - requires mock Appwrite service implementation")
		user := createTestUser(db, models.Customer)

		// Create multipart form request
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add a test file
		part, err := writer.CreateFormFile("images", "test.jpg")
		require.NoError(t, err)
		part.Write([]byte("fake image data"))

		writer.Close()

		req := httptest.NewRequest("POST", "/api/v1/reviews/upload-images", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", user.ID)
		c.Set("user_type", models.Customer)

		handler.UploadReviewImages(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("no images provided", func(t *testing.T) {
		user := createTestUser(db, models.Customer)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest("POST", "/api/v1/reviews/upload-images", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", user.ID)
		c.Set("user_type", models.Customer)

		handler.UploadReviewImages(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetReviewableProducts(t *testing.T) {
	// Setup
	db := setupTestDBWithReviewTables(t)
	handler := NewReviewHandler(db, nil)
	gin.SetMode(gin.TestMode)

	t.Run("successful retrieval", func(t *testing.T) {
		user, _, _ := createTestPurchaseData(t, db)

		req := httptest.NewRequest("GET", "/api/v1/reviews/reviewable-products?limit=10", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", user.ID)
		c.Set("user_type", models.Customer)

		handler.GetReviewableProducts(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Reviewable products retrieved successfully", response["message"])

		data := response["data"].(map[string]interface{})
		products, ok := data["products"].([]interface{})
		require.True(t, ok, "products should be an array")
		assert.Len(t, products, 1)
		assert.Equal(t, float64(1), data["count"])
	})

	t.Run("no reviewable products", func(t *testing.T) {
		user := createTestUser(db, models.Customer)

		req := httptest.NewRequest("GET", "/api/v1/reviews/reviewable-products", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", user.ID)
		c.Set("user_type", models.Customer)

		handler.GetReviewableProducts(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		products, ok := data["products"].([]interface{})
		if ok {
			assert.Len(t, products, 0)
		} else {
			assert.Nil(t, data["products"])
		}
		assert.Equal(t, float64(0), data["count"])
	})
}
