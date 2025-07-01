package routes

import (
	fileHandler "github.com/YasserCherfaoui/MarketProGo/handlers/file"

	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"github.com/YasserCherfaoui/MarketProGo/handlers/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AppRoutes(r *gin.Engine, db *gorm.DB, gcsService *gcs.GCService, appwriteService *aw.AppwriteService) {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router := r.Group("/api/v1")
	authHandler := auth.NewAuthHandler(db)
	AuthRoutes(router, authHandler)
	CategoryRoutes(router, db, gcsService, appwriteService)
	BrandRoutes(router, db, gcsService, appwriteService)
	ProductRoutes(router, db, gcsService, appwriteService)
	UserRoutes(router, db)
	CarouselRoutes(router, db, gcsService, appwriteService)
	CartRoutes(router, db)
	OrderRoutes(router, db)
	router.GET("/file/preview/:fileId", fileHandler.ProxyFilePreview)
}
