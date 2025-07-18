# Review Submission System

## Overview

The review submission system allows authenticated users to submit reviews for products they have purchased. The system includes comprehensive validation, purchase verification, image upload capabilities, and duplicate prevention.

## Endpoints

### 1. Create Review

**POST** `/api/v1/reviews`

Creates a new product review for an authenticated user.

#### Authentication
- **Required**: JWT Bearer token
- **User Type**: Customer, Vendor, Wholesaler, Admin

#### Request Body

```json
{
  "product_variant_id": 123,
  "order_item_id": 456,
  "rating": 5,
  "title": "Great product!",
  "content": "This product exceeded my expectations. Highly recommended!",
  "images": [
    "https://example.com/image1.jpg",
    "https://example.com/image2.jpg"
  ]
}
```

#### Field Validation

| Field | Type | Required | Validation Rules |
|-------|------|----------|------------------|
| `product_variant_id` | uint | Yes | Must exist in database |
| `order_item_id` | uint | No | If provided, must belong to user and be eligible for review |
| `rating` | int | Yes | Must be between 1-5 |
| `title` | string | No | Maximum 100 characters |
| `content` | string | Yes | Maximum 1000 characters |
| `images` | []string | No | Maximum 5 images |

#### Business Rules

1. **Purchase Verification**: User must have purchased the product variant
   - Order must be delivered (`DELIVERED` status)
   - Payment must be completed (`PAID` status)
   - Order item must be active
   - Purchase must be within last 2 years

2. **Duplicate Prevention**: User cannot review the same product variant twice

3. **Review Status**: New reviews default to `PENDING` status for moderation

4. **Image Limits**: Maximum 5 images per review

#### Response

**Success (201 Created)**
```json
{
  "status": 201,
  "message": "Review submitted successfully",
  "data": {
    "review": {
      "id": 789,
      "product_variant_id": 123,
      "user_id": 456,
      "rating": 5,
      "title": "Great product!",
      "content": "This product exceeded my expectations. Highly recommended!",
      "is_verified_purchase": true,
      "status": "PENDING",
      "created_at": "2024-01-15T10:30:00Z",
      "images": [
        {
          "id": 1,
          "url": "https://example.com/image1.jpg",
          "alt_text": "Review image 1"
        }
      ],
      "user": {
        "id": 456,
        "first_name": "John",
        "last_name": "Doe"
      },
      "product_variant": {
        "id": 123,
        "name": "Product Variant",
        "sku": "SKU-001"
      }
    }
  }
}
```

**Error Responses**

| Status | Error Code | Description |
|--------|------------|-------------|
| 400 | `review/create` | Invalid request data (validation errors) |
| 400 | `review/create` | Maximum 5 images allowed per review |
| 400 | `review/create` | You have already reviewed this product |
| 401 | `review/create` | User not authenticated |
| 403 | `review/create` | No verified purchase found for this product |
| 403 | `review/create` | Purchase is too old to review (more than 2 years) |
| 404 | `review/create` | Product variant not found |
| 500 | `review/create` | Internal server error |

### 2. Upload Review Images

**POST** `/api/v1/reviews/upload-images`

Uploads images for review submission using multipart form data.

#### Authentication
- **Required**: JWT Bearer token
- **User Type**: Any authenticated user

#### Request

- **Content-Type**: `multipart/form-data`
- **Field Name**: `images` (multiple files)

#### Validation Rules

| Rule | Value |
|------|-------|
| Maximum files | 5 |
| Maximum file size | 5MB per file |
| Allowed formats | JPEG, JPG, PNG, GIF, WebP |
| Total form size | 32MB |

#### Response

**Success (200 OK)**
```json
{
  "status": 200,
  "message": "Images uploaded successfully",
  "data": {
    "images": [
      "/file/preview/file_id_1",
      "/file/preview/file_id_2"
    ]
  }
}
```

**Error Responses**

| Status | Error Code | Description |
|--------|------------|-------------|
| 400 | `review/upload-images` | No images provided |
| 400 | `review/upload-images` | Maximum 5 images allowed |
| 400 | `review/upload-images` | File size too large (max 5MB per file) |
| 400 | `review/upload-images` | Invalid file type. Only images are allowed |
| 401 | `review/upload-images` | User not authenticated |
| 500 | `review/upload-images` | Failed to upload image |

### 3. Get Reviewable Products

**GET** `/api/v1/reviews/reviewable-products`

Returns a list of products the user can review (purchased but not yet reviewed).

#### Authentication
- **Required**: JWT Bearer token
- **User Type**: Any authenticated user

#### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | int | 20 | Maximum number of products to return (max 100) |

#### Response

**Success (200 OK)**
```json
{
  "status": 200,
  "message": "Reviewable products retrieved successfully",
  "data": {
    "products": [
      {
        "order_item_id": 123,
        "product_variant_id": 456,
        "product_variant": {
          "id": 456,
          "name": "Product Variant",
          "sku": "SKU-001",
          "base_price": 100.0
        },
        "product": {
          "id": 789,
          "name": "Product Name",
          "description": "Product description"
        },
        "order": {
          "id": 101,
          "order_number": "ORD-001",
          "delivered_date": "2024-01-10T15:30:00Z"
        },
        "purchased_date": "2024-01-10T15:30:00Z",
        "quantity": 2,
        "unit_price": 100.0
      }
    ],
    "count": 1
  }
}
```

**Error Responses**

| Status | Error Code | Description |
|--------|------------|-------------|
| 401 | `review/reviewable-products` | User not authenticated |
| 500 | `review/reviewable-products` | Failed to get reviewable products |

## Implementation Details

### Purchase Verification Logic

The system verifies purchases using two methods:

1. **Product Variant Verification**: Checks if user has any delivered orders for the specific product variant
2. **Order Item Verification**: Verifies a specific order item belongs to the user and is eligible

```go
// Verify purchase using product variant
purchaseResult, err := handler.VerifyPurchase(userID, productVariantID)

// Verify purchase using order item
purchaseResult, err := handler.VerifyPurchaseWithOrderItemID(userID, orderItemID)
```

### Image Upload Process

1. **File Validation**: Checks file type, size, and count
2. **Appwrite Storage**: Uploads files to Appwrite storage service
3. **URL Generation**: Returns public URLs for uploaded images
4. **Database Storage**: Stores image URLs in `review_images` table

### Database Schema

#### ProductReview Table
```sql
CREATE TABLE product_reviews (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  product_variant_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  order_item_id BIGINT,
  rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
  title VARCHAR(100),
  content VARCHAR(1000) NOT NULL,
  is_verified_purchase BOOLEAN DEFAULT TRUE,
  status VARCHAR(20) DEFAULT 'PENDING',
  moderated_by BIGINT,
  moderated_at TIMESTAMP,
  moderation_reason VARCHAR(500),
  helpful_count INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  
  INDEX idx_product_variant (product_variant_id),
  INDEX idx_user (user_id),
  INDEX idx_order_item (order_item_id),
  UNIQUE KEY unique_user_product (user_id, product_variant_id)
);
```

#### ReviewImage Table
```sql
CREATE TABLE review_images (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  product_review_id BIGINT NOT NULL,
  url VARCHAR(500) NOT NULL,
  alt_text VARCHAR(100),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  
  INDEX idx_product_review (product_review_id)
);
```

## Security Considerations

1. **Authentication**: All endpoints require valid JWT tokens
2. **Authorization**: Users can only review products they have purchased
3. **Input Validation**: Comprehensive validation of all input fields
4. **File Upload Security**: File type and size restrictions
5. **SQL Injection Prevention**: Parameterized queries via GORM
6. **XSS Prevention**: Input sanitization and output encoding

## Error Handling

The system uses consistent error responses with:
- HTTP status codes
- Error codes for client handling
- Descriptive error messages
- Structured JSON responses

## Testing

Comprehensive unit tests cover:
- Valid review creation scenarios
- Purchase verification edge cases
- Validation error handling
- Duplicate review prevention
- Image upload functionality
- Error response formats

## Performance Considerations

1. **Database Indexes**: Optimized indexes on frequently queried fields
2. **Eager Loading**: Preloads related data to minimize N+1 queries
3. **File Upload Limits**: Reasonable limits to prevent abuse
4. **Caching**: Consider caching for frequently accessed review data

## Integration Points

1. **Appwrite Storage**: For image file storage
2. **Order System**: For purchase verification
3. **Product System**: For product variant validation
4. **User System**: For authentication and user data
5. **Moderation System**: For review approval workflow

## Future Enhancements

1. **Review Templates**: Predefined review templates for common scenarios
2. **Review Analytics**: Track review submission patterns
3. **Bulk Operations**: Allow multiple review submissions
4. **Review Drafts**: Save review drafts before submission
5. **Review Scheduling**: Schedule reviews for later submission 