# Product Review System - Models Documentation

## Overview

The Product Review System provides comprehensive review functionality for the Algeria Market platform, allowing customers to submit reviews for products they have purchased, sellers to respond to reviews, and administrators to moderate content. The system includes purchase verification, rating aggregation, helpfulness tracking, and image support.

## Core Models

### ProductReview

The main review entity that represents a customer's review for a specific product variant.

```go
type ProductReview struct {
    gorm.Model
    ProductVariantID uint           `json:"product_variant_id" gorm:"index"`
    ProductVariant   ProductVariant `json:"product_variant"`
    UserID           uint           `json:"user_id" gorm:"index"`
    User             User           `json:"user"`
    OrderItemID      *uint          `json:"order_item_id" gorm:"index"`
    OrderItem        *OrderItem     `json:"order_item,omitempty"`

    Rating             int    `json:"rating" validate:"required,min=1,max=5"`
    Title              string `json:"title" validate:"max=100"`
    Content            string `json:"content" validate:"max=1000"`
    IsVerifiedPurchase bool   `json:"is_verified_purchase"`

    // Moderation
    Status           ReviewStatus `json:"status" gorm:"type:varchar(20);default:'PENDING'"`
    ModeratedBy      *uint        `json:"moderated_by"`
    ModeratedAt      *time.Time   `json:"moderated_at"`
    ModerationReason string       `json:"moderation_reason" validate:"max=500"`

    // Engagement
    HelpfulCount int `json:"helpful_count" gorm:"default:0"`

    // Relationships
    Images         []ReviewImage   `json:"images" gorm:"foreignKey:ProductReviewID"`
    SellerResponse *SellerResponse `json:"seller_response,omitempty" gorm:"foreignKey:ProductReviewID"`
    HelpfulVotes   []ReviewHelpful `json:"-" gorm:"foreignKey:ProductReviewID"`
}
```

#### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | uint | Auto | Primary key (auto-generated) |
| `created_at` | time.Time | Auto | Review submission timestamp |
| `updated_at` | time.Time | Auto | Last modification timestamp |
| `deleted_at` | *time.Time | Auto | Soft delete timestamp |
| `product_variant_id` | uint | Yes | Reference to the product variant being reviewed |
| `user_id` | uint | Yes | Reference to the user who wrote the review |
| `order_item_id` | *uint | No | Reference to the order item for purchase verification |
| `rating` | int | Yes | Rating from 1-5 stars |
| `title` | string | No | Review title (max 100 characters) |
| `content` | string | No | Review content (max 1000 characters) |
| `is_verified_purchase` | bool | Auto | Set to true if review is linked to a delivered order |
| `status` | ReviewStatus | Auto | Moderation status (PENDING, APPROVED, REJECTED, FLAGGED) |
| `moderated_by` | *uint | No | Admin user ID who moderated the review |
| `moderated_at` | *time.Time | No | Timestamp when review was moderated |
| `moderation_reason` | string | No | Reason for rejection/flagging (max 500 characters) |
| `helpful_count` | int | Auto | Count of helpful votes |

#### Business Logic Methods

- `IsApproved()` - Returns true if review is approved and visible
- `IsPending()` - Returns true if review is pending moderation
- `IsRejected()` - Returns true if review was rejected
- `IsFlagged()` - Returns true if review was flagged
- `CanBeModifiedBy(userID, userType)` - Checks if user can modify review
- `CanBeDeletedBy(userID, userType)` - Checks if user can delete review
- `HasSellerResponse()` - Returns true if review has seller response
- `GetReviewerName()` - Returns formatted reviewer name

### ReviewStatus

Enumeration for review moderation status.

```go
type ReviewStatus string

const (
    ReviewStatusPending  ReviewStatus = "PENDING"
    ReviewStatusApproved ReviewStatus = "APPROVED"
    ReviewStatusRejected ReviewStatus = "REJECTED"
    ReviewStatusFlagged  ReviewStatus = "FLAGGED"
)
```

### ReviewImage

Represents images attached to a product review.

```go
type ReviewImage struct {
    gorm.Model
    ProductReviewID uint   `json:"product_review_id" gorm:"index"`
    URL             string `json:"url" validate:"required,url"`
    AltText         string `json:"alt_text" validate:"max=100"`
}
```

#### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | uint | Auto | Primary key |
| `product_review_id` | uint | Yes | Reference to the review |
| `url` | string | Yes | Image URL (must be valid URL) |
| `alt_text` | string | No | Alt text for accessibility (max 100 characters) |

### SellerResponse

Represents a seller's response to a customer review.

```go
type SellerResponse struct {
    gorm.Model
    ProductReviewID uint   `json:"product_review_id" gorm:"uniqueIndex"`
    UserID          uint   `json:"user_id" gorm:"index"`
    User            User   `json:"user"`
    Content         string `json:"content" validate:"required,max=500"`
}
```

#### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | uint | Auto | Primary key |
| `product_review_id` | uint | Yes | Reference to the review (unique) |
| `user_id` | uint | Yes | Seller user ID |
| `content` | string | Yes | Response content (max 500 characters) |

**Note:** The `uniqueIndex` on `product_review_id` ensures only one response per review.

### ReviewHelpful

Tracks helpful/unhelpful votes from users on reviews.

```go
type ReviewHelpful struct {
    gorm.Model
    ProductReviewID uint `json:"product_review_id" gorm:"index"`
    UserID          uint `json:"user_id" gorm:"index"`
    IsHelpful       bool `json:"is_helpful"`
}
```

#### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | uint | Auto | Primary key |
| `product_review_id` | uint | Yes | Reference to the review |
| `user_id` | uint | Yes | User who voted |
| `is_helpful` | bool | Yes | True for helpful, false for unhelpful |

**Note:** A composite unique index on `(product_review_id, user_id)` prevents duplicate votes.

### ProductRating

Stores aggregated rating data for product variants.

```go
type ProductRating struct {
    gorm.Model
    ProductVariantID uint    `json:"product_variant_id" gorm:"uniqueIndex"`
    AverageRating    float64 `json:"average_rating" gorm:"type:decimal(3,1);default:0.0"`
    TotalReviews     int     `json:"total_reviews" gorm:"default:0"`
    RatingBreakdown  string  `json:"rating_breakdown"`
}
```

#### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | uint | Auto | Primary key |
| `product_variant_id` | uint | Yes | Reference to product variant (unique) |
| `average_rating` | float64 | Auto | Average rating rounded to 1 decimal place |
| `total_reviews` | int | Auto | Total number of approved reviews |
| `rating_breakdown` | string | Auto | JSON string with count per rating: `{"1":0,"2":1,"3":2,"4":5,"5":10}` |

## Database Relationships

```
ProductVariant (1) ←→ (N) ProductReview
User (1) ←→ (N) ProductReview  
OrderItem (1) ←→ (0..1) ProductReview
ProductReview (1) ←→ (N) ReviewImage
ProductReview (1) ←→ (0..1) SellerResponse
ProductReview (1) ←→ (N) ReviewHelpful
ProductVariant (1) ←→ (1) ProductRating
```

## Validation Rules

### ProductReview
- `rating`: Required, integer between 1-5
- `title`: Optional, maximum 100 characters
- `content`: Optional, maximum 1000 characters
- `moderation_reason`: Optional, maximum 500 characters

### ReviewImage
- `url`: Required, must be valid URL format
- `alt_text`: Optional, maximum 100 characters

### SellerResponse
- `content`: Required, maximum 500 characters

## Database Indexes

### Performance Indexes
- `idx_product_reviews_variant_status` - Composite index on `(product_variant_id, status)` for efficient product review queries
- `idx_product_reviews_user_status` - Composite index on `(user_id, status)` for user review history
- `idx_product_reviews_order_item` - Index on `order_item_id` for purchase verification queries
- `idx_seller_responses_user` - Index on `user_id` for seller dashboard queries

### Constraint Indexes
- `idx_review_helpful_unique_vote` - Unique composite index on `(product_review_id, user_id)` to prevent duplicate votes
- `uniqueIndex` on `SellerResponse.product_review_id` - Ensures one response per review
- `uniqueIndex` on `ProductRating.product_variant_id` - Ensures one rating record per variant

## Business Rules

### Review Submission
1. User must have purchased the product variant (verified through OrderItem)
2. Only one review per user per product variant
3. Reviews start with PENDING status
4. Rating must be between 1-5 stars
5. Content and title are optional but have character limits

### Purchase Verification
- Reviews are linked to OrderItem records from delivered orders
- `IsVerifiedPurchase` is automatically set when OrderItemID is provided
- Reviews without purchase verification may be flagged for moderation

### Moderation Workflow
1. **PENDING**: New reviews awaiting moderation
2. **APPROVED**: Reviews visible to public
3. **REJECTED**: Reviews hidden with rejection reason
4. **FLAGGED**: Reviews temporarily hidden for review

### Helpfulness System
- Users can vote helpful/unhelpful on reviews
- One vote per user per review (enforced by unique index)
- Helpful count is automatically updated
- Votes can be changed by updating the existing record

### Rating Aggregation
- Average rating is calculated from approved reviews only
- Rating breakdown tracks count per star rating (1-5)
- Aggregation is updated when reviews are added, modified, or deleted
- Decimal precision is limited to 1 decimal place

### Seller Responses
- Only one response per review (enforced by unique constraint)
- Responses are limited to 500 characters
- Seller must own the product being reviewed
- Responses can be updated by the original responder

## Security Considerations

### Authorization
- Users can only modify their own reviews
- Admins can modify any review
- Sellers can only respond to reviews of their products
- Purchase verification prevents fake reviews

### Data Integrity
- Soft deletes maintain referential integrity
- Unique constraints prevent duplicate data
- Foreign key relationships ensure data consistency
- Validation rules prevent invalid data

### Privacy
- Reviewer names can be anonymized if needed
- Moderation actions are logged with admin user ID
- Review content is validated to prevent XSS

## Performance Considerations

### Query Optimization
- Indexes on frequently queried fields
- Composite indexes for common query patterns
- Efficient pagination for large review datasets
- Preloading of related data to reduce N+1 queries

### Caching Strategy
- Product rating aggregations can be cached
- Review counts per product can be cached
- Consider Redis for frequently accessed review data

### Database Optimization
- Appropriate field types and sizes
- Efficient indexing strategy
- Soft deletes for data retention
- Proper foreign key relationships 