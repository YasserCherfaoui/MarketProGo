# Database Migrations

This document describes the database migration system for the review module.

## Overview

The migration system provides a structured way to manage database schema changes for the review system. It includes:

- **Versioned migrations** with rollback support
- **Performance indexes** for optimal query performance
- **Foreign key constraints** for data integrity
- **Database views** for common query patterns
- **Command-line tools** for migration management

## Migration Files

### `database/migrations.go`
Contains the core migration system with the following migrations:

1. **001_create_review_tables** - Creates the main review tables
2. **002_create_review_indexes** - Adds performance indexes
3. **003_create_review_constraints** - Adds foreign key and unique constraints
4. **004_add_review_moderation_log** - Adds moderation logging table
5. **005_optimize_review_queries** - Adds optimization views and composite indexes

### `cmd/migrate/main.go`
Command-line tool for managing migrations.

## Tables Created

### Core Review Tables

#### `product_reviews`
- Main review table storing customer reviews
- Fields: product_variant_id, user_id, rating, title, content, status, etc.
- Indexes: variant_status, user_status, rating, created_at, helpful_count

#### `review_images`
- Images attached to reviews
- Fields: product_review_id, url, alt_text
- Indexes: review_id

#### `seller_responses`
- Seller responses to customer reviews
- Fields: product_review_id, user_id, content
- Indexes: user, review_id
- Constraints: One response per review

#### `review_helpful_votes`
- User votes on review helpfulness
- Fields: product_review_id, user_id, is_helpful
- Indexes: review_user (composite), user
- Constraints: One vote per user per review

#### `product_ratings`
- Aggregated rating data for product variants
- Fields: product_variant_id, average_rating, total_reviews, rating_breakdown
- Indexes: variant, average_rating
- Constraints: One rating record per variant

#### `review_moderation_logs`
- Audit trail for moderation actions
- Fields: review_id, admin_id, old_status, new_status, reason, moderated_at
- Indexes: review_id, admin_id, moderated_at, status_change

## Indexes

### Performance Indexes

#### Product Reviews
- `idx_product_reviews_variant_status` - (product_variant_id, status)
- `idx_product_reviews_user_status` - (user_id, status)
- `idx_product_reviews_rating` - (rating)
- `idx_product_reviews_created_at` - (created_at DESC)
- `idx_product_reviews_helpful_count` - (helpful_count DESC)
- `idx_product_reviews_order_item` - (order_item_id) WHERE order_item_id IS NOT NULL
- `idx_product_reviews_moderated_by` - (moderated_by) WHERE moderated_by IS NOT NULL

#### Composite Indexes
- `idx_reviews_variant_rating_status` - (product_variant_id, rating, status)
- `idx_reviews_user_created_status` - (user_id, created_at DESC, status)
- `idx_reviews_helpful_rating_status` - (helpful_count DESC, rating DESC, status)
- `idx_reviews_moderated_status` - (moderated_at DESC, status) WHERE moderated_at IS NOT NULL

#### Other Tables
- `idx_review_images_review_id` - (product_review_id)
- `idx_seller_responses_user` - (user_id)
- `idx_seller_responses_review_id` - (product_review_id)
- `idx_review_helpful_review_user` - (product_review_id, user_id)
- `idx_review_helpful_user` - (user_id)
- `idx_product_ratings_variant` - (product_variant_id)
- `idx_product_ratings_average` - (average_rating DESC)

## Constraints

### Foreign Key Constraints
- `fk_product_reviews_variant` - product_reviews.product_variant_id → product_variants.id (CASCADE)
- `fk_product_reviews_user` - product_reviews.user_id → users.id (CASCADE)
- `fk_product_reviews_order_item` - product_reviews.order_item_id → order_items.id (SET NULL)
- `fk_product_reviews_moderated_by` - product_reviews.moderated_by → users.id (SET NULL)
- `fk_review_images_review` - review_images.product_review_id → product_reviews.id (CASCADE)
- `fk_seller_responses_review` - seller_responses.product_review_id → product_reviews.id (CASCADE)
- `fk_seller_responses_user` - seller_responses.user_id → users.id (CASCADE)
- `fk_review_helpful_review` - review_helpful_votes.product_review_id → product_reviews.id (CASCADE)
- `fk_review_helpful_user` - review_helpful_votes.user_id → users.id (CASCADE)
- `fk_product_ratings_variant` - product_ratings.product_variant_id → product_variants.id (CASCADE)

### Unique Constraints
- `unique_review_per_user_variant` - (user_id, product_variant_id)
- `unique_helpful_vote_per_user_review` - (product_review_id, user_id)
- `unique_seller_response_per_review` - (product_review_id)
- `unique_product_rating_per_variant` - (product_variant_id)

## Database Views

### `review_statistics`
A view that provides aggregated review statistics for product variants:

```sql
CREATE OR REPLACE VIEW review_statistics AS
SELECT 
    pv.id as product_variant_id,
    pv.name as variant_name,
    p.id as product_id,
    p.name as product_name,
    COALESCE(pr.average_rating, 0) as average_rating,
    COALESCE(pr.total_reviews, 0) as total_reviews,
    COALESCE(pr.rating_breakdown, '{}') as rating_breakdown,
    COUNT(DISTINCT CASE WHEN r.status = 'APPROVED' THEN r.id END) as approved_reviews,
    COUNT(DISTINCT CASE WHEN r.status = 'PENDING' THEN r.id END) as pending_reviews,
    COUNT(DISTINCT CASE WHEN r.status = 'REJECTED' THEN r.id END) as rejected_reviews,
    COUNT(DISTINCT CASE WHEN r.status = 'FLAGGED' THEN r.id END) as flagged_reviews
FROM product_variants pv
JOIN products p ON pv.product_id = p.id
LEFT JOIN product_ratings pr ON pv.id = pr.product_variant_id
LEFT JOIN product_reviews r ON pv.id = r.product_variant_id
GROUP BY pv.id, pv.name, p.id, p.name, pr.average_rating, pr.total_reviews, pr.rating_breakdown
```

## Usage

### Running Migrations

Migrations are automatically run when the application starts in release mode. You can also run them manually using the command-line tool:

```bash
# Run all migrations
go run cmd/migrate/main.go -action up

# Check migration status
go run cmd/migrate/main.go -action status

# Rollback a specific migration
go run cmd/migrate/main.go -action rollback -migration 005_optimize_review_queries
```

### Programmatic Usage

```go
import "github.com/YasserCherfaoui/MarketProGo/database"

// Run all migrations
err := database.RunMigrations(db)

// Get migration status
migrations, err := database.GetMigrationStatus(db)

// Rollback a migration
err := database.RollbackMigration(db, "migration_name")
```

## Migration Lifecycle

1. **Development**: Migrations are developed and tested locally
2. **Testing**: Migrations are tested in staging environment
3. **Production**: Migrations are applied to production database
4. **Rollback**: If needed, migrations can be rolled back safely

## Best Practices

### Adding New Migrations

1. Create a new migration function in `database/migrations.go`
2. Add it to the migrations slice in `RunMigrations()`
3. Add corresponding rollback function
4. Update tests in `database/migrations_test.go`
5. Test thoroughly before deploying

### Migration Naming

Use the format: `XXX_description` where XXX is a three-digit number.

Examples:
- `001_create_review_tables`
- `002_add_review_indexes`
- `003_update_review_constraints`

### Testing Migrations

1. Test migration application
2. Test rollback functionality
3. Test data integrity after migration
4. Test performance impact
5. Test in staging environment before production

## Troubleshooting

### Common Issues

1. **Migration already applied**: The system tracks applied migrations and won't re-apply them
2. **Rollback fails**: Ensure the rollback function exists and handles all created objects
3. **Constraint violations**: Check that data is consistent before applying constraints
4. **Index creation fails**: Verify table exists and no conflicting indexes

### Debugging

1. Check migration status: `go run cmd/migrate/main.go -action status`
2. Review database logs for specific error messages
3. Test migrations in isolation using the test suite
4. Verify database permissions and connectivity

## Performance Considerations

### Index Strategy

- **Composite indexes** for common query patterns
- **Partial indexes** for nullable columns
- **Covering indexes** for frequently accessed data
- **Regular maintenance** of index statistics

### Query Optimization

- Use the `review_statistics` view for aggregated data
- Leverage indexes for filtering and sorting
- Consider pagination for large result sets
- Monitor query performance in production

## Security

### Data Protection

- Foreign key constraints prevent orphaned records
- Unique constraints prevent duplicate data
- Cascade deletes maintain referential integrity
- Audit logging tracks all moderation actions

### Access Control

- Database user should have appropriate permissions
- Application-level authorization controls access
- Sensitive data is properly encrypted
- Audit trails are maintained for compliance 