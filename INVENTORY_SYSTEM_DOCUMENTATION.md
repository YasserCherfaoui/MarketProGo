# Inventory Management System Documentation

## Overview

The Inventory Management System provides comprehensive warehouse and stock management capabilities for the Algeria Market platform. It includes warehouse management, stock level tracking, inventory adjustments, stock transfers, and detailed reporting with analytics dashboard.

## Features

- **Warehouse Management**: Create, update, and manage multiple warehouses
- **Stock Level Tracking**: Monitor inventory levels across all warehouses
- **Stock Adjustments**: Adjust inventory quantities with audit trail
- **Bulk Operations**: Perform bulk stock adjustments for efficiency
- **Stock Transfers**: Transfer inventory between warehouses
- **Dashboard Analytics**: Comprehensive reporting and analytics
- **Alert System**: Low stock and expiring item alerts
- **Audit Trail**: Complete history of all stock movements

## Database Schema

### Core Models

#### InventoryItem
Tracks stock for a specific product variant in a warehouse.

```go
type InventoryItem struct {
    gorm.Model
    ProductVariantID uint           `json:"product_variant_id"`
    ProductVariant   ProductVariant `json:"-"`
    WarehouseID      uint           `json:"warehouse_id"`
    Warehouse        Warehouse      `json:"warehouse"`
    Quantity         int            `json:"quantity"`
    Reserved         int            `json:"reserved"`
    BatchNumber      string         `json:"batch_number"`
    ExpiryDate       *time.Time     `json:"expiry_date"`
    Status           string         `json:"status"` // active, expired, damaged
}
```

#### Warehouse
Represents a physical warehouse location.

```go
type Warehouse struct {
    gorm.Model
    Name           string          `json:"name"`
    Code           string          `json:"code"`
    AddressID      uint            `json:"address_id"`
    Address        Address         `json:"address"`
    IsActive       bool            `json:"is_active"`
    InventoryItems []InventoryItem `json:"inventory_items"`
}
```

#### StockMovement
Audit trail for all inventory movements.

```go
type StockMovement struct {
    gorm.Model
    InventoryItemID uint          `json:"inventory_item_id"`
    InventoryItem   InventoryItem `json:"inventory_item"`
    MovementType    string        `json:"movement_type"` // adjustment_in, adjustment_out, transfer_in, transfer_out, sold, returned
    Quantity        int           `json:"quantity"`
    Reason          string        `json:"reason"`
    Notes           string        `json:"notes"`
    Reference       string        `json:"reference"` // Order ID, Transfer ID, etc.
    UserID          *uint         `json:"user_id"`
    User            *User         `json:"user,omitempty"`
}
```

## API Endpoints

### Base URL
All inventory endpoints are prefixed with `/api/v1/inventory`

### Authentication
All endpoints require authentication via the `AuthMiddleware()`. Admin privileges may be required (commented out in current implementation).

---

## Dashboard & Analytics

### GET /api/v1/inventory/dashboard
Get comprehensive inventory dashboard data for admin overview.

**Query Parameters:**
- `days_back` (optional): Number of days back for movement history (default: 30)

**Response:**
```json
{
    "status": "success",
    "message": "Inventory dashboard data retrieved successfully",
    "data": {
        "summary": {
            "total_warehouses": 5,
            "active_warehouses": 4,
            "total_products": 150,
            "total_quantity": 5000,
            "reserved_quantity": 200,
            "available_quantity": 4800,
            "total_value": 125000.50,
            "low_stock_count": 12,
            "out_of_stock_count": 3,
            "expiring_items_count": 8
        },
        "warehouse_summary": [
            {
                "warehouse_id": 1,
                "warehouse_name": "Main Warehouse",
                "warehouse_code": "WH001",
                "product_count": 75,
                "total_quantity": 2500,
                "total_value": 62500.25,
                "low_stock_count": 5,
                "is_active": true
            }
        ],
        "low_stock_items": [
            {
                "product_variant_id": 123,
                "product_name": "Premium Coffee",
                "variant_name": "1kg",
                "sku": "COFFEE-1KG-001",
                "warehouse_name": "Main Warehouse",
                "current_quantity": 5,
                "reserved_quantity": 2,
                "available_quantity": 3,
                "reorder_level": 10
            }
        ],
        "expiring_items": [
            {
                "product_variant_id": 456,
                "product_name": "Fresh Milk",
                "variant_name": "1L",
                "sku": "MILK-1L-001",
                "warehouse_name": "Cold Storage",
                "batch_number": "BATCH-2024-001",
                "quantity": 50,
                "expiry_date": "2024-02-15T00:00:00Z",
                "days_to_expiry": 7
            }
        ],
        "top_products_by_value": [
            {
                "product_variant_id": 789,
                "product_name": "Premium Electronics",
                "variant_name": "Standard",
                "sku": "ELEC-STD-001",
                "total_quantity": 100,
                "total_value": 25000.00,
                "cost_price": 250.00
            }
        ],
        "recent_movements": [
            {
                "movement_id": 1001,
                "product_name": "Coffee Beans",
                "variant_name": "500g",
                "sku": "COFFEE-500G-001",
                "warehouse_name": "Main Warehouse",
                "movement_type": "adjustment_in",
                "quantity": 100,
                "reason": "New stock arrival",
                "created_at": "2024-01-15T10:30:00Z",
                "user_name": "admin@example.com"
            }
        ],
        "stock_alerts": [
            {
                "alert_type": "low_stock",
                "product_variant_id": 123,
                "product_name": "Premium Coffee",
                "variant_name": "1kg",
                "sku": "COFFEE-1KG-001",
                "warehouse_name": "Main Warehouse",
                "current_quantity": 5,
                "threshold_quantity": 10,
                "severity": "warning"
            }
        ]
    }
}
```

---

## Warehouse Management

### POST /api/v1/inventory/warehouses
Create a new warehouse.

**Request Body:**
```json
{
    "name": "North Warehouse",
    "code": "WH-NORTH-001",
    "address_id": 123,
    "is_active": true
}
```

**Response:**
```json
{
    "status": "success",
    "message": "Warehouse created successfully",
    "data": {
        "id": 5,
        "name": "North Warehouse",
        "code": "WH-NORTH-001",
        "address_id": 123,
        "address": {
            "id": 123,
            "street_address1": "123 Industrial Ave",
            "city": "Algiers",
            "postal_code": "16000",
            "country": "Algeria"
        },
        "is_active": true,
        "created_at": "2024-01-15T10:30:00Z"
    }
}
```

### GET /api/v1/inventory/warehouses
Get all warehouses with filtering and pagination.

**Query Parameters:**
- `name` (optional): Filter by warehouse name (case-insensitive partial match)
- `code` (optional): Filter by warehouse code (case-insensitive partial match)
- `is_active` (optional): Filter by active status (true/false)
- `include_stock` (optional): Include stock summary for each warehouse (default: false)
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 20, max: 100)

**Response:**
```json
{
    "status": "success",
    "message": "Warehouses retrieved successfully",
    "data": {
        "data": [
            {
                "id": 1,
                "name": "Main Warehouse",
                "code": "WH001",
                "address": {
                    "street_address1": "123 Main St",
                    "city": "Algiers"
                },
                "is_active": true
            }
        ],
        "total": 5,
        "page": 1,
        "page_size": 20
    }
}
```

### GET /api/v1/inventory/warehouses/:id
Get a single warehouse with detailed stock information.

**Response:**
```json
{
    "status": "success",
    "message": "Warehouse retrieved successfully",
    "data": {
        "warehouse": {
            "id": 1,
            "name": "Main Warehouse",
            "code": "WH001",
            "address": {
                "street_address1": "123 Main St",
                "city": "Algiers"
            },
            "is_active": true
        },
        "stock_summary": {
            "total_products": 75,
            "total_quantity": 2500,
            "low_stock_items": 5,
            "out_of_stock_items": 2,
            "total_value": 62500.25,
            "reserved_quantity": 100
        }
    }
}
```

### PUT /api/v1/inventory/warehouses/:id
Update warehouse details.

**Request Body:**
```json
{
    "name": "Updated Warehouse Name",
    "code": "WH-UPDATED-001",
    "address_id": 456,
    "is_active": false
}
```

### DELETE /api/v1/inventory/warehouses/:id
Delete a warehouse (soft delete). Fails if warehouse has inventory items.

**Response:**
```json
{
    "status": "success",
    "message": "Warehouse deleted successfully",
    "data": null
}
```

---

## Stock Management

### GET /api/v1/inventory/stock
Get stock levels with filtering and pagination.

**Query Parameters:**
- `warehouse_id` (optional): Filter by warehouse ID
- `product_variant_id` (optional): Filter by product variant ID
- `sku` (optional): Filter by product SKU (partial match)
- `status` (optional): Filter by status (active, expired, damaged)
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 50, max: 100)

**Response:**
```json
{
    "status": "success",
    "message": "Stock levels retrieved successfully",
    "data": {
        "data": [
            {
                "id": 1001,
                "product_variant_id": 123,
                "warehouse_id": 1,
                "quantity": 150,
                "reserved": 20,
                "batch_number": "BATCH-2024-001",
                "expiry_date": "2024-12-31T00:00:00Z",
                "status": "active",
                "product_variant": {
                    "id": 123,
                    "name": "1kg",
                    "sku": "COFFEE-1KG-001",
                    "product": {
                        "name": "Premium Coffee"
                    }
                },
                "warehouse": {
                    "id": 1,
                    "name": "Main Warehouse",
                    "code": "WH001"
                },
                "available_quantity": 130,
                "stock_status": "in_stock"
            }
        ],
        "total": 150,
        "page": 1,
        "page_size": 50
    }
}
```

### POST /api/v1/inventory/stock/adjust
Adjust stock level for a specific product variant in a warehouse.

**Request Body:**
```json
{
    "product_variant_id": 123,
    "warehouse_id": 1,
    "quantity": 50,
    "batch_number": "BATCH-2024-002",
    "expiry_date": "2024-12-31",
    "reason": "New stock arrival",
    "notes": "Received from supplier ABC"
}
```

**Response:**
```json
{
    "status": "success",
    "message": "Stock adjusted successfully",
    "data": {
        "id": 1002,
        "product_variant_id": 123,
        "warehouse_id": 1,
        "quantity": 200,
        "reserved": 20,
        "batch_number": "BATCH-2024-002",
        "expiry_date": "2024-12-31T00:00:00Z",
        "status": "active",
        "product_variant": {
            "name": "1kg",
            "sku": "COFFEE-1KG-001",
            "product": {
                "name": "Premium Coffee"
            }
        },
        "warehouse": {
            "name": "Main Warehouse",
            "code": "WH001"
        }
    }
}
```

---

## Request/Response Data Structures

### StockAdjustmentRequest
```go
type StockAdjustmentRequest struct {
    ProductVariantID uint    `json:"product_variant_id" binding:"required"`
    WarehouseID      uint    `json:"warehouse_id" binding:"required"`
    Quantity         int     `json:"quantity" binding:"required"`
    BatchNumber      string  `json:"batch_number"`
    ExpiryDate       *string `json:"expiry_date"` // YYYY-MM-DD format
    Reason           string  `json:"reason" binding:"required"`
    Notes            string  `json:"notes"`
}
```

### BulkStockAdjustmentRequest
```go
type BulkStockAdjustmentRequest struct {
    WarehouseID uint                     `json:"warehouse_id" binding:"required"`
    Items       []StockAdjustmentRequest `json:"items" binding:"required,dive"`
    Reason      string                   `json:"reason" binding:"required"`
    Notes       string                   `json:"notes"`
}
```

### StockTransferRequest
```go
type StockTransferRequest struct {
    ProductVariantID  uint   `json:"product_variant_id" binding:"required"`
    FromWarehouseID   uint   `json:"from_warehouse_id" binding:"required"`
    ToWarehouseID     uint   `json:"to_warehouse_id" binding:"required"`
    Quantity          int    `json:"quantity" binding:"required,min=1"`
    TransferReference string `json:"transfer_reference"`
    Notes             string `json:"notes"`
}
```

### CreateWarehouseRequest
```go
type CreateWarehouseRequest struct {
    Name      string `json:"name" binding:"required"`
    Code      string `json:"code" binding:"required"`
    AddressID uint   `json:"address_id" binding:"required"`
    IsActive  *bool  `json:"is_active"`
}
```

### UpdateWarehouseRequest
```go
type UpdateWarehouseRequest struct {
    Name      *string `json:"name"`
    Code      *string `json:"code"`
    AddressID *uint   `json:"address_id"`
    IsActive  *bool   `json:"is_active"`
}
```

## Stock Status Types

- `in_stock`: Quantity > 10
- `low_stock`: Quantity > 0 and <= 10
- `out_of_stock`: Quantity = 0

## Movement Types

- `adjustment_in`: Stock increase via adjustment
- `adjustment_out`: Stock decrease via adjustment
- `transfer_in`: Stock received from transfer
- `transfer_out`: Stock sent via transfer
- `sold`: Stock sold to customer
- `returned`: Stock returned from customer

## Alert Types

- `low_stock`: Items with quantity <= 10
- `out_of_stock`: Items with quantity = 0
- `expiring_soon`: Items expiring within 30 days
- `expired`: Items past expiry date

## Alert Severity Levels

- `critical`: Urgent attention required (out of stock, expired, very low stock)
- `warning`: Attention needed (low stock, expiring soon)
- `info`: Informational only

## Usage Examples

### Creating a Warehouse
```bash
curl -X POST http://localhost:8080/api/v1/inventory/warehouses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "South Warehouse",
    "code": "WH-SOUTH-001",
    "address_id": 456,
    "is_active": true
  }'
```

### Adjusting Stock
```bash
curl -X POST http://localhost:8080/api/v1/inventory/stock/adjust \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "product_variant_id": 123,
    "warehouse_id": 1,
    "quantity": 100,
    "batch_number": "BATCH-2024-003",
    "expiry_date": "2024-12-31",
    "reason": "Stock replenishment",
    "notes": "Monthly restock from supplier"
  }'
```

### Getting Dashboard Data
```bash
curl -X GET "http://localhost:8080/api/v1/inventory/dashboard?days_back=7" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Error Handling

The API returns standard HTTP status codes:

- `200 OK`: Successful GET request
- `201 Created`: Successful POST request (resource created)
- `400 Bad Request`: Invalid request data or business logic violation
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

Error responses follow this format:
```json
{
    "status": "error",
    "message": "Detailed error message",
    "data": null
}
```

## Authentication

All endpoints require authentication via Bearer token:
```
Authorization: Bearer YOUR_JWT_TOKEN
```

## Commented Features

Several advanced features are implemented but commented out in the routes:

- Bulk stock adjustments
- Stock transfers between warehouses
- Stock reservations
- Detailed stock movement history
- Advanced reporting endpoints
- Stock alert management

These can be enabled by uncommenting the relevant routes in `routes/inventory_routes.go`.

## Performance Considerations

- Use pagination for large datasets
- Database queries are optimized with proper joins and indexing
- Stock calculations use efficient aggregation queries
- Dashboard data includes caching-friendly structures

## Future Enhancements

- Real-time inventory updates via WebSocket
- Barcode scanning integration
- Advanced forecasting and demand planning
- Integration with supplier systems
- Mobile app support for warehouse operations
- Multi-currency support for international operations 