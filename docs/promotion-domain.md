# Promotion Domain

This document covers the Promotion domain, including promotion management endpoints, request/response formats, related models, and middleware.

---

## Overview

The Promotion domain manages marketing banners and promotions, which can be linked to products, categories, or brands. Promotions have scheduling, images, and optional call-to-action buttons.

---

## Endpoints

| Method | Path             | Description           | Auth Required |
|--------|------------------|----------------------|--------------|
| POST   | /promotions      | Create a promotion   | Yes          |
| GET    | /promotions      | List all promotions  | No           |
| GET    | /promotions/:id  | Get promotion by ID  | No           |
| PUT    | /promotions/:id  | Update a promotion   | Yes          |
| DELETE | /promotions/:id  | Delete a promotion   | Yes          |

---

## Request/Response Formats

### Example: Create Promotion

```json
{
  "title": "Summer Sale",
  "description": "Up to 50% off!",
  "image": "https://...",
  "button_text": "Shop Now",
  "button_link": "/products",
  "start_date": "2024-06-01T00:00:00Z",
  "end_date": "2024-06-30T23:59:59Z",
  "is_active": true,
  "product_id": 1,
  "category_id": null,
  "brand_id": null
}
```

### Example: Promotion Response

```json
{
  "id": 1,
  "title": "Summer Sale",
  "description": "Up to 50% off!",
  "image": "https://...",
  "button_text": "Shop Now",
  "button_link": "/products",
  "start_date": "2024-06-01T00:00:00Z",
  "end_date": "2024-06-30T23:59:59Z",
  "is_active": true,
  "product": { ... },
  "category": null,
  "brand": null
}
```

---

## Referenced Models

- **Promotion**: See `docs/models.md` for full struct.
- **Product**, **Category**, **Brand**: Promotions can be linked to these entities.

---

## Middleware

- `AuthMiddleware`: Required for all write operations.

---

For more details, see the Go source files in `handlers/promotion/` and `models/promotion.go`. 