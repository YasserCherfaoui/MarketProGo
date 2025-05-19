package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"github.com/YasserCherfaoui/MarketProGo/handlers/carousel"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CarouselRoutes(router *gin.RouterGroup, db *gorm.DB, gcsService *gcs.GCService) {
	carouselHandler := carousel.NewCarouselHandler(db, gcsService)
	carouselRouter := router.Group("/carousel")
	{
		carouselRouter.GET("/", carouselHandler.GetCarousel)
	}

	carouselRouter.Use(middlewares.AuthMiddleware())
	{
		carouselRouter.POST("/", carouselHandler.CreateCarousel)
	}

}
