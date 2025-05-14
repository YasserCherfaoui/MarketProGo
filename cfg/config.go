package cfg

import (
	"log"
	"os"

	"github.com/joho/godotenv" // Optional: for loading .env files
)

// AppConfig holds all application configurations
type AppConfig struct {
	Port               string
	GCSCredentialsFile string
	GCSProjectID       string
	GCSBucketName      string
	DatabaseDSN        string
	GinMode            string
	DBHost             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBPort             string
}

// LoadConfig loads configuration from environment variables
// You can extend this to load from a .env file or other sources
func LoadConfig() (*AppConfig, error) {
	// Optional: Load .env file for local development
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading, relying on environment variables.")
	}

	cfg := &AppConfig{
		Port:               getEnv("PORT", "8080"),
		GCSCredentialsFile: getEnv("GCS_CREDENTIALS_FILE", ""), // Empty means use ADC
		GCSProjectID:       getEnv("GCS_PROJECT_ID", ""),       // Often optional with ADC
		GCSBucketName:      getEnv("GCS_BUCKET_NAME", ""),
		DatabaseDSN:        getEnv("DATABASE_DSN", "files.db"), // Default to SQLite
		GinMode:            getEnv("GIN_MODE", "debug"),        // "release" for production
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBUser:             getEnv("DB_USER", "admin"),
		DBPassword:         getEnv("DB_PASSWORD", "securepass"),
		DBName:             getEnv("DB_NAME", "main"),
		DBPort:             getEnv("DB_PORT", "5434"),
	}

	if cfg.GCSBucketName == "" {
		log.Fatal("FATAL: GCS_BUCKET_NAME environment variable not set.")
	}

	log.Println("Configuration loaded successfully.")
	return cfg, nil
}

// Helper function to get an environment variable or return a default value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
