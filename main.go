package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/cfg"
	"github.com/YasserCherfaoui/MarketProGo/database"
	"github.com/YasserCherfaoui/MarketProGo/email"
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	emailHandler "github.com/YasserCherfaoui/MarketProGo/handlers/email"
	"github.com/YasserCherfaoui/MarketProGo/redis"
	"github.com/YasserCherfaoui/MarketProGo/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := cfg.LoadConfig()
	if err != nil {
		log.Fatalf("FATAL: Could not load configuration: %v", err)
	}

	// Validate Revolut configuration
	if cfg.Revolut.APIKey == "" {
		log.Fatal("ERROR: REVOLUT_API_KEY environment variable is not set. Please configure your Revolut API credentials.")
	}
	if cfg.Revolut.BaseURL == "" {
		log.Fatal("ERROR: Revolut base URL is not configured.")
	}

	log.Printf("Revolut configuration loaded - BaseURL: %s, IsSandbox: %t", cfg.Revolut.BaseURL, cfg.Revolut.IsSandbox)

	r := gin.Default()
	config := cors.Config{
		AllowOrigins:     []string{"*", "http://localhost:5173", "http://127.0.0.1:5173"}, // Adjust origins
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // Cache preflight request for 12 hours
	}
	// GCS
	ctx, cancelInit := context.WithTimeout(context.Background(), 30*time.Second) // 30s timeout for init
	defer cancelInit()
	gcsService, err := gcs.NewGCSService(ctx, cfg.GCSCredentialsFile, cfg.GCSProjectID, cfg.GCSBucketName)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize GCS service: %v", err)
	}
	defer func() {
		if err := gcsService.Close(); err != nil {
			log.Printf("ERROR: Failed to close GCS service: %v", err)
		}
	}()

	// Appwrite
	appwriteClient := aw.NewAppwriteClient(cfg)
	appwriteService := aw.NewAppwriteService(appwriteClient)

	r.Use(cors.New(config))
	db, err := database.ConnectDB()
	if err != nil {
		panic(err)
	}

	// Initialize Redis service for email queue
	redisService, err := redis.NewRedisService(&redis.RedisConfig{
		UpstashURL:   cfg.Redis.UpstashURL,
		UpstashToken: cfg.Redis.UpstashToken,
		PoolSize:     cfg.Redis.PoolSize,
	})
	if err != nil {
		log.Printf("WARNING: Failed to initialize Redis service: %v", err)
		log.Printf("Email queue functionality will be disabled")
	}

	// Initialize email system components
	var emailProvider email.EmailProvider
	var emailQueue email.EmailQueue
	var templateEngine email.TemplateEngine
	var emailAnalytics email.EmailAnalytics

	// Initialize email provider (Graph API or Mock)
	log.Printf("🔧 EMAIL: Initializing email provider...")
	log.Printf("🔍 EMAIL: Outlook config - TenantID: %s, ClientID: %s, SenderEmail: %s",
		cfg.Outlook.TenantID, cfg.Outlook.ClientID, cfg.Outlook.SenderEmail)

	// For testing purposes, use mock provider
	log.Printf("⚠️ EMAIL: Using mock provider for development/testing")
	emailProvider = email.NewMockEmailProvider(
		cfg.Email.SenderEmail,
		cfg.Email.SenderName,
	)

	// Try to initialize Graph provider first with panic recovery
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("🚨 EMAIL: Panic occurred during Graph provider initialization: %v", r)
				log.Printf("⚠️ EMAIL: Falling back to mock provider for development")
				emailProvider = email.NewMockEmailProvider(
					cfg.Email.SenderEmail,
					cfg.Email.SenderName,
				)
			}
		}()

		// Check if required Outlook config is present
		if cfg.Outlook.TenantID == "" || cfg.Outlook.ClientID == "" || cfg.Outlook.ClientSecret == "" {
			log.Printf("⚠️ EMAIL: Missing Outlook configuration - TenantID: %s, ClientID: %s, ClientSecret: %s",
				cfg.Outlook.TenantID, cfg.Outlook.ClientID,
				func() string {
					if cfg.Outlook.ClientSecret == "" {
						return "EMPTY"
					}
					return "SET"
				}())
			log.Printf("⚠️ EMAIL: Falling back to mock provider for development")
			emailProvider = email.NewMockEmailProvider(
				cfg.Email.SenderEmail,
				cfg.Email.SenderName,
			)
			return
		}

		graphProvider, err := email.NewGraphEmailProvider(&cfg.Outlook)
		if err != nil {
			log.Printf("⚠️ EMAIL: Failed to initialize Graph provider: %v", err)
			log.Printf("⚠️ EMAIL: Falling back to mock provider for development")
			emailProvider = email.NewMockEmailProvider(
				cfg.Email.SenderEmail,
				cfg.Email.SenderName,
			)
		} else {
			log.Printf("✅ EMAIL: Successfully initialized Graph provider")
			emailProvider = graphProvider
		}
	}()

	// Initialize template engine
	templateEngine = email.NewHTMLTemplateEngine("templates/emails")
	if err := templateEngine.ReloadTemplates(); err != nil {
		log.Printf("WARNING: Failed to load email templates: %v", err)
	}

	// Initialize email queue if Redis is available
	if redisService != nil {
		emailQueue = email.NewRedisEmailQueue(redisService.GetClient(), "email_queue")
	} else {
		// Use mock queue when Redis is not available
		emailQueue = email.NewMockEmailQueue()
		log.Printf("Using mock email queue (Redis not available)")
	}

	// Initialize email analytics
	emailAnalytics = email.NewEmailAnalytics(db)

	// Initialize email service
	emailService := email.NewEmailService(
		emailProvider,
		templateEngine,
		emailQueue,
		emailAnalytics,
		&cfg.Email,
		db,
	)

	// Initialize email trigger service (will be used for business event integrations)
	emailTriggerService := email.NewEmailTriggerService(emailService, db)

	// Initialize email handler
	emailHandler := emailHandler.NewEmailHandler(emailService, db)

	// Start email queue processor in background
	go func() {
		log.Printf("🚀 EMAIL: Starting email queue processor...")
		for {
			// Dequeue email from queue
			email, err := emailQueue.Dequeue()
			if err != nil {
				log.Printf("❌ EMAIL: Queue dequeue error: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}
			if email == nil {
				// Queue is empty, wait a bit before checking again
				time.Sleep(2 * time.Second)
				continue
			}

			log.Printf("📧 EMAIL: Processing email ID: %d, Subject: %s, To: %s",
				email.ID, email.Subject, email.Recipients[0].Email)

			// Send email via provider
			err = emailProvider.SendEmail(email)
			if err != nil {
				log.Printf("❌ EMAIL: Failed to send email ID %d: %v", email.ID, err)
				// Mark as failed in queue
				if err := emailQueue.MarkAsFailed(fmt.Sprintf("%d", email.ID), err.Error()); err != nil {
					log.Printf("❌ EMAIL: Failed to mark email as failed: %v", err)
				}
			} else {
				log.Printf("✅ EMAIL: Successfully sent email ID %d to %s",
					email.ID, email.Recipients[0].Email)
				// Mark as processed in queue
				if err := emailQueue.MarkAsProcessed(fmt.Sprintf("%d", email.ID)); err != nil {
					log.Printf("❌ EMAIL: Failed to mark email as processed: %v", err)
				}
			}
		}
	}()

	// Start email retry worker in background
	go func() {
		log.Printf("🔄 EMAIL: Starting email retry worker...")
		for {
			// Check for failed emails every 5 minutes
			time.Sleep(5 * time.Minute)

			if err := emailService.RetryFailedEmails(); err != nil {
				log.Printf("❌ EMAIL: Failed to retry emails: %v", err)
			}
		}
	}()

	routes.AppRoutes(r, db, gcsService, appwriteService, cfg, emailTriggerService)
	routes.SetupEmailRoutes(r, emailHandler)
	r.Run()
}
