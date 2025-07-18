# Order Domain

This document covers the Order domain, including order and invoice management endpoints, request/response formats, related models, and middleware.

---

## Overview

The Order domain manages customer and admin order workflows, including placing orders, updating status, payment, and invoice management. Orders are linked to users, products, and inventory.

---

## Endpoints

### Customer Order Endpoints

| Method | Path                | Description                | Auth Required |
|--------|---------------------|----------------------------|--------------|
| POST   | /orders/place       | Place a new order          | Yes          |
| GET    | /orders             | List user's orders         | Yes          |
| GET    | /orders/:id         | Get order by ID            | Yes          |
| PUT    | /orders/:id/cancel  | Cancel an order            | Yes          |

### Admin Order Endpoints

| Method | Path                      | Description                | Auth Required |
|--------|---------------------------|----------------------------|--------------|
| GET    | /admin/orders             | List all orders            | Yes (Admin)  |
| GET    | /admin/orders/stats       | Get order statistics       | Yes (Admin)  |
| GET    | /admin/orders/:id         | Get order by ID            | Yes (Admin)  |
| PUT    | /admin/orders/:id/status  | Update order status        | Yes (Admin)  |
| PUT    | /admin/orders/:id/payment | Update payment status      | Yes (Admin)  |

### Admin Invoice Endpoints

| Method | Path                      | Description                | Auth Required |
|--------|---------------------------|----------------------------|--------------|
| POST   | /admin/invoices           | Create invoice             | Yes (Admin)  |
| GET    | /admin/invoices           | List invoices              | Yes (Admin)  |
| GET    | /admin/invoices/:id       | Get invoice by ID          | Yes (Admin)  |
| PUT    | /admin/invoices/:id       | Update invoice             | Yes (Admin)  |

---

## Request/Response Formats

### Example: Place Order

```json
{
  "user_id": 1,
  "items": [
    {"product_variant_id": 2, "quantity": 3}
  ],
  "shipping_address_id": 5,
  "payment_method": "CASH_ON_DELIVERY",
  "customer_notes": "Please deliver after 5pm."
}
```

### Example: Order Response

```json
{
  "id": 1,
  "order_number": "ORD-2024-0001",
  "user": { ... },
  "status": "PENDING",
  "payment_status": "PENDING",
  "total_amount": 100.0,
  "final_amount": 95.0,
  "items": [ ... ],
  "shipping_address": { ... },
  "order_date": "2024-05-01T12:00:00Z"
}
```

---

## Referenced Models

- **Order**: See `docs/models.md` for full struct.
- **OrderItem**: See `docs/models.md` for full struct.
- **Invoice**: See `docs/models.md` for full struct.
- **User**, **ProductVariant**, **Address**.

---

## Middleware

- `AuthMiddleware`: Required for all order and invoice endpoints.
- `AdminMiddleware`: (Planned) For admin-only endpoints.

---

For more details, see the Go source files in `handlers/order/` and `models/order.go`. 

## Business Rules (updated)

- When placing an order, the backend re-validates each item for the latest `min_quantity` and price tiers.
- If the quantity is below the variant’s minimum, the order is rejected.
- The correct price is selected from price tiers based on the ordered quantity.
- This ensures all orders always respect the latest business rules, even if the cart was manipulated. 