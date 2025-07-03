# Address Management System Documentation

## Overview

The Address Management System provides comprehensive address management capabilities for users in the Algeria Market platform. It allows users to manage multiple shipping and billing addresses with support for default address designation, CRUD operations, and address validation.

## Features

- **Multiple Address Support**: Users can manage multiple addresses
- **Default Address**: Automatic management of default address designation
- **Address Validation**: Required field validation for address components
- **User-Scoped Access**: Addresses are private to each user
- **Order Integration**: Addresses can be used for shipping in orders
- **Warehouse Integration**: Addresses can be assigned to warehouses

## Database Schema

### Address Model

The `Address` model represents a physical address for users, warehouses, or companies.

```go
type Address struct {
    gorm.Model
    StreetAddress1 string `gorm:"not null" json:"street_address1"`
    StreetAddress2 string `json:"street_address2"`
    City           string `gorm:"not null" json:"city"`
    State          string `json:"state"`
    PostalCode     string `gorm:"not null" json:"postal_code"`
    Country        string `gorm:"not null" json:"country"`
    IsDefault      bool   `gorm:"default:false" json:"is_default"`

    // Relations
    UserID *uint `json:"user_id"`
    User   *User `json:"user" gorm:"foreignKey:UserID"`
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | uint | Auto | Primary key (auto-generated) |
| `created_at` | time.Time | Auto | Record creation timestamp |
| `updated_at` | time.Time | Auto | Record last update timestamp |
| `deleted_at` | *time.Time | Auto | Soft delete timestamp |
| `street_address1` | string | Yes | Primary street address line |
| `street_address2` | string | No | Secondary address line (apartment, suite, etc.) |
| `city` | string | Yes | City name |
| `state` | string | No | State/Province name |
| `postal_code` | string | Yes | Postal/ZIP code |
| `country` | string | Yes | Country name |
| `is_default` | bool | No | Whether this is the user's default address |
| `user_id` | *uint | No | User ID (nullable for system addresses) |

### Relationships

- **User**: Belongs to a User (one-to-many relationship)
- **Orders**: Can be referenced by orders as shipping address
- **Warehouses**: Can be assigned to warehouses for location

## API Endpoints

### Base URL
All address endpoints are prefixed with `/api/v1/users/addresses`

### Authentication
All endpoints require user authentication via the `AuthMiddleware()`. Users can only access their own addresses.

---

## Address Management

### POST /api/v1/users/addresses
Create a new address for the authenticated user.

**Request Body:**
```json
{
    "street_address1": "123 Main Street",
    "street_address2": "Apartment 4B",
    "city": "Algiers",
    "state": "Algiers Province",
    "postal_code": "16000",
    "country": "Algeria",
    "is_default": true
}
```

**Response:**
```json
{
    "status": "success",
    "message": "Address created successfully",
    "data": {
        "id": 1,
        "street_address1": "123 Main Street",
        "street_address2": "Apartment 4B",
        "city": "Algiers",
        "state": "Algiers Province",
        "postal_code": "16000",
        "country": "Algeria",
        "is_default": true,
        "user_id": 15,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
    }
}
```

**Behavior:**
- If `is_default` is set to `true`, automatically unsets other default addresses for the user
- Validates required fields: `street_address1`, `city`, `postal_code`, `country`
- Associates the address with the authenticated user

### GET /api/v1/users/addresses
Get all addresses for the authenticated user.

**Response:**
```json
{
    "status": "success",
    "message": "Addresses retrieved successfully",
    "data": [
        {
            "id": 1,
            "street_address1": "123 Main Street",
            "street_address2": "Apartment 4B",
            "city": "Algiers",
            "state": "Algiers Province",
            "postal_code": "16000",
            "country": "Algeria",
            "is_default": true,
            "user_id": 15,
            "created_at": "2024-01-15T10:30:00Z",
            "updated_at": "2024-01-15T10:30:00Z"
        },
        {
            "id": 2,
            "street_address1": "456 Business Ave",
            "street_address2": "",
            "city": "Oran",
            "state": "Oran Province",
            "postal_code": "31000",
            "country": "Algeria",
            "is_default": false,
            "user_id": 15,
            "created_at": "2024-01-10T14:20:00Z",
            "updated_at": "2024-01-10T14:20:00Z"
        }
    ]
}
```

**Behavior:**
- Returns addresses ordered by default status (default first), then by creation date (newest first)
- Only returns addresses belonging to the authenticated user

### GET /api/v1/users/addresses/:id
Get a specific address by ID for the authenticated user.

**Path Parameters:**
- `id` (required): Address ID

**Response:**
```json
{
    "status": "success",
    "message": "Address retrieved successfully",
    "data": {
        "id": 1,
        "street_address1": "123 Main Street",
        "street_address2": "Apartment 4B",
        "city": "Algiers",
        "state": "Algiers Province",
        "postal_code": "16000",
        "country": "Algeria",
        "is_default": true,
        "user_id": 15,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
    }
}
```

**Error Responses:**
- `404 Not Found`: Address doesn't exist or doesn't belong to the user

### PUT /api/v1/users/addresses/:id
Update an existing address for the authenticated user.

**Path Parameters:**
- `id` (required): Address ID

**Request Body:**
```json
{
    "street_address1": "789 Updated Street",
    "city": "Constantine",
    "postal_code": "25000",
    "is_default": false
}
```

**Response:**
```json
{
    "status": "success",
    "message": "Address updated successfully",
    "data": {
        "id": 1,
        "street_address1": "789 Updated Street",
        "street_address2": "Apartment 4B",
        "city": "Constantine",
        "state": "Algiers Province",
        "postal_code": "25000",
        "country": "Algeria",
        "is_default": false,
        "user_id": 15,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T15:45:00Z"
    }
}
```

**Behavior:**
- Only updates provided fields (partial updates supported)
- If `is_default` is set to `true`, automatically unsets other default addresses
- Validates that address belongs to the authenticated user

### DELETE /api/v1/users/addresses/:id
Delete an address for the authenticated user.

**Path Parameters:**
- `id` (required): Address ID

**Response:**
```json
{
    "status": "success",
    "message": "Address deleted successfully",
    "data": null
}
```

**Behavior:**
- Soft deletes the address (sets `deleted_at` timestamp)
- Prevents deletion if address is being used in existing orders
- If the deleted address was default, automatically sets the oldest remaining address as default
- Validates that address belongs to the authenticated user

### PUT /api/v1/users/addresses/:id/default
Set a specific address as the default address for the authenticated user.

**Path Parameters:**
- `id` (required): Address ID

**Response:**
```json
{
    "status": "success",
    "message": "Default address set successfully",
    "data": {
        "id": 2,
        "street_address1": "456 Business Ave",
        "street_address2": "",
        "city": "Oran",
        "state": "Oran Province",
        "postal_code": "31000",
        "country": "Algeria",
        "is_default": true,
        "user_id": 15,
        "created_at": "2024-01-10T14:20:00Z",
        "updated_at": "2024-01-15T16:00:00Z"
    }
}
```

**Behavior:**
- Automatically unsets all other default addresses for the user
- Returns success immediately if address is already default
- Validates that address belongs to the authenticated user

---

## Request/Response Data Structures

### CreateAddressRequest
```go
type CreateAddressRequest struct {
    StreetAddress1 string `json:"street_address1" binding:"required"`
    StreetAddress2 string `json:"street_address2"`
    City           string `json:"city" binding:"required"`
    State          string `json:"state"`
    PostalCode     string `json:"postal_code" binding:"required"`
    Country        string `json:"country" binding:"required"`
    IsDefault      bool   `json:"is_default"`
}
```

### UpdateAddressRequest
```go
type UpdateAddressRequest struct {
    StreetAddress1 *string `json:"street_address1"`
    StreetAddress2 *string `json:"street_address2"`
    City           *string `json:"city"`
    State          *string `json:"state"`
    PostalCode     *string `json:"postal_code"`
    Country        *string `json:"country"`
    IsDefault      *bool   `json:"is_default"`
}
```

## Business Rules

### Default Address Management
- Each user can have only one default address at a time
- When setting an address as default, all other addresses for that user are automatically set to non-default
- When creating a new address with `is_default: true`, existing default addresses are automatically unset
- When deleting a default address, the system automatically promotes the oldest remaining address to default

### Address Validation
- `street_address1`, `city`, `postal_code`, and `country` are required fields
- `street_address2` and `state` are optional
- All string fields are trimmed of whitespace
- Country names should follow a consistent format (consider using ISO country codes)

### User Access Control
- Users can only access, modify, or delete their own addresses
- Address operations are automatically scoped to the authenticated user
- System-level addresses (warehouses, companies) have `user_id` set to `null`

## Usage Examples

### Creating a New Address
```bash
curl -X POST http://localhost:8080/api/v1/users/addresses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "street_address1": "123 Rue de la Paix",
    "street_address2": "Appartement 5",
    "city": "Algiers",
    "state": "Algiers Province",
    "postal_code": "16000",
    "country": "Algeria",
    "is_default": true
  }'
```

### Getting All User Addresses
```bash
curl -X GET http://localhost:8080/api/v1/users/addresses \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Updating an Address
```bash
curl -X PUT http://localhost:8080/api/v1/users/addresses/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "street_address1": "456 Updated Street",
    "postal_code": "16001"
  }'
```

### Setting Default Address
```bash
curl -X PUT http://localhost:8080/api/v1/users/addresses/2/default \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Deleting an Address
```bash
curl -X DELETE http://localhost:8080/api/v1/users/addresses/3 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Error Handling

The API returns standard HTTP status codes:

- `200 OK`: Successful GET/PUT request
- `201 Created`: Successful POST request (address created)
- `400 Bad Request`: Invalid request data or validation errors
- `401 Unauthorized`: Authentication required
- `404 Not Found`: Address not found or doesn't belong to user
- `500 Internal Server Error`: Server error

### Common Error Responses

#### Validation Error
```json
{
    "status": "error",
    "message": "Key: 'CreateAddressRequest.StreetAddress1' Error:Tag: 'required'",
    "data": null
}
```

#### Address Not Found
```json
{
    "status": "error",
    "message": "Address not found",
    "data": null
}
```

#### Address In Use (Cannot Delete)
```json
{
    "status": "error",
    "message": "Cannot delete address that is used in orders",
    "data": null
}
```

## Integration Points

### Order System
- Addresses are referenced by orders for shipping information
- Orders store `shipping_address_id` linking to the Address model
- Prevents deletion of addresses that are referenced by existing orders

### Warehouse System
- Warehouses reference addresses for their physical location
- Warehouse model has `address_id` field linking to Address model
- Used for inventory management and logistics

### Company System
- Companies can have associated addresses for business locations
- Used for B2B customer management and billing

## Security Considerations

- **User Isolation**: All address operations are automatically scoped to the authenticated user
- **Access Control**: Users cannot access or modify addresses belonging to other users
- **Soft Deletes**: Addresses are soft-deleted to maintain referential integrity with orders
- **Input Validation**: All address fields are validated for required data and proper formatting

## Performance Considerations

- Address queries are indexed on `user_id` for efficient user-scoped lookups
- Default address queries use compound index on `(user_id, is_default)`
- Soft delete queries exclude deleted records using `deleted_at IS NULL` condition
- Address ordering optimizes for default address first, then creation date

## Future Enhancements

- **Address Validation API**: Integration with postal address validation services
- **Geocoding**: Automatic latitude/longitude coordinates for addresses
- **Address Autocomplete**: Integration with mapping services for address suggestions
- **International Formats**: Support for different international address formats
- **Address Labels**: Custom labels for addresses (Home, Work, etc.)
- **Bulk Operations**: Import/export multiple addresses
- **Address History**: Track address change history for audit purposes 