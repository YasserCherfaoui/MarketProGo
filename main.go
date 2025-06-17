package main

import (
	"context"
	"log"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/cfg"
	"github.com/YasserCherfaoui/MarketProGo/database"
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"github.com/YasserCherfaoui/MarketProGo/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := cfg.LoadConfig()
	if err != nil {
		log.Fatalf("FATAL: Could not load configuration: %v", err)
	}
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

	routes.AppRoutes(r, db, gcsService, appwriteService)
	r.Run()

}
