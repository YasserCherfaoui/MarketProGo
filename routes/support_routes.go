package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/email"
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"github.com/YasserCherfaoui/MarketProGo/handlers/support"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SupportRoutes registers all support-related routes
func SupportRoutes(router *gin.RouterGroup, db *gorm.DB, gcsService *gcs.GCService, appwriteService *aw.AppwriteService, emailTriggerSvc *email.EmailTriggerService) {
	supportHandler := support.NewSupportHandler(db, gcsService, appwriteService, emailTriggerSvc)

	// Support tickets routes
	tickets := router.Group("/tickets", middlewares.AuthMiddleware())
	{
		tickets.POST("/", supportHandler.CreateTicket)
		tickets.GET("/", supportHandler.GetUserTickets)
		tickets.GET("/:id", supportHandler.GetTicket)
		tickets.PUT("/:id", supportHandler.UpdateTicket)
		tickets.DELETE("/:id", supportHandler.DeleteTicket)
		tickets.POST("/:id/responses", supportHandler.AddTicketResponse)
	}

	// Admin-only ticket routes
	adminTickets := router.Group("/admin/tickets", middlewares.AuthMiddleware())
	{
		adminTickets.GET("/", supportHandler.GetAllTickets)
	}

	// Abuse reports routes
	abuse := router.Group("/abuse", middlewares.AuthMiddleware())
	{
		abuse.POST("/reports", supportHandler.CreateAbuseReport)
		abuse.GET("/reports", supportHandler.GetUserAbuseReports)
		abuse.GET("/reports/:id", supportHandler.GetAbuseReport)
		abuse.PUT("/reports/:id", supportHandler.UpdateAbuseReport)
		abuse.DELETE("/reports/:id", supportHandler.DeleteAbuseReport)
	}

	// Admin-only abuse report routes
	adminAbuse := router.Group("/admin/abuse", middlewares.AuthMiddleware())
	{
		adminAbuse.GET("/reports", supportHandler.GetAllAbuseReports)
	}
	router.POST("/contact/inquiries", supportHandler.CreateContactInquiry)

	// Contact inquiries routes
	contact := router.Group("/contact", middlewares.AuthMiddleware())
	{
		contact.GET("/inquiries", supportHandler.GetUserContactInquiries)
		contact.GET("/inquiries/:id", supportHandler.GetContactInquiry)
		contact.PUT("/inquiries/:id", supportHandler.UpdateContactInquiry)
		contact.DELETE("/inquiries/:id", supportHandler.DeleteContactInquiry)
	}

	// Admin-only contact inquiry routes
	adminContact := router.Group("/admin/contact", middlewares.AuthMiddleware())
	{
		adminContact.GET("/inquiries", supportHandler.GetAllContactInquiries)
		adminContact.POST("/inquiries/:id/reply", supportHandler.ReplyToContactInquiry)
	}

	// Disputes routes
	disputes := router.Group("/disputes", middlewares.AuthMiddleware())
	{
		disputes.POST("/", supportHandler.CreateDispute)
		disputes.GET("/", supportHandler.GetUserDisputes)
		disputes.GET("/:id", supportHandler.GetDispute)
		disputes.PUT("/:id", supportHandler.UpdateDispute)
		disputes.DELETE("/:id", supportHandler.DeleteDispute)
		disputes.POST("/:id/responses", supportHandler.AddDisputeResponse)
	}

	// Admin-only dispute routes
	adminDisputes := router.Group("/admin/disputes", middlewares.AuthMiddleware())
	{
		adminDisputes.GET("/", supportHandler.GetAllDisputes)
	}
}
