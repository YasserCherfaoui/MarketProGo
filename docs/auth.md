# Authentication & Authorization

This document describes the authentication and authorization mechanisms, including JWT usage, login flow, and middleware.

---

## Overview

Authentication is handled using JSON Web Tokens (JWT). Users log in to receive a JWT, which must be included in the `Authorization` header for protected endpoints. Middleware validates the token and attaches user info to the request context.

---

## Login Flow

1. User submits credentials to the login endpoint.
2. If valid, the server generates a JWT containing user ID, type, and (optionally) company ID.
3. The JWT is returned to the client and must be sent in the `Authorization: Bearer <token>` header for protected requests.

---

## JWT Token Structure

Tokens are generated and validated using the `utils/auth/jwt.go` utility. Example claim structure:

```go
type MyClaims struct {
    UserID    uint            `json:"user_id"`
    UserType  models.UserType `json:"user_type"`
    CompanyID *uint           `json:"company_id"`
    jwt.StandardClaims
}
```

- **GenerateToken**: Creates a JWT for a user.
- **ValidateToken**: Parses and validates a JWT, returning claims.

---

## Auth Middleware

The `AuthMiddleware` (see `middlewares/auth.go`) protects routes by:
- Checking for the `Authorization` header.
- Validating the JWT.
- Attaching user info (`user`, `user_id`, `user_type`) to the Gin context.

Example usage in a route:

```go
router.Use(middlewares.AuthMiddleware())
```

---

## Example Protected Request

```http
GET /api/v1/products HTTP/1.1
Authorization: Bearer <your-jwt-token>
```

---

## Error Handling

If the token is missing or invalid, the middleware returns a 401 Unauthorized response.

---

For more details, see `utils/auth/jwt.go` and `middlewares/auth.go`. 