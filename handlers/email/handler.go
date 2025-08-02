package email

import (
	"strconv"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/email"
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Import email service types

// EmailService interface for dependency injection
type EmailService interface {
	SendEmail(template string, data map[string]interface{}, recipient models.EmailRecipient) error
	SendBulkEmail(template string, data map[string]interface{}, recipients []models.EmailRecipient) error
	SendTransactionalEmail(emailType models.EmailType, data map[string]interface{}, recipient models.EmailRecipient) error
	GetEmailStatus(emailID string) (models.EmailStatus, error)
	RetryFailedEmail(emailID string) error
	RetryFailedEmails() error
	GetEmailMetrics(timeRange email.TimeRange) (*email.EmailMetrics, error)
	GetQueueSize() (int64, error)
}

// EmailHandler handles email-related HTTP requests
type EmailHandler struct {
	emailService EmailService
	db           *gorm.DB
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(emailService EmailService, db *gorm.DB) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
		db:           db,
	}
}

// SendEmailRequest represents the request body for sending an email
type SendEmailRequest struct {
	Template  string                 `json:"template" binding:"required"`
	Data      map[string]interface{} `json:"data" binding:"required"`
	Recipient models.EmailRecipient  `json:"recipient" binding:"required"`
}

// SendBulkEmailRequest represents the request body for sending bulk emails
type SendBulkEmailRequest struct {
	Template   string                  `json:"template" binding:"required"`
	Data       map[string]interface{}  `json:"data" binding:"required"`
	Recipients []models.EmailRecipient `json:"recipients" binding:"required"`
}

// SendTransactionalEmailRequest represents the request body for sending transactional emails
type SendTransactionalEmailRequest struct {
	EmailType models.EmailType       `json:"email_type" binding:"required"`
	Data      map[string]interface{} `json:"data" binding:"required"`
	Recipient models.EmailRecipient  `json:"recipient" binding:"required"`
}

// EmailMetricsRequest represents the request body for getting email metrics
type EmailMetricsRequest struct {
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
}

// SendEmail handles sending a single email
func (h *EmailHandler) SendEmail(c *gin.Context) {
	var req SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate recipient email
	if req.Recipient.Email == "" {
		response.GenerateBadRequestResponse(c, "MISSING_EMAIL", "Recipient email is required")
		return
	}

	// Send email
	err := h.emailService.SendEmail(req.Template, req.Data, req.Recipient)
	if err != nil {
		response.GenerateInternalServerErrorResponse(c, "EMAIL_SEND_FAILED", "Failed to send email")
		return
	}

	response.GenerateSuccessResponse(c, "Email sent successfully", gin.H{
		"message": "Email has been queued for delivery",
	})
}

// SendBulkEmail handles sending bulk emails
func (h *EmailHandler) SendBulkEmail(c *gin.Context) {
	var req SendBulkEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate recipients
	if len(req.Recipients) == 0 {
		response.GenerateBadRequestResponse(c, "MISSING_RECIPIENTS", "At least one recipient is required")
		return
	}

	// Validate recipient emails
	for _, recipient := range req.Recipients {
		if recipient.Email == "" {
			response.GenerateBadRequestResponse(c, "INVALID_EMAIL", "All recipients must have valid email addresses")
			return
		}
	}

	// Send bulk emails
	err := h.emailService.SendBulkEmail(req.Template, req.Data, req.Recipients)
	if err != nil {
		response.GenerateInternalServerErrorResponse(c, "BULK_EMAIL_FAILED", "Failed to send bulk emails")
		return
	}

	response.GenerateSuccessResponse(c, "Bulk emails sent successfully", gin.H{
		"message": "Bulk emails have been queued for delivery",
		"count":   len(req.Recipients),
	})
}

// SendTransactionalEmail handles sending transactional emails
func (h *EmailHandler) SendTransactionalEmail(c *gin.Context) {
	var req SendTransactionalEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate recipient email
	if req.Recipient.Email == "" {
		response.GenerateBadRequestResponse(c, "MISSING_EMAIL", "Recipient email is required")
		return
	}

	// Send transactional email
	err := h.emailService.SendTransactionalEmail(req.EmailType, req.Data, req.Recipient)
	if err != nil {
		response.GenerateInternalServerErrorResponse(c, "TRANSACTIONAL_EMAIL_FAILED", "Failed to send transactional email")
		return
	}

	response.GenerateSuccessResponse(c, "Transactional email sent successfully", gin.H{
		"message": "Transactional email has been queued for delivery",
		"type":    req.EmailType,
	})
}

// GetEmailStatus retrieves the status of an email
func (h *EmailHandler) GetEmailStatus(c *gin.Context) {
	emailID := c.Param("id")
	if emailID == "" {
		response.GenerateBadRequestResponse(c, "MISSING_EMAIL_ID", "Email ID is required")
		return
	}

	status, err := h.emailService.GetEmailStatus(emailID)
	if err != nil {
		response.GenerateInternalServerErrorResponse(c, "EMAIL_STATUS_FAILED", "Failed to get email status")
		return
	}

	response.GenerateSuccessResponse(c, "Email status retrieved successfully", gin.H{
		"email_id": emailID,
		"status":   status,
	})
}

// RetryFailedEmail retries a failed email
func (h *EmailHandler) RetryFailedEmail(c *gin.Context) {
	emailID := c.Param("id")
	if emailID == "" {
		response.GenerateBadRequestResponse(c, "MISSING_EMAIL_ID", "Email ID is required")
		return
	}

	err := h.emailService.RetryFailedEmail(emailID)
	if err != nil {
		response.GenerateInternalServerErrorResponse(c, "EMAIL_RETRY_FAILED", "Failed to retry email")
		return
	}

	response.GenerateSuccessResponse(c, "Email retry initiated successfully", gin.H{
		"email_id": emailID,
		"message":  "Email has been queued for retry",
	})
}

// RetryAllFailedEmails retries all failed emails
func (h *EmailHandler) RetryAllFailedEmails(c *gin.Context) {
	err := h.emailService.RetryFailedEmails()
	if err != nil {
		response.GenerateInternalServerErrorResponse(c, "EMAIL_RETRY_FAILED", "Failed to retry emails")
		return
	}

	response.GenerateSuccessResponse(c, "Failed emails retry initiated successfully", gin.H{
		"message": "All failed emails have been queued for retry with exponential backoff",
	})
}

// GetEmailMetrics retrieves email metrics for a time range
func (h *EmailHandler) GetEmailMetrics(c *gin.Context) {
	var req EmailMetricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		response.GenerateBadRequestResponse(c, "INVALID_START_DATE", "Invalid start date format. Use YYYY-MM-DD")
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		response.GenerateBadRequestResponse(c, "INVALID_END_DATE", "Invalid end date format. Use YYYY-MM-DD")
		return
	}

	// Set end date to end of day
	endDate = endDate.Add(24*time.Hour - time.Second)

	timeRange := email.TimeRange{
		Start: startDate,
		End:   endDate,
	}

	metrics, err := h.emailService.GetEmailMetrics(timeRange)
	if err != nil {
		response.GenerateInternalServerErrorResponse(c, "METRICS_FAILED", "Failed to get email metrics")
		return
	}

	response.GenerateSuccessResponse(c, "Email metrics retrieved successfully", gin.H{
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
		"metrics":    metrics,
	})
}

// GetEmailList retrieves a list of emails with pagination
func (h *EmailHandler) GetEmailList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	emailType := c.Query("type")

	offset := (page - 1) * limit

	var emails []models.Email
	query := h.db.Model(&models.Email{})

	// Apply filters
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if emailType != "" {
		query = query.Where("type = ?", emailType)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get emails with pagination
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&emails).Error
	if err != nil {
		response.GenerateInternalServerErrorResponse(c, "EMAIL_LIST_FAILED", "Failed to get email list")
		return
	}

	response.GenerateSuccessResponse(c, "Email list retrieved successfully", gin.H{
		"emails": emails,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetEmailTemplates retrieves available email templates
func (h *EmailHandler) GetEmailTemplates(c *gin.Context) {
	// For now, return a static list of available templates
	templates := []string{
		"welcome",
		"password_reset",
		"order_confirmation",
		"order_status_update",
		"payment_success",
		"payment_failed",
		"promotional",
		"cart_recovery",
		"security_alert",
		"admin_notification",
	}

	response.GenerateSuccessResponse(c, "Email templates retrieved successfully", gin.H{
		"templates": templates,
	})
}

// GetQueueStatus retrieves the current queue status
func (h *EmailHandler) GetQueueStatus(c *gin.Context) {
	// Get queue size
	queueSize, err := h.emailService.(*email.EmailServiceImplementation).GetQueueSize()
	if err != nil {
		response.GenerateInternalServerErrorResponse(c, "QUEUE_STATUS_FAILED", "Failed to get queue status")
		return
	}

	response.GenerateSuccessResponse(c, "Queue status retrieved successfully", gin.H{
		"queue_size": queueSize,
		"status":     "active",
	})
}

// TestDatabaseConnection tests if the database is working
func (h *EmailHandler) TestDatabaseConnection(c *gin.Context) {
	// Test database connection by trying to create a simple email record
	testEmail := &models.Email{
		Type:        models.EmailTypePromotional,
		Template:    "test",
		Recipients:  []models.EmailRecipient{{Email: "test@example.com", Name: "Test"}},
		SenderEmail: "test@example.com",
		SenderName:  "Test Sender",
		Subject:     "Test Subject",
		HTMLContent: "<h1>Test</h1>",
		TextContent: "Test",
		Status:      models.EmailStatusPending,
	}

	if err := h.db.Create(testEmail).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "DB_TEST_FAILED", "Database test failed: "+err.Error())
		return
	}

	// Clean up the test record
	h.db.Delete(testEmail)

	response.GenerateSuccessResponse(c, "Database connection test successful", gin.H{
		"message": "Database is working properly",
	})
}
