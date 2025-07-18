package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupReviewModelTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&User{},
		&Product{},
		&ProductVariant{},
		&ProductReview{},
		&ReviewImage{},
		&SellerResponse{},
		&ReviewHelpful{},
		&ProductRating{},
		&ReviewModerationLog{},
	)
	require.NoError(t, err)

	return db
}

func createTestUser(t *testing.T, db *gorm.DB, userType UserType) *User {
	user := &User{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		UserType:  userType,
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func createTestUserWithEmail(t *testing.T, db *gorm.DB, email string, userType UserType) *User {
	user := &User{
		Email:     email,
		FirstName: "Test",
		LastName:  "User",
		UserType:  userType,
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func createTestProductVariant(t *testing.T, db *gorm.DB) *ProductVariant {
	product := &Product{
		Name:     "Test Product",
		IsActive: true,
	}
	err := db.Create(product).Error
	require.NoError(t, err)

	variant := &ProductVariant{
		ProductID: product.ID,
		Name:      "Test Variant",
		SKU:       "TEST-SKU",
		BasePrice: 10.0,
		IsActive:  true,
	}
	err = db.Create(variant).Error
	require.NoError(t, err)

	return variant
}

// TestProductReview_BeforeCreate tests the GORM BeforeCreate hook
func TestProductReview_BeforeCreate(t *testing.T) {
	db := setupReviewModelTestDB(t)
	user := createTestUser(t, db, Customer)
	variant := createTestProductVariant(t, db)

	// Test that default status is set when empty
	review := &ProductReview{
		ProductVariantID: variant.ID,
		UserID:           user.ID,
		Rating:           5,
		Content:          "Great product!",
		Status:           "", // Empty status
	}

	err := db.Create(review).Error
	require.NoError(t, err)
	assert.Equal(t, ReviewStatusPending, review.Status)

	// Test that existing status is preserved
	review2 := &ProductReview{
		ProductVariantID: variant.ID,
		UserID:           user.ID,
		Rating:           4,
		Content:          "Good product",
		Status:           ReviewStatusApproved, // Pre-set status
	}

	err = db.Create(review2).Error
	require.NoError(t, err)
	assert.Equal(t, ReviewStatusApproved, review2.Status)
}

// TestProductReview_StatusMethods tests the status checking methods
func TestProductReview_StatusMethods(t *testing.T) {
	db := setupReviewModelTestDB(t)
	user := createTestUser(t, db, Customer)
	variant := createTestProductVariant(t, db)

	// Test pending review
	review := &ProductReview{
		ProductVariantID: variant.ID,
		UserID:           user.ID,
		Rating:           5,
		Content:          "Test review",
		Status:           ReviewStatusPending,
	}
	err := db.Create(review).Error
	require.NoError(t, err)

	assert.True(t, review.IsPending())
	assert.False(t, review.IsApproved())
	assert.False(t, review.IsRejected())
	assert.False(t, review.IsFlagged())

	// Test approved review
	review.Status = ReviewStatusApproved
	err = db.Save(review).Error
	require.NoError(t, err)

	assert.False(t, review.IsPending())
	assert.True(t, review.IsApproved())
	assert.False(t, review.IsRejected())
	assert.False(t, review.IsFlagged())

	// Test rejected review
	review.Status = ReviewStatusRejected
	err = db.Save(review).Error
	require.NoError(t, err)

	assert.False(t, review.IsPending())
	assert.False(t, review.IsApproved())
	assert.True(t, review.IsRejected())
	assert.False(t, review.IsFlagged())

	// Test flagged review
	review.Status = ReviewStatusFlagged
	err = db.Save(review).Error
	require.NoError(t, err)

	assert.False(t, review.IsPending())
	assert.False(t, review.IsApproved())
	assert.False(t, review.IsRejected())
	assert.True(t, review.IsFlagged())
}

// TestProductReview_CanBeModifiedBy tests the modification permission logic
func TestProductReview_CanBeModifiedBy(t *testing.T) {
	db := setupReviewModelTestDB(t)
	customer := createTestUserWithEmail(t, db, "customer@example.com", Customer)
	admin := createTestUserWithEmail(t, db, "admin@example.com", Admin)
	variant := createTestProductVariant(t, db)

	// Create a pending review by customer
	review := &ProductReview{
		ProductVariantID: variant.ID,
		UserID:           customer.ID,
		Rating:           5,
		Content:          "Test review",
		Status:           ReviewStatusPending,
	}
	err := db.Create(review).Error
	require.NoError(t, err)

	// Test admin can modify any review
	assert.True(t, review.CanBeModifiedBy(admin.ID, Admin))

	// Test customer can modify their own pending review
	assert.True(t, review.CanBeModifiedBy(customer.ID, Customer))

	// Test customer cannot modify other customer's review
	otherCustomer := createTestUserWithEmail(t, db, "other@example.com", Customer)
	assert.False(t, review.CanBeModifiedBy(otherCustomer.ID, Customer))

	// Test customer cannot modify rejected review
	review.Status = ReviewStatusRejected
	err = db.Save(review).Error
	require.NoError(t, err)
	assert.False(t, review.CanBeModifiedBy(customer.ID, Customer))

	// Test customer can modify their own approved review
	review.Status = ReviewStatusApproved
	err = db.Save(review).Error
	require.NoError(t, err)
	assert.True(t, review.CanBeModifiedBy(customer.ID, Customer))
}

// TestProductReview_CanBeDeletedBy tests the deletion permission logic
func TestProductReview_CanBeDeletedBy(t *testing.T) {
	db := setupReviewModelTestDB(t)
	customer := createTestUserWithEmail(t, db, "customer2@example.com", Customer)
	admin := createTestUserWithEmail(t, db, "admin2@example.com", Admin)
	variant := createTestProductVariant(t, db)

	// Create a review by customer
	review := &ProductReview{
		ProductVariantID: variant.ID,
		UserID:           customer.ID,
		Rating:           5,
		Content:          "Test review",
		Status:           ReviewStatusApproved,
	}
	err := db.Create(review).Error
	require.NoError(t, err)

	// Test admin can delete any review
	assert.True(t, review.CanBeDeletedBy(admin.ID, Admin))

	// Test customer can delete their own review
	assert.True(t, review.CanBeDeletedBy(customer.ID, Customer))

	// Test customer cannot delete other customer's review
	otherCustomer := createTestUserWithEmail(t, db, "other2@example.com", Customer)
	assert.False(t, review.CanBeDeletedBy(otherCustomer.ID, Customer))
}

// TestProductReview_HasSellerResponse tests the seller response checking
func TestProductReview_HasSellerResponse(t *testing.T) {
	db := setupReviewModelTestDB(t)
	customer := createTestUserWithEmail(t, db, "customer3@example.com", Customer)
	seller := createTestUserWithEmail(t, db, "seller@example.com", Vendor)
	variant := createTestProductVariant(t, db)

	// Create a review
	review := &ProductReview{
		ProductVariantID: variant.ID,
		UserID:           customer.ID,
		Rating:           5,
		Content:          "Test review",
		Status:           ReviewStatusApproved,
	}
	err := db.Create(review).Error
	require.NoError(t, err)

	// Initially no seller response
	assert.False(t, review.HasSellerResponse())

	// Add seller response
	sellerResponse := &SellerResponse{
		ProductReviewID: review.ID,
		UserID:          seller.ID,
		Content:         "Thank you for your review!",
	}
	err = db.Create(sellerResponse).Error
	require.NoError(t, err)

	// Reload review with seller response
	err = db.Preload("SellerResponse").First(review, review.ID).Error
	require.NoError(t, err)

	assert.True(t, review.HasSellerResponse())
	assert.NotNil(t, review.SellerResponse)
	assert.Equal(t, sellerResponse.Content, review.SellerResponse.Content)
}

// TestProductReview_GetReviewerName tests the reviewer name generation
func TestProductReview_GetReviewerName(t *testing.T) {
	db := setupReviewModelTestDB(t)
	variant := createTestProductVariant(t, db)

	// Test with full name
	user1 := &User{
		Email:     "user1@example.com",
		FirstName: "John",
		LastName:  "Doe",
		UserType:  Customer,
	}
	err := db.Create(user1).Error
	require.NoError(t, err)

	review1 := &ProductReview{
		ProductVariantID: variant.ID,
		UserID:           user1.ID,
		Rating:           5,
		Content:          "Test review",
		Status:           ReviewStatusApproved,
	}
	err = db.Create(review1).Error
	require.NoError(t, err)

	// Reload with user data
	err = db.Preload("User").First(review1, review1.ID).Error
	require.NoError(t, err)

	assert.Equal(t, "John Doe", review1.GetReviewerName())

	// Test with first name only
	user2 := &User{
		Email:     "user2@example.com",
		FirstName: "Jane",
		UserType:  Customer,
	}
	err = db.Create(user2).Error
	require.NoError(t, err)

	review2 := &ProductReview{
		ProductVariantID: variant.ID,
		UserID:           user2.ID,
		Rating:           4,
		Content:          "Test review 2",
		Status:           ReviewStatusApproved,
	}
	err = db.Create(review2).Error
	require.NoError(t, err)

	// Reload with user data
	err = db.Preload("User").First(review2, review2.ID).Error
	require.NoError(t, err)

	assert.Equal(t, "Jane", review2.GetReviewerName())

	// Test with no name (should return "Anonymous")
	user3 := &User{
		Email:    "user3@example.com",
		UserType: Customer,
	}
	err = db.Create(user3).Error
	require.NoError(t, err)

	review3 := &ProductReview{
		ProductVariantID: variant.ID,
		UserID:           user3.ID,
		Rating:           3,
		Content:          "Test review 3",
		Status:           ReviewStatusApproved,
	}
	err = db.Create(review3).Error
	require.NoError(t, err)

	// Reload with user data
	err = db.Preload("User").First(review3, review3.ID).Error
	require.NoError(t, err)

	assert.Equal(t, "Anonymous", review3.GetReviewerName())
}

// TestReviewImage_Validation tests the ReviewImage validation
func TestReviewImage_Validation(t *testing.T) {
	db := setupReviewModelTestDB(t)
	user := createTestUserWithEmail(t, db, "customer5@example.com", Customer)
	variant := createTestProductVariant(t, db)

	// Create a review first
	review := &ProductReview{
		ProductVariantID: variant.ID,
		UserID:           user.ID,
		Rating:           5,
		Content:          "Test review",
		Status:           ReviewStatusApproved,
	}
	err := db.Create(review).Error
	require.NoError(t, err)

	// Test valid review image
	image := &ReviewImage{
		ProductReviewID: review.ID,
		URL:             "https://example.com/image.jpg",
		AltText:         "Product image",
	}
	err = db.Create(image).Error
	require.NoError(t, err)

	// Note: GORM validation tags are not enforced at the database level in SQLite
	// These tests would need a custom validator to work properly
	// For now, we'll test that valid images can be created
	assert.NotZero(t, image.ID)
}

// TestSellerResponse_Validation tests the SellerResponse validation
func TestSellerResponse_Validation(t *testing.T) {
	db := setupReviewModelTestDB(t)
	customer := createTestUserWithEmail(t, db, "customer4@example.com", Customer)
	seller := createTestUserWithEmail(t, db, "seller2@example.com", Vendor)
	variant := createTestProductVariant(t, db)

	// Create a review first
	review := &ProductReview{
		ProductVariantID: variant.ID,
		UserID:           customer.ID,
		Rating:           5,
		Content:          "Test review",
		Status:           ReviewStatusApproved,
	}
	err := db.Create(review).Error
	require.NoError(t, err)

	// Test valid seller response
	response := &SellerResponse{
		ProductReviewID: review.ID,
		UserID:          seller.ID,
		Content:         "Thank you for your review!",
	}
	err = db.Create(response).Error
	require.NoError(t, err)

	// Note: GORM validation tags are not enforced at the database level in SQLite
	// These tests would need a custom validator to work properly
	// For now, we'll test that valid responses can be created
	assert.NotZero(t, response.ID)

	// Test that we can't create a second response for the same review (unique constraint)
	duplicateResponse := &SellerResponse{
		ProductReviewID: review.ID,
		UserID:          seller.ID,
		Content:         "Another response",
	}
	err = db.Create(duplicateResponse).Error
	assert.Error(t, err) // Should fail due to unique constraint
}

// TestProductRating_TableName tests the table name override
func TestProductRating_TableName(t *testing.T) {
	rating := &ProductRating{}
	assert.Equal(t, "product_ratings", rating.TableName())
}

// TestReviewModerationLog_TableName tests the table name override
func TestReviewModerationLog_TableName(t *testing.T) {
	log := &ReviewModerationLog{}
	assert.Equal(t, "review_moderation_logs", log.TableName())
}

// TestReviewModerationLog_Creation tests the moderation log creation
func TestReviewModerationLog_Creation(t *testing.T) {
	db := setupReviewModelTestDB(t)
	admin := createTestUserWithEmail(t, db, "admin3@example.com", Admin)
	customer := createTestUserWithEmail(t, db, "customer6@example.com", Customer)
	variant := createTestProductVariant(t, db)

	// Create a review
	review := &ProductReview{
		ProductVariantID: variant.ID,
		UserID:           customer.ID,
		Rating:           5,
		Content:          "Test review",
		Status:           ReviewStatusPending,
	}
	err := db.Create(review).Error
	require.NoError(t, err)

	// Create moderation log
	log := &ReviewModerationLog{
		ReviewID:    review.ID,
		AdminID:     admin.ID,
		OldStatus:   ReviewStatusPending,
		NewStatus:   ReviewStatusApproved,
		Reason:      "Review approved",
		ModeratedAt: time.Now(),
	}
	err = db.Create(log).Error
	require.NoError(t, err)

	// Verify log was created
	var foundLog ReviewModerationLog
	err = db.Preload("Review").Preload("Admin").First(&foundLog, log.ID).Error
	require.NoError(t, err)

	assert.Equal(t, review.ID, foundLog.ReviewID)
	assert.Equal(t, admin.ID, foundLog.AdminID)
	assert.Equal(t, ReviewStatusPending, foundLog.OldStatus)
	assert.Equal(t, ReviewStatusApproved, foundLog.NewStatus)
	assert.Equal(t, "Review approved", foundLog.Reason)
	assert.NotNil(t, foundLog.Review)
	assert.NotNil(t, foundLog.Admin)
}

// TestReviewHelpful_Creation tests the helpful vote creation
func TestReviewHelpful_Creation(t *testing.T) {
	db := setupReviewModelTestDB(t)
	customer := createTestUser(t, db, Customer)
	variant := createTestProductVariant(t, db)

	// Create a review
	review := &ProductReview{
		ProductVariantID: variant.ID,
		UserID:           customer.ID,
		Rating:           5,
		Content:          "Test review",
		Status:           ReviewStatusApproved,
	}
	err := db.Create(review).Error
	require.NoError(t, err)

	// Create helpful vote
	vote := &ReviewHelpful{
		ProductReviewID: review.ID,
		UserID:          customer.ID,
		IsHelpful:       true,
	}
	err = db.Create(vote).Error
	require.NoError(t, err)

	// Verify vote was created
	var foundVote ReviewHelpful
	err = db.First(&foundVote, vote.ID).Error
	require.NoError(t, err)

	assert.Equal(t, review.ID, foundVote.ProductReviewID)
	assert.Equal(t, customer.ID, foundVote.UserID)
	assert.True(t, foundVote.IsHelpful)
}

// TestProductReview_SoftDelete tests soft delete functionality
func TestProductReview_SoftDelete(t *testing.T) {
	db := setupReviewModelTestDB(t)
	user := createTestUser(t, db, Customer)
	variant := createTestProductVariant(t, db)

	// Create a review
	review := &ProductReview{
		ProductVariantID: variant.ID,
		UserID:           user.ID,
		Rating:           5,
		Content:          "Test review",
		Status:           ReviewStatusApproved,
	}
	err := db.Create(review).Error
	require.NoError(t, err)

	// Soft delete the review
	err = db.Delete(review).Error
	require.NoError(t, err)

	// Verify review is soft deleted (not found with normal query)
	var foundReview ProductReview
	err = db.First(&foundReview, review.ID).Error
	assert.Error(t, err) // Should not be found

	// Verify review exists with Unscoped query
	err = db.Unscoped().First(&foundReview, review.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, foundReview.DeletedAt) // Should have deleted_at timestamp
}

// TestProductReview_Relationships tests the model relationships
func TestProductReview_Relationships(t *testing.T) {
	db := setupReviewModelTestDB(t)
	user := createTestUser(t, db, Customer)
	variant := createTestProductVariant(t, db)

	// Create a review with relationships
	review := &ProductReview{
		ProductVariantID: variant.ID,
		UserID:           user.ID,
		Rating:           5,
		Content:          "Test review",
		Status:           ReviewStatusApproved,
	}
	err := db.Create(review).Error
	require.NoError(t, err)

	// Add images
	images := []ReviewImage{
		{
			ProductReviewID: review.ID,
			URL:             "https://example.com/image1.jpg",
			AltText:         "Image 1",
		},
		{
			ProductReviewID: review.ID,
			URL:             "https://example.com/image2.jpg",
			AltText:         "Image 2",
		},
	}
	for i := range images {
		err = db.Create(&images[i]).Error
		require.NoError(t, err)
	}

	// Add helpful votes
	votes := []ReviewHelpful{
		{
			ProductReviewID: review.ID,
			UserID:          user.ID,
			IsHelpful:       true,
		},
	}
	for i := range votes {
		err = db.Create(&votes[i]).Error
		require.NoError(t, err)
	}

	// Load review with all relationships
	err = db.Preload("ProductVariant").
		Preload("User").
		Preload("Images").
		Preload("HelpfulVotes").
		First(review, review.ID).Error
	require.NoError(t, err)

	// Verify relationships
	assert.NotNil(t, review.ProductVariant)
	assert.Equal(t, variant.ID, review.ProductVariant.ID)

	assert.NotNil(t, review.User)
	assert.Equal(t, user.ID, review.User.ID)

	assert.Len(t, review.Images, 2)
	assert.Len(t, review.HelpfulVotes, 1)
}
