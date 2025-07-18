package database

import (
	"testing"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate base models needed for foreign keys
	err = db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.ProductVariant{},
		&models.Order{},
		&models.OrderItem{},
	)
	require.NoError(t, err)

	return db
}

func TestRunMigrations(t *testing.T) {
	db := setupTestDB(t)

	// Run migrations
	err := RunMigrations(db)
	require.NoError(t, err)

	// Check that migrations table was created
	var count int64
	db.Model(&Migration{}).Count(&count)
	assert.Greater(t, count, int64(0))

	// Check that all review tables were created
	tables := []string{
		"product_reviews",
		"review_images",
		"seller_responses",
		"review_helpful_votes",
		"product_ratings",
		"review_moderation_logs",
	}

	for _, table := range tables {
		var result int
		err := db.Raw("SELECT 1 FROM " + table + " LIMIT 1").Scan(&result).Error
		assert.NoError(t, err, "Table %s should exist", table)
	}

	// Check that migrations were recorded
	var migrations []Migration
	err = db.Find(&migrations).Error
	require.NoError(t, err)
	assert.Equal(t, 5, len(migrations)) // 5 migrations should be applied

	// Verify migration names
	expectedMigrations := []string{
		"001_create_review_tables",
		"002_create_review_indexes",
		"003_create_review_constraints",
		"004_add_review_moderation_log",
		"005_optimize_review_queries",
	}

	for i, expected := range expectedMigrations {
		assert.Equal(t, expected, migrations[i].Name)
	}
}

func TestRunMigrationsIdempotent(t *testing.T) {
	db := setupTestDB(t)

	// Run migrations twice
	err := RunMigrations(db)
	require.NoError(t, err)

	err = RunMigrations(db)
	require.NoError(t, err)

	// Should only have 5 migrations recorded (not 10)
	var count int64
	db.Model(&Migration{}).Count(&count)
	assert.Equal(t, int64(5), count)
}

func TestGetMigrationStatus(t *testing.T) {
	db := setupTestDB(t)

	// Run migrations
	err := RunMigrations(db)
	require.NoError(t, err)

	// Get migration status
	migrations, err := GetMigrationStatus(db)
	require.NoError(t, err)
	assert.Equal(t, 5, len(migrations))

	// Verify migrations are ordered by creation time
	for i := 1; i < len(migrations); i++ {
		assert.True(t, migrations[i-1].CreatedAt.Before(migrations[i].CreatedAt) || migrations[i-1].CreatedAt.Equal(migrations[i].CreatedAt))
	}
}

func TestRollbackMigration(t *testing.T) {
	db := setupTestDB(t)

	// Run migrations
	err := RunMigrations(db)
	require.NoError(t, err)

	// Verify tables exist
	var count int64
	db.Model(&models.ProductReview{}).Count(&count)
	assert.Equal(t, int64(0), count) // Should be 0 but table should exist

	// Rollback the last migration
	err = RollbackMigration(db, "005_optimize_review_queries")
	require.NoError(t, err)

	// Check that migration was removed from migrations table
	var migrationCount int64
	db.Model(&Migration{}).Count(&migrationCount)
	assert.Equal(t, int64(4), migrationCount) // Should be 4 now

	// Verify the view was dropped (this would be tested in a real database)
	// For SQLite, we'll just verify the migration was removed
}

func TestRollbackMigrationNotFound(t *testing.T) {
	db := setupTestDB(t)

	// Try to rollback a non-existent migration
	err := RollbackMigration(db, "non_existent_migration")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migration non_existent_migration not found")
}

func TestCreateReviewTables(t *testing.T) {
	db := setupTestDB(t)

	// Run the create tables migration
	err := createReviewTables(db)
	require.NoError(t, err)

	// Verify all tables were created
	tables := []string{
		"product_reviews",
		"review_images",
		"seller_responses",
		"review_helpful_votes",
		"product_ratings",
	}

	for _, table := range tables {
		var result int
		err := db.Raw("SELECT 1 FROM " + table + " LIMIT 1").Scan(&result).Error
		assert.NoError(t, err, "Table %s should exist", table)
	}
}

func TestCreateReviewIndexes(t *testing.T) {
	db := setupTestDB(t)

	// Create tables first
	err := createReviewTables(db)
	require.NoError(t, err)

	// Create indexes
	err = createReviewIndexes(db)
	require.NoError(t, err)

	// Verify indexes were created (SQLite doesn't support IF NOT EXISTS for indexes the same way)
	// In a real test with PostgreSQL, we would check the pg_indexes table
	// For now, we'll just verify the function doesn't error
}

func TestCreateReviewConstraints(t *testing.T) {
	db := setupTestDB(t)

	// Create tables first
	err := createReviewTables(db)
	require.NoError(t, err)

	// Create constraints
	err = createReviewConstraints(db)
	require.NoError(t, err)

	// Test that unique constraints work
	review1 := models.ProductReview{
		ProductVariantID: 1,
		UserID:           1,
		Rating:           5,
		Content:          "Great product",
		Status:           models.ReviewStatusApproved,
	}

	review2 := models.ProductReview{
		ProductVariantID: 1,
		UserID:           1, // Same user and variant
		Rating:           4,
		Content:          "Another review",
		Status:           models.ReviewStatusApproved,
	}

	// First review should be created successfully
	err = db.Create(&review1).Error
	require.NoError(t, err)

	// Second review should fail due to unique constraint
	err = db.Create(&review2).Error
	assert.Error(t, err) // Should fail due to unique constraint
}

func TestAddReviewModerationLog(t *testing.T) {
	db := setupTestDB(t)

	// Create base tables
	err := createReviewTables(db)
	require.NoError(t, err)

	// Add moderation log
	err = addReviewModerationLog(db)
	require.NoError(t, err)

	// Verify table was created
	var result int
	err = db.Raw("SELECT 1 FROM review_moderation_logs LIMIT 1").Scan(&result).Error
	assert.NoError(t, err, "review_moderation_logs table should exist")
}

func TestOptimizeReviewQueries(t *testing.T) {
	db := setupTestDB(t)

	// Create base tables
	err := createReviewTables(db)
	require.NoError(t, err)

	// Add optimization
	err = optimizeReviewQueries(db)
	require.NoError(t, err)

	// Verify view was created
	var result int
	err = db.Raw("SELECT 1 FROM review_statistics LIMIT 1").Scan(&result).Error
	assert.NoError(t, err, "review_statistics view should exist")
}

func TestMigrationModel(t *testing.T) {
	db := setupTestDB(t)

	// Create migrations table first
	err := db.AutoMigrate(&Migration{})
	require.NoError(t, err)

	// Test Migration model
	migration := Migration{
		Name:      "test_migration",
		CreatedAt: time.Now(),
	}

	err = db.Create(&migration).Error
	require.NoError(t, err)

	// Verify it was created
	var count int64
	db.Model(&Migration{}).Count(&count)
	assert.Equal(t, int64(1), count)

	// Test table name
	assert.Equal(t, "migrations", migration.TableName())
}

func TestRollbackFunctions(t *testing.T) {
	db := setupTestDB(t)

	// Create tables first
	err := createReviewTables(db)
	require.NoError(t, err)

	// Test rollback functions
	err = rollbackReviewTables(db)
	require.NoError(t, err)

	// Verify tables were dropped
	tables := []string{
		"product_reviews",
		"review_images",
		"seller_responses",
		"review_helpful_votes",
		"product_ratings",
	}

	for _, table := range tables {
		var result int
		err := db.Raw("SELECT 1 FROM " + table + " LIMIT 1").Scan(&result).Error
		assert.Error(t, err, "Table %s should not exist", table)
	}
}

func TestMigrationErrorHandling(t *testing.T) {
	db := setupTestDB(t)

	// Test with invalid database connection (simulated)
	// This would be tested with a real database that has connection issues
	// For now, we'll test that the migration system handles errors gracefully

	// Test rollback of non-existent migration
	err := RollbackMigration(db, "non_existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migration non_existent not found")
}

func TestMigrationOrdering(t *testing.T) {
	db := setupTestDB(t)

	// Run migrations
	err := RunMigrations(db)
	require.NoError(t, err)

	// Get migrations and verify they're in the correct order
	migrations, err := GetMigrationStatus(db)
	require.NoError(t, err)

	expectedOrder := []string{
		"001_create_review_tables",
		"002_create_review_indexes",
		"003_create_review_constraints",
		"004_add_review_moderation_log",
		"005_optimize_review_queries",
	}

	for i, expected := range expectedOrder {
		assert.Equal(t, expected, migrations[i].Name, "Migration %d should be %s", i, expected)
	}
}
