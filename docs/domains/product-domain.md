# Product Domain

This document covers the Product domain, including product and product variant management endpoints, request/response formats, related models, and middleware.

---

## Overview

The Product domain manages all product-related data, including CRUD operations, variants, options, images, and inventory integration. Product variants represent specific SKUs (e.g., size, color).

---

## Endpoints

### Products

| Method | Path                | Description                | Auth Required |
|--------|---------------------|----------------------------|--------------|
| GET    | /products           | List all products          | No           |
| GET    | /products/:id       | Get product by ID          | No           |
| POST   | /products           | Create a new product       | Yes          |
| PUT    | /products/:id       | Update a product           | Yes          |
| DELETE | /products/:id       | Delete a product           | Yes          |

### Product Variants

| Method | Path                    | Description                        | Auth Required |
|--------|-------------------------|------------------------------------|--------------|
| GET    | /product-variants       | List/search product variants       | Yes          |

---

## Dynamic Pricing & Quantity Discounts

- Each product variant can define a `min_quantity` (minimum allowed purchase quantity).
- Each variant can have multiple `price_tiers` (quantity-based price breaks).
- When adding to cart or placing an order, the backend enforces `min_quantity` and selects the correct price tier based on the requested quantity.
- Example price tiers:

```json
"price_tiers": [
  { "min_quantity": 1, "price": 10.0 },
  { "min_quantity": 11, "price": 9.0 },
  { "min_quantity": 51, "price": 8.5 }
]
```

---

## Request/Response Formats (updated)

### Example: VariantData (request)

```json
{
  "name": "Small",
  "sku": "SKU123",
  "base_price": 10.0,
  "min_quantity": 1,
  "price_tiers": [
    { "min_quantity": 1, "price": 10.0 },
    { "min_quantity": 11, "price": 9.0 },
    { "min_quantity": 51, "price": 8.5 }
  ],
  "cost_price": 7.0,
  "weight": 0.5,
  "weight_unit": "kg",
  "is_active": true
}
```

### Example: ProductVariant (response)

```json
{
  "id": 1,
  "product_id": 1,
  "name": "Small",
  "sku": "SKU123",
  "base_price": 10.0,
  "min_quantity": 1,
  "price_tiers": [
    { "min_quantity": 1, "price": 10.0 },
    { "min_quantity": 11, "price": 9.0 },
    { "min_quantity": 51, "price": 8.5 }
  ],
  "cost_price": 7.0,
  "weight": 0.5,
  "weight_unit": "kg",
  "is_active": true,
  "images": [ ... ],
  "option_values": [ ... ],
  "inventory_items": [ ... ]
}
```

---

## Business Rules (updated)

- The API enforces `min_quantity` for all cart and order operations.
- The correct price is always selected from `price_tiers` based on the requested quantity.
- If no price tier matches, the base price is used.
- Requests below `min_quantity` are rejected.

---

## Referenced Models

- **Product**: See `docs/models.md` for full struct.
- **ProductVariant**: See `docs/models.md` for full struct.
- **ProductOption**, **ProductOptionValue**, **ProductImage**, **Tag**, **Category**, **Brand**.

---

## Middleware

- `AuthMiddleware`: Required for all write operations and for accessing product variants.

---

For more details, see the Go source files in `handlers/product/` and `models/product.go`. 