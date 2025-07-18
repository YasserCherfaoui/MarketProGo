# Task 6: Review Display Endpoints - Documentation

## Overview
Implements public endpoints for retrieving product reviews, supporting pagination, filtering, and inclusion of related data (images, seller responses, etc.).

## Endpoints

### 1. Get Reviews for a Product Variant
- **Route:** `GET /api/v1/reviews/product/:productVariantId`
- **Access:** Public
- **Query Parameters:**
  - `page` (default: 1)
  - `limit` (default: 10, max: 50)
  - `sort` (created_at, rating, helpful_count, updated_at)
  - `order` (asc, desc)
  - `rating` (1-5, optional filter)
- **Response:**
  - Paginated list of approved reviews for the product variant
  - Each review includes: user info, rating, title, content, images, helpful count, seller response (if any), timestamps
  - Pagination info: page, limit, total, total_pages, has_next, has_prev
  - Filters used
  - Rating statistics (average, total reviews, breakdown)

### 2. Get Single Review by ID
- **Route:** `GET /api/v1/reviews/:id`
- **Access:** Public
- **Response:**
  - Full review details (same as above) for a single review by ID
  - Only approved reviews are accessible

## Business Rules
- Only reviews with `status = APPROVED` are returned
- Pagination defaults to 10 per page, max 50
- Sorting and filtering are validated against allowed fields/values
- Seller response is included if present
- Images are included for each review
- Rating statistics are included in product review list endpoint

## Data Structure

### Review Object
```json
{
  "id": 123,
  "product_variant_id": 456,
  "user": {
    "id": 1,
    "first_name": "",
    "last_name": "",
    "name": "Anonymous"
  },
  "rating": 5,
  "title": "Great product!",
  "content": "Excellent quality",
  "is_verified_purchase": true,
  "helpful_count": 0,
  "images": [
    { "url": "...", "alt_text": "..." }
  ],
  "seller_response": {
    "id": 789,
    "content": "Thank you!",
    "user": { "id": 2, "first_name": "", "last_name": "", "name": "Anonymous" },
    "created_at": "...",
    "updated_at": "..."
  },
  "created_at": "...",
  "updated_at": "..."
}
```

### Product Review List Response
```json
{
  "success": true,
  "data": {
    "reviews": [ /* array of review objects */ ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 5,
      "total_pages": 1,
      "has_next": false,
      "has_prev": false
    },
    "filters": {
      "rating": "",
      "sort": "created_at",
      "order": "desc"
    },
    "rating_stats": {
      "average_rating": 4.2,
      "total_reviews": 5,
      "rating_breakdown": { "1": 0, "2": 1, "3": 1, "4": 1, "5": 2 }
    }
  }
}
```

## Error Handling
- Returns 400 for invalid IDs, filters, or pagination params
- Returns 404 for not found or non-approved reviews
- Returns 500 for database errors
- Uses consistent error response format via `utils/response`

## Testing
- Comprehensive unit tests for:
  - Success (single and multiple reviews)
  - Pagination, sorting, filtering
  - Edge cases (no reviews, invalid params, not found)
  - Seller response and images inclusion
  - Error scenarios (invalid ID, DB errors)
- All tests pass and cover all business rules

## Security
- Public endpoints, but only approved reviews are accessible
- No sensitive user data is exposed
- SQL injection prevented via GORM parameterization

## Status
**Complete**

---

**Next Phase:** Task 7 - Review Helpfulness System 