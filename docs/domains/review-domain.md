# Review Domain

This document covers the Review domain, including product review management, rating aggregation, moderation, seller responses, and helpfulness voting. The review system ensures only verified purchasers can review products and provides comprehensive moderation tools.

---

## Overview

The Review domain manages customer reviews for product variants, including review submission, rating aggregation, moderation workflows, seller responses, and helpfulness voting. The system enforces purchase verification and provides role-based access control for different user types.

---

## Endpoints

### Public Review Access

| Method | Path                           | Description                    | Auth Required |
|--------|--------------------------------|--------------------------------|--------------|
| GET    | /reviews/:id                   | Get single review by ID        | No           |
| GET    | /reviews/product/:variantId    | Get reviews for product variant| No           |

### Customer Review Management

| Method | Path                           | Description                    | Auth Required |
|--------|--------------------------------|--------------------------------|--------------|
| POST   | /reviews                       | Create a new review            | Yes (Customer)|
| PUT    | /reviews/:id                   | Update user's review           | Yes (Owner)  |
| DELETE | /reviews/:id                   | Delete user's review           | Yes (Owner)  |
| GET    | /reviews/user/me               | Get user's own reviews         | Yes (Customer)|
| GET    | /reviews/reviewable-products   | Get products user can review   | Yes (Customer)|
| POST   | /reviews/upload-images         | Upload review images           | Yes (Customer)|

### Review Engagement

| Method | Path                           | Description                    | Auth Required |
|--------|--------------------------------|--------------------------------|--------------|
| POST   | /reviews/:id/helpful           | Mark review as helpful/not     | Yes (Customer)|

### Seller Management

| Method | Path                           | Description                    | Auth Required |
|--------|--------------------------------|--------------------------------|--------------|
| POST   | /reviews/:id/response          | Create seller response         | Yes (Seller) |
| PUT    | /reviews/:id/response          | Update seller response         | Yes (Seller) |
| GET    | /seller/reviews                | Get reviews for seller's products| Yes (Seller)|

### Admin Moderation

| Method | Path                           | Description                    | Auth Required |
|--------|--------------------------------|--------------------------------|--------------|
| GET    | /admin/reviews                 | Get all reviews for moderation | Yes (Admin)  |
| PUT    | /admin/reviews/:id/moderate    | Moderate review status         | Yes (Admin)  |
| DELETE | /admin/reviews/:id             | Admin delete review            | Yes (Admin)  |
| GET    | /admin/reviews/stats           | Get moderation statistics      | Yes (Admin)  |

---

## Review Status Workflow

```
PENDING → APPROVED (visible to public)
     ↓
   REJECTED (hidden from public)
     ↓
   FLAGGED (requires re-review)
```

### Status Descriptions

- **PENDING**: New review awaiting moderation
- **APPROVED**: Review approved and visible to public
- **REJECTED**: Review rejected and hidden from public
- **FLAGGED**: Review flagged for additional review

---

## Request/Response Formats

### Create Review Request

```json
{
  "product_variant_id": 123,
  "rating": 5,
  "title": "Excellent product!",
  "content": "This product exceeded my expectations. Great quality and fast shipping.",
  "images": [
    {
      "url": "https://example.com/image1.jpg",
      "alt_text": "Product in use"
    }
  ]
}
```

### Review Response

```json
{
  "id": 456,
  "product_variant_id": 123,
  "user_id": 789,
  "user": {
    "id": 789,
    "first_name": "John",
    "last_name": "Doe"
  },
  "rating": 5,
  "title": "Excellent product!",
  "content": "This product exceeded my expectations. Great quality and fast shipping.",
  "is_verified_purchase": true,
  "status": "APPROVED",
  "helpful_count": 3,
  "images": [
    {
      "id": 1,
      "url": "https://example.com/image1.jpg",
      "alt_text": "Product in use"
    }
  ],
  "seller_response": {
    "id": 1,
    "content": "Thank you for your review! We're glad you're happy with the product.",
    "user": {
      "id": 101,
      "first_name": "Seller",
      "last_name": "Name"
    },
    "created_at": "2024-01-15T10:30:00Z"
  },
  "created_at": "2024-01-10T14:20:00Z",
  "updated_at": "2024-01-10T14:20:00Z"
}
```

### Product Reviews Response

```json
{
  "reviews": [
    {
      "id": 456,
      "rating": 5,
      "title": "Excellent product!",
      "content": "This product exceeded my expectations.",
      "is_verified_purchase": true,
      "helpful_count": 3,
      "user": {
        "first_name": "John",
        "last_name": "Doe"
      },
      "created_at": "2024-01-10T14:20:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 25,
    "total_pages": 3
  },
  "rating_summary": {
    "average_rating": 4.2,
    "total_reviews": 25,
    "rating_breakdown": {
      "1": 1,
      "2": 2,
      "3": 3,
      "4": 8,
      "5": 11
    }
  }
}
```

### Seller Response Request

```json
{
  "content": "Thank you for your review! We're glad you're happy with the product and appreciate your feedback."
}
```

### Admin Moderation Request

```json
{
  "status": "APPROVED",
  "reason": "Review meets community guidelines"
}
```

---

## Business Rules

### Review Creation
- **Purchase Verification**: Only users who have purchased the product variant can review it
- **One Review Per Purchase**: Users can only review each product variant once per purchase
- **Rating Validation**: Rating must be between 1-5 stars
- **Content Limits**: Title max 100 characters, content max 1000 characters
- **Image Limits**: Maximum 5 images per review

### Review Moderation
- **Auto-Approval**: Reviews from verified purchasers are auto-approved
- **Manual Review**: Flagged reviews require admin approval
- **Status Transitions**: Only admins can change review status
- **Audit Trail**: All moderation actions are logged

### Seller Responses
- **One Response Per Review**: Sellers can only respond once per review
- **Response Limit**: Response content max 500 characters
- **Timing**: Responses can be added anytime after review creation

### Helpfulness Voting
- **One Vote Per User**: Users can only vote once per review
- **Vote Changes**: Users can change their vote
- **Anonymous Voting**: Votes are tracked but not publicly visible

### Rating Aggregation
- **Automatic Updates**: Product ratings update automatically when reviews change
- **Breakdown Tracking**: Maintains count of each rating level (1-5 stars)
- **Average Calculation**: Weighted average based on all approved reviews

---

## Purchase Verification

The system verifies purchase eligibility through:

1. **Order Item Link**: Reviews are linked to specific order items
2. **Order Status Check**: Only completed orders qualify
3. **Product Variant Match**: Must match the exact variant being reviewed
4. **User Ownership**: Order must belong to the reviewing user

### Verification Process

```go
// Pseudo-code for purchase verification
func VerifyPurchase(userID, productVariantID uint) bool {
    orderItem := FindOrderItem(userID, productVariantID)
    return orderItem != nil && 
           orderItem.Order.Status == "COMPLETED" &&
           !HasExistingReview(userID, orderItem.ID)
}
```

---

## Moderation System

### Automatic Moderation
- **Spam Detection**: Filters obvious spam content
- **Profanity Filter**: Blocks inappropriate language
- **Duplicate Detection**: Prevents duplicate reviews

### Manual Moderation
- **Admin Dashboard**: Comprehensive review management interface
- **Bulk Actions**: Process multiple reviews simultaneously
- **Status Management**: Approve, reject, or flag reviews
- **Reason Tracking**: Document moderation decisions

### Moderation Statistics
- **Pending Count**: Number of reviews awaiting moderation
- **Approval Rate**: Percentage of reviews approved
- **Response Time**: Average time to moderate reviews
- **Flagged Reviews**: Reviews requiring attention

---

## Rating Aggregation

### Automatic Updates
The system automatically recalculates product ratings when:

- New review is approved
- Existing review is updated
- Review is deleted or rejected
- Review status changes

### Rating Calculation

```go
// Pseudo-code for rating aggregation
func UpdateProductRating(productVariantID uint) {
    reviews := GetApprovedReviews(productVariantID)
    
    totalRating := 0
    ratingCounts := make(map[int]int)
    
    for _, review := range reviews {
        totalRating += review.Rating
        ratingCounts[review.Rating]++
    }
    
    averageRating := float64(totalRating) / float64(len(reviews))
    
    UpdateProductRatingRecord(productVariantID, averageRating, len(reviews), ratingCounts)
}
```

---

## Referenced Models

- **ProductReview**: Main review entity with moderation and engagement
- **ReviewImage**: Images attached to reviews
- **SellerResponse**: Seller responses to customer reviews
- **ReviewHelpful**: Helpfulness voting tracking
- **ProductRating**: Aggregated rating data for product variants
- **ReviewModerationLog**: Audit trail for moderation actions

See `docs/database/models.md` for complete model definitions.

---

## Middleware

- **AuthMiddleware**: Required for all authenticated endpoints
- **SellerMiddleware**: Required for seller-specific endpoints
- **AdminMiddleware**: Required for admin moderation endpoints
- **PurchaseVerificationMiddleware**: Ensures review eligibility

---

## Security Features

### Access Control
- **Role-Based Access**: Different permissions for customers, sellers, and admins
- **Ownership Validation**: Users can only modify their own reviews
- **Admin Override**: Admins can modify any review

### Data Protection
- **Input Validation**: Comprehensive validation for all inputs
- **SQL Injection Prevention**: Parameterized queries
- **XSS Protection**: Content sanitization
- **Rate Limiting**: Prevents review spam

### Audit Trail
- **Moderation Logs**: Complete history of moderation actions
- **Review History**: Track all changes to reviews
- **User Activity**: Monitor review patterns

---

## Performance Optimizations

### Database Indexing
- **Composite Indexes**: Optimized queries for common patterns
- **Foreign Key Indexes**: Fast relationship lookups
- **Status Indexes**: Efficient moderation filtering

### Caching Strategy
- **Rating Cache**: Cached product ratings for fast display
- **Review Cache**: Frequently accessed reviews cached
- **Statistics Cache**: Moderation statistics cached

### Pagination
- **Efficient Queries**: Limit result sets for large datasets
- **Cursor-Based**: Smooth pagination for review lists
- **Optimized Counts**: Fast total count calculations

---

For more details, see the Go source files in `handlers/review/` and `models/review.go`. 