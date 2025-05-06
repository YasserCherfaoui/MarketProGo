package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AppRoutes(r *gin.Engine, db *gorm.DB) {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
}
