package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/YasserCherfaoui/MarketProGo/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Parse command line flags
	var (
		action        = flag.String("action", "up", "Migration action: up, status, rollback")
		migrationName = flag.String("migration", "", "Migration name for rollback")
		envFile       = flag.String("env", ".env", "Environment file path")
	)
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(*envFile); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Connect to database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Execute action
	switch *action {
	case "up":
		fmt.Println("Running migrations...")
		if err := database.RunMigrations(db); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		fmt.Println("Migrations completed successfully!")

	case "status":
		fmt.Println("Migration status:")
		migrations, err := database.GetMigrationStatus(db)
		if err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}

		if len(migrations) == 0 {
			fmt.Println("No migrations have been applied.")
		} else {
			fmt.Printf("%-30s %-20s\n", "Migration", "Applied At")
			fmt.Println(string(make([]byte, 50, 50)))
			for _, migration := range migrations {
				fmt.Printf("%-30s %-20s\n", migration.Name, migration.CreatedAt.Format("2006-01-02 15:04:05"))
			}
		}

	case "rollback":
		if *migrationName == "" {
			log.Fatal("Migration name is required for rollback. Use -migration flag.")
		}
		fmt.Printf("Rolling back migration: %s\n", *migrationName)
		if err := database.RollbackMigration(db, *migrationName); err != nil {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
		fmt.Printf("Migration %s rolled back successfully!\n", *migrationName)

	default:
		fmt.Println("Usage: migrate [options]")
		fmt.Println("Options:")
		fmt.Println("  -action string")
		fmt.Println("        Migration action: up, status, rollback (default 'up')")
		fmt.Println("  -migration string")
		fmt.Println("        Migration name for rollback")
		fmt.Println("  -env string")
		fmt.Println("        Environment file path (default '.env')")
		fmt.Println("\nExamples:")
		fmt.Println("  migrate -action up")
		fmt.Println("  migrate -action status")
		fmt.Println("  migrate -action rollback -migration 001_create_review_tables")
		os.Exit(1)
	}
}
