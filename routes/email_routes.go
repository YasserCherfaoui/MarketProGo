package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/handlers/email"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
)

// SetupEmailRoutes sets up email-related routes
func SetupEmailRoutes(router *gin.Engine, emailHandler *email.EmailHandler) {
	// Email routes group
	emailGroup := router.Group("/api/v1/email")
	{
		// Public email endpoints (no authentication required)
		emailGroup.POST("/send", emailHandler.SendEmail)
		emailGroup.POST("/bulk", emailHandler.SendBulkEmail)
		emailGroup.POST("/transactional", emailHandler.SendTransactionalEmail)
		emailGroup.GET("/status/:id", emailHandler.GetEmailStatus)
		emailGroup.GET("/queue/status", emailHandler.GetQueueStatus)
		emailGroup.GET("/templates", emailHandler.GetEmailTemplates)
		emailGroup.GET("/test-db", emailHandler.TestDatabaseConnection)

		// Admin email management endpoints (require authentication)
		adminGroup := emailGroup.Group("/admin")
		adminGroup.Use(middlewares.AuthMiddleware())
		{
			adminGroup.GET("/list", emailHandler.GetEmailList)
			adminGroup.POST("/retry/:id", emailHandler.RetryFailedEmail)
			adminGroup.POST("/retry-all", emailHandler.RetryAllFailedEmails)
			adminGroup.POST("/metrics", emailHandler.GetEmailMetrics)
		}
	}
}
