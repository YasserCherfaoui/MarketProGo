# User Domain

This document covers the User domain, including user, seller, and address management endpoints, request/response formats, related models, and middleware.

---

## Overview

The User domain manages user accounts, seller registration, and user addresses. It supports B2B (company) users and address CRUD operations.

---

## Endpoints

| Method | Path                | Description                | Auth Required |
|--------|---------------------|----------------------------|--------------|
| POST   | /users/seller       | Register a new seller      | No           |
| GET    | /users              | List all users             | Yes          |
| GET    | /users/seller       | List all sellers           | Yes          |
| DELETE | /users/:id          | Delete a user              | Yes          |

### Address Management

| Method | Path                        | Description                | Auth Required |
|--------|-----------------------------|----------------------------|--------------|
| POST   | /users/addresses            | Create address             | Yes          |
| GET    | /users/addresses            | List all addresses         | Yes          |
| GET    | /users/addresses/:id        | Get address by ID          | Yes          |
| PUT    | /users/addresses/:id        | Update address             | Yes          |
| DELETE | /users/addresses/:id        | Delete address             | Yes          |
| PUT    | /users/addresses/:id/default| Set default address        | Yes          |

---

## Request/Response Formats

### Example: Register Seller

```json
{
  "email": "seller@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+213123456789"
}
```

### Example: User Response

```json
{
  "id": 1,
  "email": "seller@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+213123456789",
  "user_type": "SELLER",
  "is_active": true,
  "addresses": [ ... ]
}
```

### Example: Address Response

```json
{
  "id": 1,
  "street_address1": "123 Main St",
  "city": "Algiers",
  "state": "Algiers",
  "postal_code": "16000",
  "country": "Algeria",
  "is_default": true
}
```

---

## Referenced Models

- **User**: See `docs/models.md` for full struct.
- **Address**: See `docs/models.md` for full struct.
- **Company**: For B2B users.

---

## Middleware

- `AuthMiddleware`: Required for all user and address endpoints except seller registration.

---

For more details, see the Go source files in `handlers/user/` and `models/user.go`. 

## Cart & Order Notes

- When users add items to their cart or place orders, the system enforces minimum quantity and dynamic pricing for each product variant.
- Requests below the minimum quantity are rejected, and the correct price is selected based on quantity breaks.

--- 