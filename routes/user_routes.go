package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/handlers/user"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UserRoutes(router *gin.RouterGroup, db *gorm.DB) {
	userRouter := router.Group("/users")
	userHandler := user.NewUserHandler(db)

	// Public routes
	userRouter.POST("/seller", userHandler.CreateSeller)

	// Protected routes
	userRouter.Use(middlewares.AuthMiddleware())
	{
		userRouter.GET("", userHandler.GetAllUsers)
		userRouter.GET("/seller", userHandler.GetAllSellers)
		userRouter.DELETE("/:id", userHandler.DeleteUser)

		// Address routes
		addressRouter := userRouter.Group("/addresses")
		{
			addressRouter.POST("", userHandler.CreateAddress)
			addressRouter.GET("", userHandler.GetAddresses)
			addressRouter.GET("/:id", userHandler.GetAddress)
			addressRouter.PUT("/:id", userHandler.UpdateAddress)
			addressRouter.DELETE("/:id", userHandler.DeleteAddress)
			addressRouter.PUT("/:id/default", userHandler.SetDefaultAddress)
		}
	}
}
