# Documentation

This directory contains comprehensive documentation for the MarketProGo backend service.

## Directory Structure

### 📁 `/review-system/` - Product Review System
Complete documentation for the product review system implementation:

- **`admin-moderation.md`** - Admin moderation system with review management tools
- **`customer-review-management.md`** - Customer review update, delete, and history management
- **`seller-response.md`** - Seller response functionality to customer reviews
- **`review-helpfulness.md`** - Review helpfulness voting system
- **`review-display-endpoints.md`** - Public review display and retrieval endpoints
- **`rating-aggregation.md`** - Automatic rating calculation and aggregation system
- **`review-submission.md`** - Review creation and submission functionality
- **`purchase-verification.md`** - Purchase verification service for review eligibility
- **`review-handler-setup.md`** - Review handler structure and routing setup
- **`review-models.md`** - Database models for the review system
- **`task15-auth-integration.md`** - Authentication and authorization integration

### 📁 `/domains/` - Business Domain Documentation
Documentation for each business domain:

- **`user-domain.md`** - User management, addresses, and seller functionality
- **`order-domain.md`** - Order management, invoicing, and payment processing
- **`product-domain.md`** - Product and variant management with dynamic pricing
- **`inventory-domain.md`** - Inventory, warehouse, stock, and alert management
- **`promotion-domain.md`** - Marketing promotions and banner management
- **`brand-domain.md`** - Brand management with parent-child hierarchies
- **`review-domain.md`** - Product review system with moderation and rating aggregation

### 📁 `/api/` - API Documentation
API-level documentation:

- **`api-overview.md`** - High-level API structure and route organization
- **`auth.md`** - Authentication and authorization mechanisms

### 📁 `/database/` - Database Documentation
Database-related documentation:

- **`models.md`** - Complete database model definitions and relationships
- **`migrations.md`** - Database migration system and schema management

### 📄 `/utils.md` - Utility Documentation
Utility functions and helper documentation.

## Quick Navigation

### Getting Started
1. **API Overview** (`/api/api-overview.md`) - Understand the overall API structure
2. **Authentication** (`/api/auth.md`) - Learn about JWT authentication
3. **Models** (`/database/models.md`) - Review database schema

### Product Review System
The review system is fully implemented with 15 completed tasks:

1. **Setup**: `review-handler-setup.md` - Handler structure and routing
2. **Models**: `review-models.md` - Database models and relationships
3. **Purchase Verification**: `purchase-verification.md` - Ensures only customers can review
4. **Review Submission**: `review-submission.md` - Create and submit reviews
5. **Rating Aggregation**: `rating-aggregation.md` - Automatic rating calculations
6. **Display Endpoints**: `review-display-endpoints.md` - Public review retrieval
7. **Helpfulness System**: `review-helpfulness.md` - Review voting functionality
8. **Seller Responses**: `seller-response.md` - Seller reply to reviews
9. **Customer Management**: `customer-review-management.md` - User review management
10. **Admin Moderation**: `admin-moderation.md` - Review moderation tools
11. **Auth Integration**: `task15-auth-integration.md` - Security and access control

### Business Domains
- **Products**: `/domains/product-domain.md` - Product and variant management
- **Orders**: `/domains/order-domain.md` - Order processing and invoicing
- **Users**: `/domains/user-domain.md` - User and company management
- **Inventory**: `/domains/inventory-domain.md` - Stock and warehouse management
- **Promotions**: `/domains/promotion-domain.md` - Marketing and promotions
- **Brands**: `/domains/brand-domain.md` - Brand hierarchy management
- **Reviews**: `/domains/review-domain.md` - Product review system with moderation

## Implementation Status

### ✅ Complete Systems
- **Product Review System** - Fully implemented with comprehensive testing
- **Authentication & Authorization** - JWT-based with role-based access control
- **Database Models** - Complete GORM models with relationships
- **API Structure** - Well-organized REST API with versioning

### 🔧 Core Features
- Purchase verification for review eligibility
- Automatic rating aggregation and statistics
- Review moderation and admin tools
- Seller response functionality
- Review helpfulness voting
- Customer review management
- Comprehensive error handling and validation

### 🧪 Testing
- 104 passing tests with 68.6% coverage
- Unit tests for all handlers and models
- Integration tests for authentication
- Comprehensive edge case testing

## Development Guidelines

### Adding New Documentation
1. Place domain-specific docs in `/domains/`
2. Place API-related docs in `/api/`
3. Place database docs in `/database/`
4. Update this README with new entries

### Documentation Standards
- Use clear, descriptive titles
- Include code examples where relevant
- Document all API endpoints with request/response formats
- Include error handling and validation rules
- Provide testing strategies and examples

## Related Files

- **Implementation Status**: `.kiro/specs/product-review/IMPLEMENTATION_COMPLETE.md`
- **Task Tracking**: `.kiro/specs/product-review/tasks.md`
- **Main Application**: `main.go`
- **Routes**: `routes/` directory
- **Handlers**: `handlers/` directory
- **Models**: `models/` directory

