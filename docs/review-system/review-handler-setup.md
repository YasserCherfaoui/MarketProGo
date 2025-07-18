# Review Handler Setup Documentation

## Overview

The Review Handler system has been set up following the established patterns in the MarketProGo codebase. This document describes the handler structure, routing configuration, and middleware setup for the product review system.

## Handler Structure

### ReviewHandler

The main handler struct that manages all review-related HTTP requests:

```go
type ReviewHandler struct {
    db              *gorm.DB
    appwriteService *aw.AppwriteService
}
```

#### Dependencies
- **Database**: GORM database connection for data persistence
- **Appwrite Service**: For handling review image uploads and file management

#### Constructor
```go
func NewReviewHandler(db *gorm.DB, appwriteService *aw.AppwriteService) *ReviewHandler
```

## Routing Structure

### Route Groups

The review system is organized into several route groups based on access levels:

#### 1. Public Routes (No Authentication)
```
GET /api/v1/reviews/:id                    - Get single review by ID
GET /api/v1/reviews/product/:productVariantId - Get reviews for a product variant
```

#### 2. Authenticated Routes (JWT Required)
```
POST   /api/v1/reviews                     - Create new review
PUT    /api/v1/reviews/:id                 - Update own review
DELETE /api/v1/reviews/:id                 - Delete own review
POST   /api/v1/reviews/:id/helpful         - Mark review as helpful/unhelpful
```

#### 3. Seller Routes (Seller Role Required)
```
POST /api/v1/reviews/:id/response          - Create seller response
PUT  /api/v1/reviews/:id/response          - Update seller response
```

#### 4. Admin Routes (Admin Role Required)
```
GET   /api/v1/admin/reviews                - Get all reviews with moderation status
PUT   /api/v1/admin/reviews/:id/moderate   - Moderate review (approve/reject/flag)
DELETE /api/v1/admin/reviews/:id           - Permanently delete review
```

#### 5. Seller Dashboard Routes (Seller Role Required)
```
GET /api/v1/seller/reviews                 - Get reviews for seller's products
```

## Middleware Configuration

### Authentication Middleware

#### AuthMiddleware
- **Purpose**: Validates JWT tokens and extracts user information
- **Function**: `middlewares.AuthMiddleware()`
- **Sets Context Values**:
  - `user`: JWT claims object
  - `user_id`: User ID from token
  - `user_type`: User type (CUSTOMER, VENDOR, ADMIN, etc.)

#### AdminMiddleware
- **Purpose**: Ensures user has admin privileges
- **Function**: `middlewares.AdminMiddleware()`
- **Requirements**: User type must be `Admin`
- **Chain**: Calls `AuthMiddleware()` first, then checks role

#### SellerMiddleware
- **Purpose**: Ensures user has seller privileges
- **Function**: `middlewares.SellerMiddleware()`
- **Requirements**: User type must be `Vendor` or `Admin`
- **Chain**: Calls `AuthMiddleware()` first, then checks role

## Handler Methods

### Current Implementation Status

All handler methods are currently placeholder implementations that return HTTP 501 (Not Implemented) responses. These will be implemented in subsequent tasks:

#### Public Methods
- `GetReview(c *gin.Context)` - Get single review by ID
- `GetProductReviews(c *gin.Context)` - Get paginated reviews for product

#### Customer Methods
- `CreateReview(c *gin.Context)` - Submit new review
- `UpdateReview(c *gin.Context)` - Update own review
- `DeleteReview(c *gin.Context)` - Delete own review
- `MarkReviewHelpful(c *gin.Context)` - Vote helpful/unhelpful

#### Seller Methods
- `CreateSellerResponse(c *gin.Context)` - Respond to review
- `UpdateSellerResponse(c *gin.Context)` - Update response
- `GetSellerReviews(c *gin.Context)` - Get reviews for seller's products

#### Admin Methods
- `GetAllReviews(c *gin.Context)` - Get all reviews with filtering
- `ModerateReview(c *gin.Context)` - Approve/reject/flag review
- `AdminDeleteReview(c *gin.Context)` - Permanently delete review

## Integration Points

### Main Application Setup

The review handler is integrated into the main application in `routes/app_routes.go`:

```go
// Register Review routes
reviewHandler := review.NewReviewHandler(db, appwriteService)
RegisterReviewRoutes(router, reviewHandler)
```

### Database Integration

- **Auto-migration**: Review models are included in database auto-migration
- **Indexes**: Performance indexes are created automatically
- **Relationships**: Proper foreign key relationships with existing models

### File Storage Integration

- **Appwrite Service**: Used for review image uploads
- **File URLs**: Generated through Appwrite storage system
- **Image Validation**: URL validation for review images

## Security Considerations

### Authentication
- All protected routes require valid JWT tokens
- Token validation includes user type verification
- Failed authentication returns proper HTTP 401 responses

### Authorization
- Role-based access control for different user types
- Admin routes restricted to admin users only
- Seller routes accessible to vendors and admins
- Customer routes accessible to all authenticated users

### Input Validation
- All routes will implement proper input validation
- File upload validation for review images
- Content length limits enforced
- SQL injection prevention through GORM

## Error Handling

### Response Format
All responses follow the established API response format:

```json
{
    "status": 200,
    "message": "Success message",
    "data": {...},
    "error": null
}
```

### Error Types
- **401 Unauthorized**: Invalid or missing authentication
- **403 Forbidden**: Insufficient permissions
- **404 Not Found**: Resource not found
- **422 Validation Error**: Invalid input data
- **500 Internal Server Error**: Server-side errors

## Performance Considerations

### Database Optimization
- Proper indexing on frequently queried fields
- Efficient pagination for large review datasets
- Preloading of related data to prevent N+1 queries

### Caching Strategy
- Product rating aggregations can be cached
- Review counts per product can be cached
- Consider Redis for frequently accessed data

## Testing Strategy

### Unit Tests
- Handler method tests with mocked dependencies
- Middleware tests for authentication and authorization
- Validation logic tests

### Integration Tests
- End-to-end API tests for all routes
- Database transaction tests
- File upload integration tests

### Test Data Setup
- Test users with different roles (customer, vendor, admin)
- Test products and variants
- Test orders with delivered status
- Sample reviews with different statuses

## Next Steps

### Implementation Order
1. **Task 3**: Implement purchase verification service
2. **Task 4**: Implement review submission endpoint
3. **Task 5**: Implement rating aggregation system
4. **Task 6**: Implement review display endpoints
5. **Task 7**: Implement review helpfulness system
6. **Task 8**: Implement seller response functionality
7. **Task 9**: Implement customer review management
8. **Task 10**: Implement admin moderation system

### Dependencies
- Database models are ready (Task 1 completed)
- Handler structure is set up (Task 2 completed)
- Routing and middleware are configured
- Integration with main application is complete

## Configuration

### Environment Variables
- Database connection settings
- Appwrite configuration for file storage
- JWT secret for authentication

### Dependencies
- GORM for database operations
- Gin for HTTP routing
- Appwrite SDK for file management
- JWT library for authentication

## Monitoring and Logging

### Logging Strategy
- Request/response logging for debugging
- Error logging for failed operations
- Performance logging for slow queries
- Security logging for authentication failures

### Metrics
- Review submission rates
- Moderation queue size
- Response times for review queries
- Error rates by endpoint 