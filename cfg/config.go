package cfg

import (
	"log"
	"os"

	"github.com/joho/godotenv" // Optional: for loading .env files
)

// RevolutConfig holds Revolut API configuration
type RevolutConfig struct {
	APIKey        string
	MerchantID    string
	WebhookSecret string
	BaseURL       string // Different for sandbox and production
	IsSandbox     bool
}

// AppConfig holds all application configurations
type AppConfig struct {
	Port string
	// Google Cloud Storage
	GCSCredentialsFile string
	GCSProjectID       string
	GCSBucketName      string
	DatabaseDSN        string
	// Gin
	GinMode string
	// Database
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	// Appwrite
	AppwriteEndpoint string
	AppwriteProject  string
	AppwriteKey      string
	AppwriteBucketId string
	// Revolut configuration
	Revolut RevolutConfig
}

// LoadConfig loads configuration from environment variables
// You can extend this to load from a .env file or other sources
func LoadConfig() (*AppConfig, error) {
	// Optional: Load .env file for local development
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading, relying on environment variables.")
	}

	// Determine if we're in sandbox mode
	isSandbox := getEnv("REVOLUT_SANDBOX", "true") == "true"

	// Set base URL based on sandbox mode
	baseURL := "https://sandbox-merchant.revolut.com"
	if !isSandbox {
		baseURL = "https://merchant.revolut.com"
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
		AppwriteEndpoint:   getEnv("APPWRITE_ENDPOINT", "https://cloud.appwrite.io/v1"),
		AppwriteProject:    getEnv("APPWRITE_PROJECT", ""),
		AppwriteKey:        getEnv("APPWRITE_KEY", ""),
		AppwriteBucketId:   getEnv("APPWRITE_BUCKET_ID", ""),
		Revolut: RevolutConfig{
			APIKey:        getEnv("REVOLUT_API_KEY", ""),
			MerchantID:    getEnv("REVOLUT_MERCHANT_ID", ""),
			WebhookSecret: getEnv("REVOLUT_WEBHOOK_SECRET", ""),
			BaseURL:       baseURL,
			IsSandbox:     isSandbox,
		},
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
