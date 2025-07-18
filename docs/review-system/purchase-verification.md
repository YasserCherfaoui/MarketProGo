# Purchase Verification Service Documentation

## Overview

The Purchase Verification Service is a critical component of the Product Review System that ensures only customers who have actually purchased and received products can submit reviews. This service validates purchase history, enforces business rules, and prevents fake or unauthorized reviews.

## Core Functionality

### PurchaseVerificationResult

The service returns a structured result containing verification status and related data:

```go
type PurchaseVerificationResult struct {
    IsVerified     bool              `json:"is_verified"`
    OrderItem      *models.OrderItem `json:"order_item,omitempty"`
    Order          *models.Order     `json:"order,omitempty"`
    ProductVariant *models.ProductVariant `json:"product_variant,omitempty"`
    ErrorMessage   string            `json:"error_message,omitempty"`
}
```

## Business Rules

### Purchase Verification Criteria

A purchase is considered valid for review if ALL of the following conditions are met:

1. **Order Status**: Order must be `DELIVERED`
2. **Payment Status**: Order must be `PAID`
3. **Order Item Status**: Order item must be `active`
4. **Delivery Date**: Order must have been delivered within the last 2 years
5. **User Ownership**: Order must belong to the requesting user
6. **Product Match**: Order item must match the specific product variant

### Time-Based Restrictions

- **Review Window**: Customers can only review products purchased within the last 2 years
- **Reasoning**: Ensures reviews are based on recent experiences and current product quality
- **Exception Handling**: Orders delivered more than 2 years ago are automatically rejected

### Duplicate Prevention

- **One Review Per Purchase**: Users can only submit one review per product variant
- **Existing Review Check**: System prevents duplicate reviews for the same product variant
- **Reviewable Products**: Service can identify products that can still be reviewed

## Service Methods

### VerifyPurchase

**Purpose**: Verifies if a user has purchased a specific product variant and can review it.

**Signature**:
```go
func (h *ReviewHandler) VerifyPurchase(userID uint, productVariantID uint) (*PurchaseVerificationResult, error)
```

**Logic**:
1. Queries all delivered, paid, active order items for the user and product variant
2. Orders results by delivery date (most recent first)
3. Returns the first purchase within the 2-year window
4. If no valid purchase found, returns appropriate error message

**Usage Example**:
```go
result, err := handler.VerifyPurchase(userID, productVariantID)
if err != nil {
    // Handle database error
    return err
}

if !result.IsVerified {
    // User cannot review this product
    return fmt.Errorf("purchase verification failed: %s", result.ErrorMessage)
}

// User can review - proceed with review creation
orderItemID := result.OrderItem.ID
```

### VerifyPurchaseWithOrderItemID

**Purpose**: Verifies a specific order item belongs to the user and is eligible for review.

**Signature**:
```go
func (h *ReviewHandler) VerifyPurchaseWithOrderItemID(userID uint, orderItemID uint) (*PurchaseVerificationResult, error)
```

**Use Cases**:
- When user provides a specific order item ID for review
- For review systems that link reviews to specific purchases
- Additional validation for review authenticity

### CheckIfUserCanReview

**Purpose**: Comprehensive check that combines purchase verification with duplicate review prevention.

**Signature**:
```go
func (h *ReviewHandler) CheckIfUserCanReview(userID uint, productVariantID uint) (*PurchaseVerificationResult, error)
```

**Logic**:
1. Verifies purchase using `VerifyPurchase`
2. Checks if user has already reviewed this product variant
3. Returns detailed result with appropriate error messages

**Error Messages**:
- "No verified purchase found for this product"
- "Purchase is too old to review (more than 2 years)"
- "You have already reviewed this product"

### GetUserPurchasedProducts

**Purpose**: Retrieves all products a user has purchased and can potentially review.

**Signature**:
```go
func (h *ReviewHandler) GetUserPurchasedProducts(userID uint, limit int) ([]models.OrderItem, error)
```

**Features**:
- Returns products ordered by delivery date (most recent first)
- Optional limit parameter for pagination
- Includes full order and product variant data
- Only includes delivered, paid, active orders

### GetReviewableProductsForUser

**Purpose**: Returns products the user can still review (purchased but not yet reviewed).

**Signature**:
```go
func (h *ReviewHandler) GetReviewableProductsForUser(userID uint, limit int) ([]models.OrderItem, error)
```

**Logic**:
1. Gets all purchased products using `GetUserPurchasedProducts`
2. Filters out products that have already been reviewed
3. Returns only products eligible for new reviews

## Database Queries

### Core Verification Query

The service uses optimized database queries with proper joins and indexing:

```sql
SELECT order_items.*, orders.*, product_variants.*
FROM order_items 
JOIN orders ON orders.id = order_items.order_id
JOIN product_variants ON product_variants.id = order_items.product_variant_id
WHERE orders.user_id = ? 
  AND order_items.product_variant_id = ? 
  AND orders.status = 'DELIVERED' 
  AND orders.payment_status = 'PAID'
  AND order_items.status = 'active'
ORDER BY orders.delivered_date DESC
```

### Performance Optimizations

- **Indexes**: Leverages existing indexes on `user_id`, `product_variant_id`, `status`
- **Preloading**: Uses GORM preloading to avoid N+1 queries
- **Ordering**: Orders by delivery date to get most recent purchases first
- **Filtering**: Applies all filters at database level for efficiency

## Error Handling

### Database Errors

- **Connection Issues**: Returns wrapped error with context
- **Query Failures**: Provides detailed error messages
- **Transaction Rollback**: Ensures data consistency

### Business Logic Errors

- **No Purchase Found**: Clear message explaining why verification failed
- **Purchase Too Old**: Explains the 2-year restriction
- **Already Reviewed**: Prevents duplicate reviews with helpful message

### Error Response Format

```json
{
    "is_verified": false,
    "error_message": "Purchase is too old to review (more than 2 years)",
    "order_item": null,
    "order": null,
    "product_variant": null
}
```

## Integration Points

### Review Creation Flow

1. **User submits review** for a product variant
2. **Purchase verification** is called automatically
3. **If verified**: Review is created with `OrderItemID` and `IsVerifiedPurchase = true`
4. **If not verified**: Review submission is rejected with clear error message

### Frontend Integration

- **Review Form**: Can pre-populate with user's purchased products
- **Product Pages**: Can show "Review this product" button only for verified purchases
- **User Dashboard**: Can display reviewable products for easy access

### Admin Interface

- **Moderation**: Admins can see purchase verification status
- **Audit Trail**: Full purchase history linked to reviews
- **Fraud Detection**: Identify patterns in purchase verification

## Security Considerations

### Data Privacy

- **User Isolation**: Users can only verify their own purchases
- **Order Privacy**: Order details are only accessible to the purchasing user
- **Review Anonymity**: Purchase verification doesn't expose sensitive order data

### Fraud Prevention

- **Purchase Verification**: Prevents fake reviews from non-customers
- **Time Restrictions**: Prevents reviews of outdated products
- **Duplicate Prevention**: Ensures one review per purchase
- **Audit Trail**: Maintains complete verification history

### Authorization

- **User Ownership**: Verifies order belongs to requesting user
- **Role-Based Access**: Different verification levels for different user types
- **API Security**: All endpoints require proper authentication

## Testing Strategy

### Unit Tests

The service includes comprehensive unit tests covering:

- **Valid Purchase**: User has recent, delivered, paid purchase
- **No Purchase**: User has never purchased the product
- **Undelivered Order**: Order exists but not yet delivered
- **Old Purchase**: Purchase is more than 2 years old
- **Duplicate Review**: User has already reviewed the product
- **Multiple Purchases**: User has multiple purchases of same product

### Test Data Setup

```go
// Create test scenarios
user := createTestUser(db, models.Customer)
product := createTestProduct(db)
variant := createTestProductVariant(db, product.ID)

// Valid purchase (1 month ago)
deliveredDate := time.Now().AddDate(0, -1, 0)
order := createTestOrder(db, user.ID, models.OrderStatusDelivered, &deliveredDate)
orderItem := createTestOrderItem(db, order.ID, variant.ID)

// Test verification
result, err := handler.VerifyPurchase(user.ID, variant.ID)
assert.True(t, result.IsVerified)
```

### Edge Cases

- **Null Delivery Dates**: Handles orders without delivery dates
- **Multiple Orders**: Prioritizes most recent valid purchase
- **Cancelled Orders**: Excludes cancelled or returned items
- **Payment Issues**: Only considers fully paid orders

## Performance Considerations

### Query Optimization

- **Indexed Fields**: All query fields are properly indexed
- **Efficient Joins**: Uses optimized JOIN operations
- **Result Limiting**: Supports pagination for large datasets
- **Caching**: Can be extended with Redis caching for frequent queries

### Scalability

- **Database Sharding**: Can be extended for multi-tenant setups
- **Read Replicas**: Can use read replicas for verification queries
- **Async Processing**: Can be made asynchronous for high-volume scenarios

## Future Enhancements

### Planned Features

- **Review Window Configuration**: Configurable time limits per product category
- **Purchase Verification Levels**: Different verification requirements for different products
- **Bulk Verification**: Batch verification for multiple products
- **Analytics Integration**: Track verification patterns and success rates

### API Extensions

- **Verification History**: Track all verification attempts
- **Verification Statistics**: Provide insights into verification patterns
- **Webhook Support**: Notify external systems of verification events
- **Rate Limiting**: Prevent abuse of verification endpoints 