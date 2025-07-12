package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/handlers/promotion"
	"github.com/gin-gonic/gin"
)

func RegisterPromotionRoutes(r *gin.RouterGroup, handler *promotion.PromotionHandler) {
	r.POST("/promotions", handler.CreatePromotion)
	r.GET("/promotions", handler.GetAllPromotions)
	r.GET("/promotions/:id", handler.GetPromotion)
	r.PUT("/promotions/:id", handler.UpdatePromotion)
	r.DELETE("/promotions/:id", handler.DeletePromotion)
}
