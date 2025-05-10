package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/handlers/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AppRoutes(r *gin.Engine, db *gorm.DB) {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router := r.Group("/api/v1")
	authHandler := auth.NewAuthHandler(db)
	AuthRoutes(router, authHandler)
	CategoryRoutes(router, db)
	ProductRoutes(router, db)
	UserRoutes(router, db)
}
