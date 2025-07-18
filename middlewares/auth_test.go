package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupAuthTest() {
	gin.SetMode(gin.TestMode)
	os.Setenv("JWT_SECRET", "test-secret-key")
}

func TestAuthMiddleware(t *testing.T) {
	setupAuthTest()

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "No Authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid token format",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Bearer token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Valid token",
			authHeader:     "Bearer " + generateTestToken(1, models.Customer),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(AuthMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAdminMiddleware(t *testing.T) {
	setupAuthTest()

	tests := []struct {
		name           string
		userType       models.UserType
		expectedStatus int
	}{
		{
			name:           "Customer access denied",
			userType:       models.Customer,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Vendor access denied",
			userType:       models.Vendor,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Admin access allowed",
			userType:       models.Admin,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(AdminMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+generateTestToken(1, tt.userType))
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSellerMiddleware(t *testing.T) {
	setupAuthTest()

	tests := []struct {
		name           string
		userType       models.UserType
		expectedStatus int
	}{
		{
			name:           "Customer access denied",
			userType:       models.Customer,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Vendor access allowed",
			userType:       models.Vendor,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Admin access allowed (has seller privileges)",
			userType:       models.Admin,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(SellerMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+generateTestToken(1, tt.userType))
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthMiddlewareSetsUserContext(t *testing.T) {
	setupAuthTest()

	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, uint(1), userID)

		userType, exists := c.Get("user_type")
		assert.True(t, exists)
		assert.Equal(t, models.Customer, userType)

		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(1, models.Customer))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminMiddlewareAbortsCorrectly(t *testing.T) {
	// Test that admin middleware properly aborts and doesn't execute handler
	router := gin.New()

	// Create a flag to track if handler was called
	handlerCalled := false

	// Add admin middleware and a test handler
	router.GET("/admin/test", AdminMiddleware(), func(c *gin.Context) {
		handlerCalled = true
		c.JSON(200, gin.H{"message": "handler executed"})
	})

	// Test without token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/test", nil)
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, 401, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "auth/middleware", response["error"].(map[string]interface{})["code"])
	assert.Equal(t, "token is required", response["error"].(map[string]interface{})["description"])

	// Verify handler was not called
	assert.False(t, handlerCalled, "Handler should not be called when middleware aborts")
}

func TestSellerMiddlewareAbortsCorrectly(t *testing.T) {
	// Test that seller middleware properly aborts and doesn't execute handler
	router := gin.New()

	// Create a flag to track if handler was called
	handlerCalled := false

	// Add seller middleware and a test handler
	router.GET("/seller/test", SellerMiddleware(), func(c *gin.Context) {
		handlerCalled = true
		c.JSON(200, gin.H{"message": "handler executed"})
	})

	// Test without token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/seller/test", nil)
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, 401, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "auth/middleware", response["error"].(map[string]interface{})["code"])
	assert.Equal(t, "token is required", response["error"].(map[string]interface{})["description"])

	// Verify handler was not called
	assert.False(t, handlerCalled, "Handler should not be called when middleware aborts")
}

func generateTestToken(userID uint, userType models.UserType) string {
	token, _ := auth.GenerateToken(userID, userType, nil)
	return token
}
