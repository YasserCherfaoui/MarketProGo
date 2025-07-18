package database

import (
	"fmt"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"uniqueIndex"`
	CreatedAt time.Time
}

// TableName overrides the table name for Migration
func (Migration) TableName() string {
	return "migrations"
}

// RunMigrations runs all database migrations
func RunMigrations(db *gorm.DB) error {
	// Create migrations table if it doesn't exist
	if err := db.AutoMigrate(&Migration{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Define all migrations in order
	migrations := []struct {
		name string
		fn   func(*gorm.DB) error
	}{
		{"001_create_review_tables", createReviewTables},
		{"002_create_review_indexes", createReviewIndexes},
		{"003_create_review_constraints", createReviewConstraints},
		{"004_add_review_moderation_log", addReviewModerationLog},
		{"005_optimize_review_queries", optimizeReviewQueries},
		{"006_add_user_avatar", addUserAvatar},
	}

	// Run each migration
	for _, migration := range migrations {
		if err := runMigration(db, migration.name, migration.fn); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.name, err)
		}
	}

	return nil
}

// runMigration runs a single migration if it hasn't been run before
func runMigration(db *gorm.DB, name string, fn func(*gorm.DB) error) error {
	// Check if migration has already been run
	var count int64
	db.Model(&Migration{}).Where("name = ?", name).Count(&count)
	if count > 0 {
		fmt.Printf("Migration %s already applied, skipping\n", name)
		return nil
	}

	// Run the migration
	fmt.Printf("Running migration: %s\n", name)
	if err := fn(db); err != nil {
		return err
	}

	// Record the migration
	migration := Migration{
		Name:      name,
		CreatedAt: time.Now(),
	}
	return db.Create(&migration).Error
}

// createReviewTables creates the main review tables
func createReviewTables(db *gorm.DB) error {
	// Create ProductReview table
	if err := db.AutoMigrate(&models.ProductReview{}); err != nil {
		return fmt.Errorf("failed to create product_reviews table: %w", err)
	}

	// Create ReviewImage table
	if err := db.AutoMigrate(&models.ReviewImage{}); err != nil {
		return fmt.Errorf("failed to create review_images table: %w", err)
	}

	// Create SellerResponse table
	if err := db.AutoMigrate(&models.SellerResponse{}); err != nil {
		return fmt.Errorf("failed to create seller_responses table: %w", err)
	}

	// Create ReviewHelpful table
	if err := db.AutoMigrate(&models.ReviewHelpful{}); err != nil {
		return fmt.Errorf("failed to create review_helpful_votes table: %w", err)
	}

	// Create ProductRating table
	if err := db.AutoMigrate(&models.ProductRating{}); err != nil {
		return fmt.Errorf("failed to create product_ratings table: %w", err)
	}

	return nil
}

// createReviewIndexes creates performance indexes for review tables
func createReviewIndexes(db *gorm.DB) error {
	indexes := []struct {
		name string
		sql  string
	}{
		{
			name: "idx_product_reviews_variant_status",
			sql:  "CREATE INDEX IF NOT EXISTS idx_product_reviews_variant_status ON product_reviews (product_variant_id, status)",
		},
		{
			name: "idx_product_reviews_user_status",
			sql:  "CREATE INDEX IF NOT EXISTS idx_product_reviews_user_status ON product_reviews (user_id, status)",
		},
		{
			name: "idx_product_reviews_rating",
			sql:  "CREATE INDEX IF NOT EXISTS idx_product_reviews_rating ON product_reviews (rating)",
		},
		{
			name: "idx_product_reviews_created_at",
			sql:  "CREATE INDEX IF NOT EXISTS idx_product_reviews_created_at ON product_reviews (created_at DESC)",
		},
		{
			name: "idx_product_reviews_helpful_count",
			sql:  "CREATE INDEX IF NOT EXISTS idx_product_reviews_helpful_count ON product_reviews (helpful_count DESC)",
		},
		{
			name: "idx_product_reviews_order_item",
			sql:  "CREATE INDEX IF NOT EXISTS idx_product_reviews_order_item ON product_reviews (order_item_id) WHERE order_item_id IS NOT NULL",
		},
		{
			name: "idx_product_reviews_moderated_by",
			sql:  "CREATE INDEX IF NOT EXISTS idx_product_reviews_moderated_by ON product_reviews (moderated_by) WHERE moderated_by IS NOT NULL",
		},
		{
			name: "idx_review_images_review_id",
			sql:  "CREATE INDEX IF NOT EXISTS idx_review_images_review_id ON review_images (product_review_id)",
		},
		{
			name: "idx_seller_responses_user",
			sql:  "CREATE INDEX IF NOT EXISTS idx_seller_responses_user ON seller_responses (user_id)",
		},
		{
			name: "idx_seller_responses_review_id",
			sql:  "CREATE INDEX IF NOT EXISTS idx_seller_responses_review_id ON seller_responses (product_review_id)",
		},
		{
			name: "idx_review_helpful_review_user",
			sql:  "CREATE INDEX IF NOT EXISTS idx_review_helpful_review_user ON review_helpful_votes (product_review_id, user_id)",
		},
		{
			name: "idx_review_helpful_user",
			sql:  "CREATE INDEX IF NOT EXISTS idx_review_helpful_user ON review_helpful_votes (user_id)",
		},
		{
			name: "idx_product_ratings_variant",
			sql:  "CREATE INDEX IF NOT EXISTS idx_product_ratings_variant ON product_ratings (product_variant_id)",
		},
		{
			name: "idx_product_ratings_average",
			sql:  "CREATE INDEX IF NOT EXISTS idx_product_ratings_average ON product_ratings (average_rating DESC)",
		},
	}

	for _, idx := range indexes {
		if err := db.Exec(idx.sql).Error; err != nil {
			return fmt.Errorf("failed to create index %s: %w", idx.name, err)
		}
	}

	return nil
}

// createReviewConstraints creates foreign key constraints and unique constraints
func createReviewConstraints(db *gorm.DB) error {
	// Check if we're using SQLite (for testing) or PostgreSQL (for production)
	var dbType string
	err := db.Raw("SELECT sqlite_version()").Scan(&dbType).Error

	if err == nil && dbType != "" {
		// SQLite - constraints are handled by GORM tags, just create unique indexes
		uniqueIndexes := []struct {
			name string
			sql  string
		}{
			{
				name: "unique_review_per_user_variant",
				sql:  "CREATE UNIQUE INDEX IF NOT EXISTS unique_review_per_user_variant ON product_reviews (user_id, product_variant_id)",
			},
			{
				name: "unique_helpful_vote_per_user_review",
				sql:  "CREATE UNIQUE INDEX IF NOT EXISTS unique_helpful_vote_per_user_review ON review_helpful_votes (product_review_id, user_id)",
			},
			{
				name: "unique_seller_response_per_review",
				sql:  "CREATE UNIQUE INDEX IF NOT EXISTS unique_seller_response_per_review ON seller_responses (product_review_id)",
			},
			{
				name: "unique_product_rating_per_variant",
				sql:  "CREATE UNIQUE INDEX IF NOT EXISTS unique_product_rating_per_variant ON product_ratings (product_variant_id)",
			},
		}

		for _, idx := range uniqueIndexes {
			if err := db.Exec(idx.sql).Error; err != nil {
				return fmt.Errorf("failed to create unique index %s: %w", idx.name, err)
			}
		}
	} else {
		// PostgreSQL - create only the constraints that GORM doesn't create automatically
		// GORM already creates foreign keys and unique indexes from the model tags
		constraints := []struct {
			name string
			sql  string
		}{
			{
				name: "unique_review_per_user_variant",
				sql:  "DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'unique_review_per_user_variant') THEN ALTER TABLE product_reviews ADD CONSTRAINT unique_review_per_user_variant UNIQUE (user_id, product_variant_id); END IF; END $$;",
			},
			{
				name: "unique_helpful_vote_per_user_review",
				sql:  "DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'unique_helpful_vote_per_user_review') THEN ALTER TABLE review_helpful_votes ADD CONSTRAINT unique_helpful_vote_per_user_review UNIQUE (product_review_id, user_id); END IF; END $$;",
			},
		}

		for _, constraint := range constraints {
			if err := db.Exec(constraint.sql).Error; err != nil {
				return fmt.Errorf("failed to create constraint %s: %w", constraint.name, err)
			}
		}
	}

	return nil
}

// addReviewModerationLog adds the moderation log table
func addReviewModerationLog(db *gorm.DB) error {
	// Create ReviewModerationLog table
	if err := db.AutoMigrate(&models.ReviewModerationLog{}); err != nil {
		return fmt.Errorf("failed to create review_moderation_logs table: %w", err)
	}

	// Add indexes for moderation log
	indexes := []struct {
		name string
		sql  string
	}{
		{
			name: "idx_moderation_log_review_id",
			sql:  "CREATE INDEX IF NOT EXISTS idx_moderation_log_review_id ON review_moderation_logs (review_id)",
		},
		{
			name: "idx_moderation_log_admin_id",
			sql:  "CREATE INDEX IF NOT EXISTS idx_moderation_log_admin_id ON review_moderation_logs (admin_id)",
		},
		{
			name: "idx_moderation_log_moderated_at",
			sql:  "CREATE INDEX IF NOT EXISTS idx_moderation_log_moderated_at ON review_moderation_logs (moderated_at DESC)",
		},
		{
			name: "idx_moderation_log_status_change",
			sql:  "CREATE INDEX IF NOT EXISTS idx_moderation_log_status_change ON review_moderation_logs (old_status, new_status)",
		},
	}

	for _, idx := range indexes {
		if err := db.Exec(idx.sql).Error; err != nil {
			return fmt.Errorf("failed to create moderation log index %s: %w", idx.name, err)
		}
	}

	// Check if we're using SQLite (for testing) or PostgreSQL (for production)
	var dbType string
	err := db.Raw("SELECT sqlite_version()").Scan(&dbType).Error

	if err != nil {
		// PostgreSQL - create foreign key constraints
		constraints := []struct {
			name string
			sql  string
		}{
			{
				name: "fk_moderation_log_review",
				sql:  "DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'fk_moderation_log_review') THEN ALTER TABLE review_moderation_logs ADD CONSTRAINT fk_moderation_log_review FOREIGN KEY (review_id) REFERENCES product_reviews(id) ON DELETE CASCADE; END IF; END $$;",
			},
			{
				name: "fk_moderation_log_admin",
				sql:  "DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'fk_moderation_log_admin') THEN ALTER TABLE review_moderation_logs ADD CONSTRAINT fk_moderation_log_admin FOREIGN KEY (admin_id) REFERENCES users(id) ON DELETE CASCADE; END IF; END $$;",
			},
		}

		for _, constraint := range constraints {
			if err := db.Exec(constraint.sql).Error; err != nil {
				return fmt.Errorf("failed to create moderation log constraint %s: %w", constraint.name, err)
			}
		}
	}
	// SQLite - constraints are handled by GORM tags

	return nil
}

// optimizeReviewQueries adds additional optimization indexes and views
func optimizeReviewQueries(db *gorm.DB) error {
	// Check if we're using SQLite (for testing) or PostgreSQL (for production)
	var dbType string
	err := db.Raw("SELECT sqlite_version()").Scan(&dbType).Error

	if err == nil && dbType != "" {
		// SQLite - use CREATE VIEW instead of CREATE OR REPLACE VIEW
		viewSQL := `
			CREATE VIEW IF NOT EXISTS review_statistics AS
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
		`

		if err := db.Exec(viewSQL).Error; err != nil {
			return fmt.Errorf("failed to create review statistics view: %w", err)
		}
	} else {
		// PostgreSQL - use CREATE OR REPLACE VIEW
		viewSQL := `
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
		`

		if err := db.Exec(viewSQL).Error; err != nil {
			return fmt.Errorf("failed to create review statistics view: %w", err)
		}
	}

	// Add additional composite indexes for common query patterns
	compositeIndexes := []struct {
		name string
		sql  string
	}{
		{
			name: "idx_reviews_variant_rating_status",
			sql:  "CREATE INDEX IF NOT EXISTS idx_reviews_variant_rating_status ON product_reviews (product_variant_id, rating, status)",
		},
		{
			name: "idx_reviews_user_created_status",
			sql:  "CREATE INDEX IF NOT EXISTS idx_reviews_user_created_status ON product_reviews (user_id, created_at DESC, status)",
		},
		{
			name: "idx_reviews_helpful_rating_status",
			sql:  "CREATE INDEX IF NOT EXISTS idx_reviews_helpful_rating_status ON product_reviews (helpful_count DESC, rating DESC, status)",
		},
		{
			name: "idx_reviews_moderated_status",
			sql:  "CREATE INDEX IF NOT EXISTS idx_reviews_moderated_status ON product_reviews (moderated_at DESC, status) WHERE moderated_at IS NOT NULL",
		},
	}

	for _, idx := range compositeIndexes {
		if err := db.Exec(idx.sql).Error; err != nil {
			return fmt.Errorf("failed to create composite index %s: %w", idx.name, err)
		}
	}

	return nil
}

// RollbackMigration rolls back a specific migration
func RollbackMigration(db *gorm.DB, migrationName string) error {
	// Check if migration exists
	var migration Migration
	if err := db.Where("name = ?", migrationName).First(&migration).Error; err != nil {
		return fmt.Errorf("migration %s not found: %w", migrationName, err)
	}

	// Define rollback functions
	rollbackFunctions := map[string]func(*gorm.DB) error{
		"001_create_review_tables":      rollbackReviewTables,
		"002_create_review_indexes":     rollbackReviewIndexes,
		"003_create_review_constraints": rollbackReviewConstraints,
		"004_add_review_moderation_log": rollbackReviewModerationLog,
		"005_optimize_review_queries":   rollbackReviewQueries,
		"006_add_user_avatar":           rollbackUserAvatar,
	}

	rollbackFn, exists := rollbackFunctions[migrationName]
	if !exists {
		return fmt.Errorf("no rollback function found for migration %s", migrationName)
	}

	// Execute rollback
	if err := rollbackFn(db); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", migrationName, err)
	}

	// Remove migration record
	return db.Delete(&migration).Error
}

// rollbackReviewTables drops the review tables
func rollbackReviewTables(db *gorm.DB) error {
	// Check if we're using SQLite (for testing) or PostgreSQL (for production)
	var dbType string
	err := db.Raw("SELECT sqlite_version()").Scan(&dbType).Error

	tables := []string{
		"review_moderation_logs",
		"review_helpful_votes",
		"seller_responses",
		"review_images",
		"product_ratings",
		"product_reviews",
	}

	for _, table := range tables {
		var dropSQL string
		if err == nil && dbType != "" {
			// SQLite - no CASCADE support
			dropSQL = fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
		} else {
			// PostgreSQL - use CASCADE
			dropSQL = fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
		}

		if err := db.Exec(dropSQL).Error; err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}

// rollbackReviewIndexes drops review indexes
func rollbackReviewIndexes(db *gorm.DB) error {
	indexes := []string{
		"idx_product_reviews_variant_status",
		"idx_product_reviews_user_status",
		"idx_product_reviews_rating",
		"idx_product_reviews_created_at",
		"idx_product_reviews_helpful_count",
		"idx_product_reviews_order_item",
		"idx_product_reviews_moderated_by",
		"idx_review_images_review_id",
		"idx_seller_responses_user",
		"idx_seller_responses_review_id",
		"idx_review_helpful_review_user",
		"idx_review_helpful_user",
		"idx_product_ratings_variant",
		"idx_product_ratings_average",
	}

	for _, index := range indexes {
		if err := db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", index)).Error; err != nil {
			return fmt.Errorf("failed to drop index %s: %w", index, err)
		}
	}

	return nil
}

// rollbackReviewConstraints drops review constraints
func rollbackReviewConstraints(db *gorm.DB) error {
	// Since GORM AutoMigrate creates constraints automatically, we don't need to rollback constraints
	// that were created by the migration system. GORM will handle the constraint management.
	fmt.Println("Skipping constraint rollback - constraints are managed by GORM AutoMigrate")
	return nil
}

// rollbackReviewModerationLog drops moderation log table and related objects
func rollbackReviewModerationLog(db *gorm.DB) error {
	indexes := []string{
		"idx_moderation_log_review_id",
		"idx_moderation_log_admin_id",
		"idx_moderation_log_moderated_at",
		"idx_moderation_log_status_change",
	}

	for _, index := range indexes {
		if err := db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", index)).Error; err != nil {
			return fmt.Errorf("failed to drop moderation log index %s: %w", index, err)
		}
	}

	constraints := []string{
		"fk_moderation_log_review",
		"fk_moderation_log_admin",
	}

	for _, constraint := range constraints {
		if err := db.Exec(fmt.Sprintf("ALTER TABLE review_moderation_logs DROP CONSTRAINT IF EXISTS %s", constraint)).Error; err != nil {
			return fmt.Errorf("failed to drop moderation log constraint %s: %w", constraint, err)
		}
	}

	// Check if we're using SQLite (for testing) or PostgreSQL (for production)
	var dbType string
	err := db.Raw("SELECT sqlite_version()").Scan(&dbType).Error

	var dropSQL string
	if err == nil && dbType != "" {
		// SQLite - no CASCADE support
		dropSQL = "DROP TABLE IF EXISTS review_moderation_logs"
	} else {
		// PostgreSQL - use CASCADE
		dropSQL = "DROP TABLE IF EXISTS review_moderation_logs CASCADE"
	}

	return db.Exec(dropSQL).Error
}

// rollbackReviewQueries drops optimization objects
func rollbackReviewQueries(db *gorm.DB) error {
	// Drop the view
	if err := db.Exec("DROP VIEW IF EXISTS review_statistics").Error; err != nil {
		return fmt.Errorf("failed to drop review statistics view: %w", err)
	}

	// Drop composite indexes
	indexes := []string{
		"idx_reviews_variant_rating_status",
		"idx_reviews_user_created_status",
		"idx_reviews_helpful_rating_status",
		"idx_reviews_moderated_status",
	}

	for _, index := range indexes {
		if err := db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", index)).Error; err != nil {
			return fmt.Errorf("failed to drop composite index %s: %w", index, err)
		}
	}

	return nil
}

// addUserAvatar adds the avatar field to the users table
func addUserAvatar(db *gorm.DB) error {
	// Add avatar column to users table
	if err := db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar VARCHAR(255)").Error; err != nil {
		return fmt.Errorf("failed to add avatar column to users table: %w", err)
	}

	fmt.Println("Successfully added avatar column to users table")
	return nil
}

// rollbackUserAvatar removes the avatar field from the users table
func rollbackUserAvatar(db *gorm.DB) error {
	// Remove avatar column from users table
	if err := db.Exec("ALTER TABLE users DROP COLUMN IF EXISTS avatar").Error; err != nil {
		return fmt.Errorf("failed to remove avatar column from users table: %w", err)
	}

	fmt.Println("Successfully removed avatar column from users table")
	return nil
}

// GetMigrationStatus returns the status of all migrations
func GetMigrationStatus(db *gorm.DB) ([]Migration, error) {
	var migrations []Migration
	err := db.Order("created_at ASC").Find(&migrations).Error
	return migrations, err
}
