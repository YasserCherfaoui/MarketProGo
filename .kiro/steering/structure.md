# Project Structure

## Architecture Pattern
The project follows a **domain-driven layered architecture** with clear separation of concerns:

```
├── main.go                 # Application entry point
├── cfg/                    # Configuration management
├── database/               # Database connection setup
├── models/                 # GORM data models
├── handlers/               # HTTP request handlers (by domain)
├── routes/                 # Route definitions
├── middlewares/            # HTTP middlewares (auth, CORS, etc.)
├── utils/                  # Utility functions
├── gcs/                    # Google Cloud Storage service
├── aw/                     # Appwrite service integration
└── docs/                   # API documentation
```

## Handler Organization
Handlers are organized by business domain, each with its own package:
- `handlers/auth/` - Authentication (login, register, JWT)
- `handlers/product/` - Product management
- `handlers/inventory/` - Warehouse and stock management
- `handlers/order/` - Order processing
- `handlers/user/` - User and address management
- `handlers/cart/` - Shopping cart operations
- `handlers/brand/` - Brand management
- `handlers/category/` - Category management
- `handlers/promotion/` - Marketing promotions
- `handlers/carousel/` - Homepage banners

## Handler Pattern
Each domain handler follows a consistent structure:
```go
type DomainHandler struct {
    db              *gorm.DB
    gcsService      *gcs.GCService
    appwriteService *aw.AppwriteService
}

func NewDomainHandler(deps...) *DomainHandler {
    return &DomainHandler{...}
}
```

## Model Conventions
- All models embed `gorm.Model` for standard fields (ID, CreatedAt, UpdatedAt, DeletedAt)
- Use pointer fields for optional foreign keys (`*uint`)
- JSON tags match snake_case database columns
- Relationships defined with GORM tags (`foreignKey`, `many2many`)
- Validation tags using go-playground/validator

## API Structure
- All routes prefixed with `/api/v1`
- RESTful conventions where applicable
- Consistent JSON response format via `utils/response`
- JWT authentication middleware for protected routes
- CORS enabled for frontend integration

## File Organization Rules
- One handler per file within domain packages
- Models grouped by related entities in single files
- Utility functions organized by purpose (auth, password, response)
- Configuration centralized in `cfg/` package
- Database setup isolated in `database/` package