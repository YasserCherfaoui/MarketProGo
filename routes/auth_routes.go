package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/handlers/auth"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.RouterGroup, h *auth.AuthHandler) {

	auth := router.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/register", h.CreateUser)
	}
	protectedAuth := auth.Use(middlewares.AuthMiddleware())
	{
		protectedAuth.GET("/me", h.GetUser)
	}
}
