# Wishlist Feature Documentation

## Overview

The wishlist feature allows users to save products they're interested in for future reference. Users can add products to their wishlist, manage items, set priorities, add notes, and control visibility.

## Data Models

### Wishlist
```go
type Wishlist struct {
    gorm.Model
    UserID *uint      `json:"user_id"`
    User   *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Items  []WishlistItem `json:"items"`
}
```

### WishlistItem
```go
type WishlistItem struct {
    gorm.Model
    WishlistID uint  `json:"wishlist_id"`
    Wishlist   *Wishlist `json:"-" gorm:"foreignKey:WishlistID"`

    // Product variant reference
    ProductVariantID uint            `json:"product_variant_id"`
    ProductVariant   *ProductVariant `json:"product_variant" gorm:"foreignKey:ProductVariantID"`

    // Legacy field for backward compatibility
    ProductID *uint    `json:"product_id,omitempty"`
    Product   *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`

    // Additional metadata
    Notes     string `json:"notes"`     // User notes about the item
    Priority  int    `json:"priority"`  // Priority level (1-5, 5 being highest)
    IsPublic  bool   `json:"is_public"` // Whether the item is visible to others
}
```

## API Endpoints

### Base URL
All wishlist endpoints are prefixed with `/api/v1/wishlist`

### Authentication
All endpoints require authentication via the `AuthMiddleware()`.

### 1. Get User's Wishlist
**GET** `/api/v1/wishlist`

Retrieves the authenticated user's wishlist with all items.

**Response:**
```json
{
  "status": 200,
  "message": "Wishlist retrieved successfully",
  "data": {
    "id": 1,
    "user_id": 123,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "items": [
      {
        "id": 1,
        "wishlist_id": 1,
        "product_variant_id": 456,
        "notes": "Need this for birthday",
        "priority": 5,
        "is_public": true,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z",
        "product_variant": {
          "id": 456,
          "name": "Red T-Shirt Large",
          "base_price": 29.99,
          "product": {
            "id": 123,
            "name": "Cotton T-Shirt",
            "description": "Comfortable cotton t-shirt",
            "brand": {
              "id": 1,
              "name": "Fashion Brand"
            },
            "images": [...]
          }
        }
      }
    ]
  }
}
```

### 2. Add Item to Wishlist
**POST** `/api/v1/wishlist/items`

Adds a product variant to the user's wishlist.

**Request Body:**
```json
{
  "product_variant_id": 456,
  "product_id": 123,
  "notes": "Need this for birthday",
  "priority": 5,
  "is_public": true
}
```

**Fields:**
- `product_variant_id` (required): ID of the product variant to add
- `product_id` (optional): Legacy product ID for backward compatibility
- `notes` (optional): User notes about the item
- `priority` (optional): Priority level 1-5 (defaults to 3)
- `is_public` (optional): Whether the item is visible to others (defaults to false)

**Response:**
```json
{
  "status": 201,
  "message": "Item added to wishlist successfully",
  "data": {
    "id": 1,
    "wishlist_id": 1,
    "product_variant_id": 456,
    "notes": "Need this for birthday",
    "priority": 5,
    "is_public": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "product_variant": {
      "id": 456,
      "name": "Red T-Shirt Large",
      "base_price": 29.99,
      "product": {
        "id": 123,
        "name": "Cotton T-Shirt"
      }
    }
  }
}
```

**Error Responses:**
- `400`: Invalid request data
- `401`: User not authenticated
- `404`: Product variant not found
- `409`: Item already exists in wishlist
- `500`: Internal server error

### 3. Update Wishlist Item
**PUT** `/api/v1/wishlist/items/:id`

Updates the metadata of a wishlist item.

**URL Parameters:**
- `id`: Wishlist item ID

**Request Body:**
```json
{
  "notes": "Updated notes",
  "priority": 4,
  "is_public": false
}
```

**Fields:**
- `notes` (optional): Updated user notes
- `priority` (optional): Updated priority level 1-5
- `is_public` (optional): Updated visibility setting

**Response:**
```json
{
  "status": 200,
  "message": "Wishlist item updated successfully",
  "data": {
    "id": 1,
    "wishlist_id": 1,
    "product_variant_id": 456,
    "notes": "Updated notes",
    "priority": 4,
    "is_public": false,
    "updated_at": "2024-01-01T00:00:00Z",
    "product_variant": {
      "id": 456,
      "name": "Red T-Shirt Large",
      "product": {
        "id": 123,
        "name": "Cotton T-Shirt"
      }
    }
  }
}
```

**Error Responses:**
- `400`: Invalid request data or item ID
- `401`: User not authenticated
- `404`: Wishlist or item not found
- `500`: Internal server error

### 4. Remove Item from Wishlist
**DELETE** `/api/v1/wishlist/items/:id`

Removes an item from the user's wishlist.

**URL Parameters:**
- `id`: Wishlist item ID

**Response:**
```json
{
  "status": 200,
  "message": "Item removed from wishlist successfully",
  "data": null
}
```

**Error Responses:**
- `400`: Invalid item ID
- `401`: User not authenticated
- `404`: Wishlist or item not found
- `500`: Internal server error

## Database Schema

### Tables

#### wishlists
- `id` (primary key)
- `user_id` (foreign key to users.id)
- `created_at`
- `updated_at`
- `deleted_at`

#### wishlist_items
- `id` (primary key)
- `wishlist_id` (foreign key to wishlists.id)
- `product_variant_id` (foreign key to product_variants.id)
- `product_id` (foreign key to products.id, nullable)
- `notes` (text)
- `priority` (integer, 1-5)
- `is_public` (boolean)
- `created_at`
- `updated_at`
- `deleted_at`

### Indexes
- `idx_wishlists_user_id` on wishlists(user_id)
- `idx_wishlist_items_wishlist_id` on wishlist_items(wishlist_id)
- `idx_wishlist_items_product_variant_id` on wishlist_items(product_variant_id)
- `idx_wishlist_items_priority` on wishlist_items(priority)

## Usage Examples

### Frontend Integration

#### Add to Wishlist Button
```javascript
const addToWishlist = async (productVariantId) => {
  try {
    const response = await fetch('/api/v1/wishlist/items', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        product_variant_id: productVariantId,
        notes: 'Interested in this product',
        priority: 4,
        is_public: true
      })
    });
    
    if (response.ok) {
      showNotification('Added to wishlist!');
    }
  } catch (error) {
    console.error('Error adding to wishlist:', error);
  }
};
```

#### Display Wishlist
```javascript
const getWishlist = async () => {
  try {
    const response = await fetch('/api/v1/wishlist', {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (response.ok) {
      const data = await response.json();
      displayWishlistItems(data.data.items);
    }
  } catch (error) {
    console.error('Error fetching wishlist:', error);
  }
};
```

#### Remove from Wishlist
```javascript
const removeFromWishlist = async (itemId) => {
  try {
    const response = await fetch(`/api/v1/wishlist/items/${itemId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (response.ok) {
      showNotification('Removed from wishlist');
      refreshWishlist();
    }
  } catch (error) {
    console.error('Error removing from wishlist:', error);
  }
};
```

## Features

### Priority System
- Users can set priority levels from 1-5
- Higher numbers indicate higher priority
- Useful for organizing wishlist items

### Notes
- Users can add personal notes to wishlist items
- Helpful for remembering why they added the item

### Public/Private Visibility
- Users can control whether wishlist items are visible to others
- Enables social features and sharing

### Product Variant Support
- Supports product variants for specific configurations
- Maintains backward compatibility with legacy product IDs

### Automatic Wishlist Creation
- User's wishlist is automatically created on first item addition
- No need for explicit wishlist creation

## Security Considerations

1. **Authentication Required**: All endpoints require valid authentication
2. **User Isolation**: Users can only access their own wishlist
3. **Input Validation**: All inputs are validated and sanitized
4. **SQL Injection Protection**: Uses parameterized queries via GORM

## Performance Considerations

1. **Database Indexes**: Optimized indexes for common queries
2. **Eager Loading**: Product and variant data is preloaded
3. **Soft Deletes**: Uses soft deletes for data recovery
4. **Efficient Queries**: Optimized database queries for wishlist operations

## Future Enhancements

1. **Wishlist Sharing**: Share wishlists with friends/family
2. **Wishlist Analytics**: Track wishlist performance and conversions
3. **Price Alerts**: Notify users when wishlist items go on sale
4. **Wishlist Recommendations**: Suggest similar products
5. **Bulk Operations**: Add/remove multiple items at once
6. **Wishlist Categories**: Organize items into categories
7. **Wishlist Export**: Export wishlist data
8. **Public Wishlists**: Browse public wishlists from other users 