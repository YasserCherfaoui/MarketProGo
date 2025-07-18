package review

import (
	"fmt"
	"testing"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var userCounter int
var orderCounter int
var skuCounter int

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the models
	err = db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.ProductVariant{},
		&models.Order{},
		&models.OrderItem{},
		&models.ProductReview{},
	)
	require.NoError(t, err)

	userCounter = 0
	orderCounter = 0
	skuCounter = 0

	return db
}

func createTestUser(db *gorm.DB, userType models.UserType) models.User {
	userCounter++
	email := fmt.Sprintf("testuser%d@example.com", userCounter)
	user := models.User{
		Email:    email,
		Password: "password",
		UserType: userType,
	}
	db.Create(&user)
	return user
}

func createTestProduct(db *gorm.DB) models.Product {
	product := models.Product{
		Name:        "Test Product",
		Description: "Test Description",
		IsActive:    true,
	}
	db.Create(&product)
	return product
}

func createTestProductVariant(db *gorm.DB, productID uint) models.ProductVariant {
	skuCounter++
	sku := fmt.Sprintf("TEST-SKU-%03d", skuCounter)
	variant := models.ProductVariant{
		ProductID: productID,
		Name:      fmt.Sprintf("Test Variant %d", skuCounter),
		SKU:       sku,
		BasePrice: 10.99,
		IsActive:  true,
	}
	db.Create(&variant)
	return variant
}

func createTestOrder(db *gorm.DB, userID uint, status models.OrderStatus, deliveredDate *time.Time) models.Order {
	orderCounter++
	orderNumber := fmt.Sprintf("ORD-%03d", orderCounter)
	order := models.Order{
		OrderNumber:   orderNumber,
		UserID:        userID,
		Status:        status,
		PaymentStatus: models.PaymentStatusPaid,
		TotalAmount:   10.99,
		FinalAmount:   10.99,
		OrderDate:     time.Now(),
		DeliveredDate: deliveredDate,
	}
	db.Create(&order)
	return order
}

func createTestOrderItem(db *gorm.DB, orderID, productVariantID uint) models.OrderItem {
	orderItem := models.OrderItem{
		OrderID:          orderID,
		ProductVariantID: productVariantID,
		Quantity:         1,
		UnitPrice:        10.99,
		TotalAmount:      10.99,
		Status:           "active",
	}
	db.Create(&orderItem)
	return orderItem
}

func TestVerifyPurchase(t *testing.T) {
	db := setupTestDB(t)
	handler := &ReviewHandler{db: db}

	// Create test data
	user := createTestUser(db, models.Customer)
	product := createTestProduct(db)

	// Test case 1: Valid purchase
	variant1 := createTestProductVariant(db, product.ID)
	deliveredDate := time.Now().AddDate(0, -1, 0) // 1 month ago
	order := createTestOrder(db, user.ID, models.OrderStatusDelivered, &deliveredDate)
	orderItem := createTestOrderItem(db, order.ID, variant1.ID)

	result, err := handler.VerifyPurchase(user.ID, variant1.ID)
	require.NoError(t, err)
	assert.True(t, result.IsVerified)
	assert.NotNil(t, result.OrderItem)
	assert.Equal(t, orderItem.ID, result.OrderItem.ID)

	// Test case 2: No purchase found
	result, err = handler.VerifyPurchase(user.ID, 999)
	require.NoError(t, err)
	assert.False(t, result.IsVerified)
	if !result.IsVerified {
		assert.NotEmpty(t, result.ErrorMessage)
	}

	// Test case 3: Order not delivered
	variant2 := createTestProductVariant(db, product.ID)
	order2 := createTestOrder(db, user.ID, models.OrderStatusPending, nil)
	createTestOrderItem(db, order2.ID, variant2.ID)

	result, err = handler.VerifyPurchase(user.ID, variant2.ID)
	require.NoError(t, err)
	assert.False(t, result.IsVerified)
	if !result.IsVerified {
		assert.NotEmpty(t, result.ErrorMessage)
	}

	// Test case 4: Order too old (more than 2 years)
	variant3 := createTestProductVariant(db, product.ID)
	oldDate := time.Now().AddDate(-3, 0, 0) // 3 years ago
	order3 := createTestOrder(db, user.ID, models.OrderStatusDelivered, &oldDate)
	createTestOrderItem(db, order3.ID, variant3.ID)

	result, err = handler.VerifyPurchase(user.ID, variant3.ID)
	require.NoError(t, err)
	assert.False(t, result.IsVerified)
	if !result.IsVerified {
		assert.NotEmpty(t, result.ErrorMessage)
	}
}

func TestVerifyPurchaseWithOrderItemID(t *testing.T) {
	db := setupTestDB(t)
	handler := &ReviewHandler{db: db}

	// Create test data
	user := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	variant := createTestProductVariant(db, product.ID)

	deliveredDate := time.Now().AddDate(0, -1, 0)
	order := createTestOrder(db, user.ID, models.OrderStatusDelivered, &deliveredDate)
	orderItem := createTestOrderItem(db, order.ID, variant.ID)

	// Test case 1: Valid order item
	result, err := handler.VerifyPurchaseWithOrderItemID(user.ID, orderItem.ID)
	require.NoError(t, err)
	assert.True(t, result.IsVerified)
	assert.Equal(t, orderItem.ID, result.OrderItem.ID)

	// Test case 2: Order item not found
	result, err = handler.VerifyPurchaseWithOrderItemID(user.ID, 999)
	require.NoError(t, err)
	assert.False(t, result.IsVerified)
	assert.Contains(t, result.ErrorMessage, "Order item not found")

	// Test case 3: Order item belongs to different user
	otherUser := createTestUser(db, models.Customer)
	result, err = handler.VerifyPurchaseWithOrderItemID(otherUser.ID, orderItem.ID)
	require.NoError(t, err)
	assert.False(t, result.IsVerified)
	assert.Contains(t, result.ErrorMessage, "Order item not found")
}

func TestCheckIfUserCanReview(t *testing.T) {
	db := setupTestDB(t)
	handler := &ReviewHandler{db: db}

	// Create test data
	user := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	variant := createTestProductVariant(db, product.ID)

	deliveredDate := time.Now().AddDate(0, -1, 0)
	order := createTestOrder(db, user.ID, models.OrderStatusDelivered, &deliveredDate)
	_ = createTestOrderItem(db, order.ID, variant.ID)

	// Test case 1: User can review (no existing review)
	result, err := handler.CheckIfUserCanReview(user.ID, variant.ID)
	require.NoError(t, err)
	assert.True(t, result.IsVerified)

	// Test case 2: User already reviewed
	review := models.ProductReview{
		ProductVariantID: variant.ID,
		UserID:           user.ID,
		Rating:           5,
		Status:           models.ReviewStatusApproved,
	}
	db.Create(&review)

	result, err = handler.CheckIfUserCanReview(user.ID, variant.ID)
	require.NoError(t, err)
	assert.False(t, result.IsVerified)
	assert.Contains(t, result.ErrorMessage, "You have already reviewed this product")

	// Test case 3: User hasn't purchased
	result, err = handler.CheckIfUserCanReview(user.ID, 999)
	require.NoError(t, err)
	assert.False(t, result.IsVerified)
	assert.Contains(t, result.ErrorMessage, "No verified purchase found")
}

func TestGetUserPurchasedProducts(t *testing.T) {
	db := setupTestDB(t)
	handler := &ReviewHandler{db: db}

	// Create test data
	user := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	variant1 := createTestProductVariant(db, product.ID)
	variant2 := createTestProductVariant(db, product.ID)

	deliveredDate := time.Now().AddDate(0, -1, 0)
	order1 := createTestOrder(db, user.ID, models.OrderStatusDelivered, &deliveredDate)
	order2 := createTestOrder(db, user.ID, models.OrderStatusDelivered, &deliveredDate)

	createTestOrderItem(db, order1.ID, variant1.ID)
	createTestOrderItem(db, order2.ID, variant2.ID)

	// Test case 1: Get all purchased products
	products, err := handler.GetUserPurchasedProducts(user.ID, 0)
	require.NoError(t, err)
	assert.Len(t, products, 2)

	// Test case 2: Get with limit
	products, err = handler.GetUserPurchasedProducts(user.ID, 1)
	require.NoError(t, err)
	assert.Len(t, products, 1)

	// Test case 3: User with no purchases
	otherUser := createTestUser(db, models.Customer)
	products, err = handler.GetUserPurchasedProducts(otherUser.ID, 0)
	require.NoError(t, err)
	assert.Len(t, products, 0)
}

func TestGetReviewableProductsForUser(t *testing.T) {
	db := setupTestDB(t)
	handler := &ReviewHandler{db: db}

	// Create test data
	user := createTestUser(db, models.Customer)
	product := createTestProduct(db)
	variant1 := createTestProductVariant(db, product.ID)
	variant2 := createTestProductVariant(db, product.ID)

	deliveredDate := time.Now().AddDate(0, -1, 0)
	order1 := createTestOrder(db, user.ID, models.OrderStatusDelivered, &deliveredDate)
	order2 := createTestOrder(db, user.ID, models.OrderStatusDelivered, &deliveredDate)

	createTestOrderItem(db, order1.ID, variant1.ID)
	createTestOrderItem(db, order2.ID, variant2.ID)

	// Create a review for variant1
	review := models.ProductReview{
		ProductVariantID: variant1.ID,
		UserID:           user.ID,
		Rating:           5,
		Status:           models.ReviewStatusApproved,
	}
	db.Create(&review)

	// Test case 1: Get reviewable products (should exclude variant1)
	products, err := handler.GetReviewableProductsForUser(user.ID, 0)
	require.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, variant2.ID, products[0].ProductVariantID)

	// Test case 2: Get with limit
	products, err = handler.GetReviewableProductsForUser(user.ID, 1)
	require.NoError(t, err)
	assert.Len(t, products, 1)

	// Test case 3: User with no reviewable products
	otherUser := createTestUser(db, models.Customer)
	products, err = handler.GetReviewableProductsForUser(otherUser.ID, 0)
	require.NoError(t, err)
	assert.Len(t, products, 0)
}
