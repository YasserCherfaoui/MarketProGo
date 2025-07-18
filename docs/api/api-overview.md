# API Overview

This document provides a high-level overview of the API structure, route composition, and main modules of the backend service. The API is built with [Gin](https://gin-gonic.com/) for HTTP routing and [GORM](https://gorm.io/) for ORM/database access.

## Main Route Structure

All API endpoints are grouped under the `/api/v1` prefix. The main route groups and their purposes are as follows:

| Route Group         | Purpose/Description                        |
|---------------------|--------------------------------------------|
| `/auth`             | Authentication (login, register, JWT)      |
| `/categories`       | Product categories management              |
| `/brands`           | Brand management (with parent/child)       |
| `/products`         | Product CRUD, details, and search          |
| `/product-variants` | Product variant management (stock, etc.)   |
| `/users`            | User management, addresses, sellers        |
| `/carousel`         | Carousel banners for homepage              |
| `/cart`             | Shopping cart operations                   |
| `/orders`           | Customer order management                  |
| `/admin/orders`     | Admin order management                     |
| `/admin/invoices`   | Admin invoice management                   |
| `/inventory`        | Inventory, warehouse, stock, alerts        |
| `/promotions`       | Promotions and marketing banners           |
| `/file/preview`     | File/image proxying                        |

## Example Route Registration (from `app_routes.go`)

```go
func AppRoutes(r *gin.Engine, db *gorm.DB, gcsService *gcs.GCService, appwriteService *aw.AppwriteService) {
    router := r.Group("/api/v1")
    AuthRoutes(router, authHandler)
    CategoryRoutes(router, db, gcsService, appwriteService)
    BrandRoutes(router, db, gcsService, appwriteService)
    ProductRoutes(router, db, gcsService, appwriteService)
    UserRoutes(router, db)
    CarouselRoutes(router, db, gcsService, appwriteService)
    CartRoutes(router, db)
    OrderRoutes(router, db)
    InventoryRoutes(router, inventoryHandler)
    RegisterPromotionRoutes(router, promotionHandler)
    router.GET("/file/preview/:fileId", fileHandler.ProxyFilePreview)
}
```

## Middleware and Authentication

- **AuthMiddleware**: Most protected routes use JWT-based authentication. The middleware checks the `Authorization` header for a Bearer token, validates it, and attaches user info to the request context.
- **Admin Middleware**: (Planned) For admin-only routes, an additional middleware will be used to restrict access.

## Versioning

All endpoints are versioned under `/api/v1` to allow for future expansion and backward compatibility.

## Error Handling

All responses use a consistent JSON structure. See `docs/utils.md` for details on the response format and error handling.

---

For detailed documentation on each domain/module, see the corresponding files in this `docs/` directory. 