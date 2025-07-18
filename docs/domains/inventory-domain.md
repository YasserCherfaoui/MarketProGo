# Inventory Domain

This document covers the Inventory domain, including warehouse, stock, batch, movement, and alert management endpoints, request/response formats, related models, and middleware.

---

## Overview

The Inventory domain manages stock levels, warehouses, batch tracking, stock movements, and alerts. It is essential for accurate product availability and audit trails.

---

## Endpoints

### Warehouse Management

| Method | Path                    | Description                | Auth Required |
|--------|-------------------------|----------------------------|--------------|
| POST   | /inventory/warehouses   | Create warehouse           | Yes          |
| GET    | /inventory/warehouses   | List all warehouses        | Yes          |
| GET    | /inventory/warehouses/:id | Get warehouse by ID      | Yes          |
| PUT    | /inventory/warehouses/:id | Update warehouse         | Yes          |
| DELETE | /inventory/warehouses/:id | Delete warehouse         | Yes          |

### Product Inventory Overview

| Method | Path                    | Description                | Auth Required |
|--------|-------------------------|----------------------------|--------------|
| GET    | /inventory/products     | Product inventory overview | Yes          |

### Stock Management

| Method | Path                                | Description                | Auth Required |
|--------|-------------------------------------|----------------------------|--------------|
| GET    | /inventory/stock                    | List all stock levels      | Yes          |
| POST   | /inventory/stock/adjust             | Adjust stock               | Yes          |
| GET    | /inventory/stock/by-product/:product_variant_id | Multi-warehouse stock | Yes          |

### Batch Tracking

| Method | Path                    | Description                | Auth Required |
|--------|-------------------------|----------------------------|--------------|
| GET    | /inventory/batches      | List inventory batches     | Yes          |

### Stock Movements

| Method | Path                    | Description                | Auth Required |
|--------|-------------------------|----------------------------|--------------|
| GET    | /inventory/movements    | List stock movements       | Yes          |
| GET    | /inventory/movements/:id| Get stock movement by ID   | Yes          |

### Alerts

| Method | Path                    | Description                | Auth Required |
|--------|-------------------------|----------------------------|--------------|
| GET    | /inventory/alerts       | List stock alerts          | Yes          |

---

## Request/Response Formats

### Example: Adjust Stock

```json
{
  "product_variant_id": 1,
  "warehouse_id": 2,
  "quantity": 10,
  "reason": "Restock"
}
```

### Example: Stock Level Response

```json
{
  "product_variant_id": 1,
  "warehouse_id": 2,
  "quantity": 50,
  "reserved": 5,
  "status": "active"
}
```

---

## Referenced Models

- **InventoryItem**: See `docs/models.md` for full struct.
- **Warehouse**: See `docs/models.md` for full struct.
- **StockMovement**: See `docs/models.md` for full struct.
- **ProductVariant**, **Product**.

---

## Middleware

- `AuthMiddleware`: Required for all inventory endpoints.
- `AdminMiddleware`: (Planned) For admin-only inventory actions.

---

For more details, see the Go source files in `handlers/inventory/` and `models/product.go`. 