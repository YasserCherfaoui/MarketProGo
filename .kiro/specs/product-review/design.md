# Design Document

## Overview

The product review system will be implemented as a new domain within the existing MarketProGo architecture, following the established patterns of domain-driven layered architecture. The system will integrate with existing Product, User, and Order models to provide comprehensive review functionality with proper authorization and data integrity.

## Architecture

### Domain Structure
Following the existing pattern, the review system will be organized as:
```
handlers/review/          # HTTP request handlers
├── handler.go           # Handler struct and constructor
├── create.go           # Submit review
├── get.go              # Get single review
├── get_all.go          # Get reviews for product
├── update.go           # Update review (customer) or respond (seller)
├── delete.go           # Delete review (admin/customer)
├── moderate.go         # Admin moderation actions
└── helpful.go          # Mark review as helpful
```

### Integration Points
- **Product Model**: Reviews will reference ProductVariant for specific product reviews
- **User Model**: Reviews will be linked to authenticated users with purchase verification
- **Order Model**: Purchase verification through OrderItem relationships
- **Authentication**: JWT middleware for protected endpoints
- **File Storage**: Appwrite integration for review images

## Components and Interfaces

### Core Models

#### ProductReview Model
```go
type ProductReview struct {
    gorm.Model
    ProductVariantID uint           `json:"product_variant_id"`
    ProductVariant   ProductVariant `json:"product_variant"`
    UserID           uint           `json:"user_id"`
    User             User           `json:"user"`
    OrderItemID      *uint          `json:"order_item_id"` // For purchase verification
    OrderItem        *OrderItem     `json:"order_item,omitempty"`
    
    Rating           int            `json:"rating" validate:"min=1,max=5"`
    Title            string         `json:"title" validate:"max=100"`
    Content          string         `json:"content" validate:"max=1000"`
    IsVerifiedPurchase bool         `json:"is_verified_purchase"`
    
    // Moderation
    Status           ReviewStatus   `json:"status"`
    ModeratedBy      *uint          `json:"moderated_by"`
    ModeratedAt      *time.Time     `json:"moderated_at"`
    ModerationReason string         `json:"moderation_reason"`
    
    // Engagement
    HelpfulCount     int            `json:"helpful_count"`
    
    // Media
    Images           []ReviewImage  `json:"images"`
    
    // Seller Response
    SellerResponse   *SellerResponse `json:"seller_response,omitempty"`
}

type ReviewStatus string
const (
    ReviewStatusPending  ReviewStatus = "PENDING"
    ReviewStatusApproved ReviewStatus = "APPROVED" 
    ReviewStatusRejected ReviewStatus = "REJECTED"
    ReviewStatusFlagged  ReviewStatus = "FLAGGED"
)

type ReviewImage struct {
    gorm.Model
    ProductReviewID uint   `json:"product_review_id"`
    URL             string `json:"url"`
    AltText         string `json:"alt_text"`
}

type SellerResponse struct {
    gorm.Model
    ProductReviewID uint      `json:"product_review_id"`
    UserID          uint      `json:"user_id"` // Seller user ID
    User            User      `json:"user"`
    Content         string    `json:"content" validate:"max=500"`
    UpdatedAt       time.Time `json:"updated_at"`
}

type ReviewHelpful struct {
    gorm.Model
    ProductReviewID uint `json:"product_review_id"`
    UserID          uint `json:"user_id"`
    IsHelpful       bool `json:"is_helpful"`
}
```

#### Product Rating Aggregation
```go
type ProductRating struct {
    gorm.Model
    ProductVariantID uint    `json:"product_variant_id" gorm:"uniqueIndex"`
    AverageRating    float64 `json:"average_rating"`
    TotalReviews     int     `json:"total_reviews"`
    RatingBreakdown  string  `json:"rating_breakdown"` // JSON: {"1":0,"2":1,"3":2,"4":5,"5":10}
}
```

### Handler Structure

#### ReviewHandler
```go
type ReviewHandler struct {
    db              *gorm.DB
    appwriteService *aw.AppwriteService
}

func NewReviewHandler(db *gorm.DB, appwriteService *aw.AppwriteService) *ReviewHandler
```

### API Endpoints

#### Public Endpoints
- `GET /api/v1/products/:productId/reviews` - Get paginated reviews for a product
- `GET /api/v1/reviews/:id` - Get single review with responses

#### Authenticated Customer Endpoints  
- `POST /api/v1/reviews` - Submit new review
- `PUT /api/v1/reviews/:id` - Update own review
- `DELETE /api/v1/reviews/:id` - Delete own review
- `POST /api/v1/reviews/:id/helpful` - Mark review as helpful/unhelpful

#### Seller Endpoints
- `GET /api/v1/seller/reviews` - Get reviews for seller's products
- `POST /api/v1/reviews/:id/response` - Respond to review
- `PUT /api/v1/reviews/:id/response` - Update response

#### Admin Endpoints
- `GET /api/v1/admin/reviews` - Get all reviews with moderation status
- `PUT /api/v1/admin/reviews/:id/moderate` - Moderate review (approve/reject/flag)
- `DELETE /api/v1/admin/reviews/:id` - Permanently delete review

## Data Models

### Database Schema Relationships
```
ProductVariant (1) ←→ (N) ProductReview
User (1) ←→ (N) ProductReview  
OrderItem (1) ←→ (0..1) ProductReview
ProductReview (1) ←→ (N) ReviewImage
ProductReview (1) ←→ (0..1) SellerResponse
ProductReview (1) ←→ (N) ReviewHelpful
ProductVariant (1) ←→ (1) ProductRating
```

### Validation Rules
- Rating: Required, integer between 1-5
- Content: Optional, max 1000 characters
- Title: Optional, max 100 characters
- Images: Max 5 images per review
- Purchase verification: Required for review submission
- Duplicate prevention: One review per user per product variant

### Business Logic

#### Purchase Verification
```go
func (h *ReviewHandler) verifyPurchase(userID uint, productVariantID uint) (*models.OrderItem, error) {
    var orderItem models.OrderItem
    err := h.db.
        Joins("JOIN orders ON orders.id = order_items.order_id").
        Where("orders.user_id = ? AND order_items.product_variant_id = ? AND orders.status = ?", 
              userID, productVariantID, models.OrderStatusDelivered).
        First(&orderItem).Error
    return &orderItem, err
}
```

#### Rating Aggregation
```go
func (h *ReviewHandler) updateProductRating(productVariantID uint) error {
    // Calculate new average and breakdown
    // Update ProductRating record
    // Trigger any necessary cache invalidation
}
```

## Error Handling

### Custom Error Types
```go
type ReviewError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

const (
    ErrReviewNotFound        = "REVIEW_NOT_FOUND"
    ErrReviewAlreadyExists   = "REVIEW_ALREADY_EXISTS"
    ErrReviewNotAuthorized   = "REVIEW_NOT_AUTHORIZED"
    ErrReviewPurchaseRequired = "PURCHASE_REQUIRED"
    ErrReviewModerationFailed = "MODERATION_FAILED"
)
```

### Error Response Format
Following existing pattern using `utils/response` package:
```go
response.GenerateErrorResponse(c, "review/create", "Purchase required to review this product", 403)
response.GenerateValidationErrorResponse(c, "review/create", validationErrors)
response.GenerateNotFoundResponse(c, "review/get", "Review not found")
```

## Testing Strategy

### Unit Tests
- Model validation tests for all review models
- Business logic tests for purchase verification
- Rating aggregation calculation tests
- Authorization logic tests

### Integration Tests  
- API endpoint tests for all CRUD operations
- Authentication and authorization flow tests
- File upload integration tests with Appwrite
- Database transaction tests for rating updates

### Test Data Setup
```go
func setupReviewTestData(db *gorm.DB) {
    // Create test users (customer, seller, admin)
    // Create test products and variants
    // Create test orders with delivered status
    // Create sample reviews with different statuses
}
```

### Performance Considerations

#### Database Optimization
- Index on `product_variant_id` for review queries
- Index on `user_id` for user's review history
- Composite index on `(product_variant_id, status)` for approved reviews
- Pagination for review lists to handle large datasets

#### Caching Strategy
- Cache product rating aggregations
- Cache review counts per product
- Consider Redis for frequently accessed review data

#### Query Optimization
- Use appropriate preloading for related data
- Implement efficient pagination with cursor-based approach for large datasets
- Optimize rating calculation queries with database aggregation functions

### Security Considerations

#### Input Validation
- Sanitize review content to prevent XSS
- Validate file uploads for review images
- Rate limiting on review submission endpoints

#### Authorization
- Verify user ownership for review modifications
- Implement seller verification for response endpoints
- Admin role verification for moderation actions

#### Data Privacy
- Option to anonymize reviewer names
- Secure deletion of review data when requested
- Audit logging for moderation actions