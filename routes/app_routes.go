package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"github.com/YasserCherfaoui/MarketProGo/handlers/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AppRoutes(r *gin.Engine, db *gorm.DB, gcsService *gcs.GCService) {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router := r.Group("/api/v1")
	authHandler := auth.NewAuthHandler(db)
	AuthRoutes(router, authHandler)
	CategoryRoutes(router, db, gcsService)
	ProductRoutes(router, db, gcsService)
	UserRoutes(router, db)
}
