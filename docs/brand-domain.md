# Brand Domain

This document covers the Brand domain, including brand management endpoints, request/response formats, related models, and middleware.

---

## Overview

The Brand domain manages product brands, including CRUD operations and parent-child brand hierarchies. Brands can be linked to products and promotions.

---

## Endpoints

| Method | Path         | Description           | Auth Required |
|--------|--------------|----------------------|--------------|
| GET    | /brands      | List all brands      | No           |
| GET    | /brands/:id  | Get brand by ID      | No           |
| POST   | /brands      | Create a new brand   | Yes          |
| PUT    | /brands/:id  | Update a brand       | Yes          |
| DELETE | /brands/:id  | Delete a brand       | Yes          |

---

## Request/Response Formats

### Example: Create Brand

```json
{
  "name": "Brand Name",
  "image": "https://...",
  "slug": "brand-name",
  "is_displayed": true,
  "parent_id": null
}
```

### Example: Brand Response

```json
{
  "id": 1,
  "name": "Brand Name",
  "image": "https://...",
  "slug": "brand-name",
  "is_displayed": true,
  "parent": null,
  "children": [ ... ]
}
```

---

## Referenced Models

- **Brand**: See `docs/models.md` for full struct.
- **Product**: Brands are linked to products.

---

## Middleware

- `AuthMiddleware`: Required for all write operations.

---

For more details, see the Go source files in `handlers/brand/` and `models/brand.go`. 