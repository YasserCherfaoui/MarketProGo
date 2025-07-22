package routes

import (
	fileHandler "github.com/YasserCherfaoui/MarketProGo/handlers/file"

	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/cfg"
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"github.com/YasserCherfaoui/MarketProGo/handlers/auth"
	"github.com/YasserCherfaoui/MarketProGo/handlers/inventory"
	"github.com/YasserCherfaoui/MarketProGo/handlers/payment"
	"github.com/YasserCherfaoui/MarketProGo/handlers/promotion"
	"github.com/YasserCherfaoui/MarketProGo/handlers/review"
	paymentService "github.com/YasserCherfaoui/MarketProGo/payment"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AppRoutes(r *gin.Engine, db *gorm.DB, gcsService *gcs.GCService, appwriteService *aw.AppwriteService, config *cfg.AppConfig) {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router := r.Group("/api/v1")
	authHandler := auth.NewAuthHandler(db)
	inventoryHandler := inventory.NewInventoryHandler(db, gcsService, appwriteService)

	AuthRoutes(router, authHandler)
	CategoryRoutes(router, db, gcsService, appwriteService)
	BrandRoutes(router, db, gcsService, appwriteService)
	ProductRoutes(router, db, gcsService, appwriteService)
	UserRoutes(router, db)
	CarouselRoutes(router, db, gcsService, appwriteService)
	CartRoutes(router, db)
	OrderRoutes(router, db)
	InventoryRoutes(router, inventoryHandler)

	// Register Promotion routes
	promotionHandler := promotion.NewPromotionHandler(db, gcsService, appwriteService)
	RegisterPromotionRoutes(router, promotionHandler)

	// Register Review routes
	reviewHandler := review.NewReviewHandler(db, appwriteService)
	RegisterReviewRoutes(router, reviewHandler)

	// Register Payment routes
	revolutPaymentService := paymentService.NewRevolutPaymentService(db, &config.Revolut)
	paymentHandler := payment.NewPaymentHandler(db, revolutPaymentService)
	SetupPaymentRoutes(r, paymentHandler)

	router.GET("/file/preview/:fileId", fileHandler.ProxyFilePreview)
}
