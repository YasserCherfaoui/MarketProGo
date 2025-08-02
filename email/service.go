package email

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/cfg"
	"github.com/YasserCherfaoui/MarketProGo/models"
	"gorm.io/gorm"
)

// EmailService interface defines the contract for email service operations
type EmailService interface {
	SendEmail(template string, data map[string]interface{}, recipient models.EmailRecipient) error
	SendBulkEmail(template string, data map[string]interface{}, recipients []models.EmailRecipient) error
	SendTransactionalEmail(emailType models.EmailType, data map[string]interface{}, recipient models.EmailRecipient) error
	GetEmailStatus(emailID string) (models.EmailStatus, error)
	RetryFailedEmail(emailID string) error
	RetryFailedEmails() error
	GetEmailMetrics(timeRange TimeRange) (*EmailMetrics, error)
	GetQueueSize() (int64, error)
}

// TimeRange represents a time range for metrics
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// EmailMetrics represents email performance metrics
type EmailMetrics struct {
	SentCount      int     `json:"sent_count"`
	DeliveredCount int     `json:"delivered_count"`
	OpenedCount    int     `json:"opened_count"`
	ClickedCount   int     `json:"clicked_count"`
	BouncedCount   int     `json:"bounced_count"`
	DeliveryRate   float64 `json:"delivery_rate"`
	OpenRate       float64 `json:"open_rate"`
	ClickRate      float64 `json:"click_rate"`
}

// EmailServiceImplementation implements EmailService
type EmailServiceImplementation struct {
	provider       EmailProvider
	templateEngine TemplateEngine
	queue          EmailQueue
	analytics      EmailAnalytics
	config         *cfg.EmailConfig
	db             *gorm.DB
}

// NewEmailService creates a new email service instance
func NewEmailService(
	provider EmailProvider,
	templateEngine TemplateEngine,
	queue EmailQueue,
	analytics EmailAnalytics,
	config *cfg.EmailConfig,
	db *gorm.DB,
) *EmailServiceImplementation {
	return &EmailServiceImplementation{
		provider:       provider,
		templateEngine: templateEngine,
		queue:          queue,
		analytics:      analytics,
		config:         config,
		db:             db,
	}
}

// SendEmail sends a single email
func (s *EmailServiceImplementation) SendEmail(template string, data map[string]interface{}, recipient models.EmailRecipient) error {
	// Render email content
	htmlContent, textContent, err := s.templateEngine.RenderTemplate(template, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Create email record
	email := &models.Email{
		Type:        models.EmailTypePromotional, // Default type, can be overridden
		Template:    template,
		Recipients:  []models.EmailRecipient{recipient},
		SenderEmail: s.config.SenderEmail,
		SenderName:  s.config.SenderName,
		Subject:     s.getSubjectFromData(data),
		HTMLContent: htmlContent,
		TextContent: textContent,
		Status:      models.EmailStatusPending,
		RetryCount:  0,
	}

	// Save email to database
	if err := s.db.Create(email).Error; err != nil {
		return fmt.Errorf("failed to save email to database: %w", err)
	}

	// Queue email for sending
	if err := s.queue.Enqueue(email); err != nil {
		return fmt.Errorf("failed to queue email: %w", err)
	}

	// Track email sent
	if err := s.analytics.TrackEmailSent(email); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to track email sent: %v\n", err)
	}

	return nil
}

// SendBulkEmail sends multiple emails in bulk
func (s *EmailServiceImplementation) SendBulkEmail(template string, data map[string]interface{}, recipients []models.EmailRecipient) error {
	// Render email content once
	htmlContent, textContent, err := s.templateEngine.RenderTemplate(template, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Create emails for each recipient
	var emails []*models.Email
	for _, recipient := range recipients {
		email := &models.Email{
			Type:        models.EmailTypePromotional,
			Template:    template,
			Recipients:  []models.EmailRecipient{recipient},
			SenderEmail: s.config.SenderEmail,
			SenderName:  s.config.SenderName,
			Subject:     s.getSubjectFromData(data),
			HTMLContent: htmlContent,
			TextContent: textContent,
			Status:      models.EmailStatusPending,
			RetryCount:  0,
		}

		// Save email to database
		if err := s.db.Create(email).Error; err != nil {
			return fmt.Errorf("failed to save email to database: %w", err)
		}

		emails = append(emails, email)
	}

	// Queue all emails
	for _, email := range emails {
		if err := s.queue.Enqueue(email); err != nil {
			return fmt.Errorf("failed to queue email: %w", err)
		}

		// Track email sent
		if err := s.analytics.TrackEmailSent(email); err != nil {
			fmt.Printf("Failed to track email sent: %v\n", err)
		}
	}

	return nil
}

// SendTransactionalEmail sends a transactional email
func (s *EmailServiceImplementation) SendTransactionalEmail(emailType models.EmailType, data map[string]interface{}, recipient models.EmailRecipient) error {
	// Get template name based on email type
	templateName := s.getTemplateNameForType(emailType)
	if templateName == "" {
		return fmt.Errorf("no template found for email type: %s", emailType)
	}

	// Render email content
	htmlContent, textContent, err := s.templateEngine.RenderTemplate(templateName, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Create email record
	email := &models.Email{
		Type:        emailType,
		Template:    templateName,
		Recipients:  []models.EmailRecipient{recipient},
		SenderEmail: s.config.SenderEmail,
		SenderName:  s.config.SenderName,
		Subject:     s.getSubjectFromData(data),
		HTMLContent: htmlContent,
		TextContent: textContent,
		Status:      models.EmailStatusPending,
		RetryCount:  0,
	}

	// Save email to database
	if err := s.db.Create(email).Error; err != nil {
		return fmt.Errorf("failed to save email to database: %w", err)
	}

	// Queue email for sending
	if err := s.queue.Enqueue(email); err != nil {
		return fmt.Errorf("failed to queue email: %w", err)
	}

	// Track email sent
	if err := s.analytics.TrackEmailSent(email); err != nil {
		fmt.Printf("Failed to track email sent: %v\n", err)
	}

	return nil
}

// GetEmailStatus retrieves the status of an email
func (s *EmailServiceImplementation) GetEmailStatus(emailID string) (models.EmailStatus, error) {
	var email models.Email
	if err := s.db.Where("id = ?", emailID).First(&email).Error; err != nil {
		return "", fmt.Errorf("failed to get email: %w", err)
	}

	return email.Status, nil
}

// RetryFailedEmail retries a failed email
func (s *EmailServiceImplementation) RetryFailedEmail(emailID string) error {
	var email models.Email
	if err := s.db.Where("id = ?", emailID).First(&email).Error; err != nil {
		return fmt.Errorf("failed to get email: %w", err)
	}

	// Check if email is in failed status
	if email.Status != models.EmailStatusFailed {
		return fmt.Errorf("email is not in failed status")
	}

	// Check retry limit (max 3 retries)
	if email.RetryCount >= 3 {
		return fmt.Errorf("email has exceeded maximum retry attempts")
	}

	// Update retry count
	email.RetryCount++
	email.Status = models.EmailStatusPending

	// Save updated email
	if err := s.db.Save(&email).Error; err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	// Re-queue email
	if err := s.queue.Enqueue(&email); err != nil {
		return fmt.Errorf("failed to re-queue email: %w", err)
	}

	return nil
}

// RetryFailedEmails retries all failed emails with exponential backoff
func (s *EmailServiceImplementation) RetryFailedEmails() error {
	var failedEmails []models.Email
	if err := s.db.Where("status = ? AND retry_count < ?", models.EmailStatusFailed, 3).Find(&failedEmails).Error; err != nil {
		return fmt.Errorf("failed to get failed emails: %w", err)
	}

	for _, email := range failedEmails {
		// Calculate exponential backoff delay
		delay := time.Duration(1<<email.RetryCount) * time.Minute // 1, 2, 4 minutes

		// Add jitter to prevent thundering herd
		jitter := time.Duration(rand.Intn(30)) * time.Second
		delay += jitter

		// Schedule retry after delay
		go func(email models.Email, delay time.Duration) {
			time.Sleep(delay)

			// Update retry count and status
			email.RetryCount++
			email.Status = models.EmailStatusPending

			if err := s.db.Save(&email).Error; err != nil {
				fmt.Printf("Failed to update email %d for retry: %v\n", email.ID, err)
				return
			}

			// Re-queue email
			if err := s.queue.Enqueue(&email); err != nil {
				fmt.Printf("Failed to re-queue email %d: %v\n", email.ID, err)
			}
		}(email, delay)
	}

	return nil
}

// GetEmailMetrics retrieves email metrics for a time range
func (s *EmailServiceImplementation) GetEmailMetrics(timeRange TimeRange) (*EmailMetrics, error) {
	// Query database for metrics
	var sentCount, deliveredCount, openedCount, clickedCount, bouncedCount int64

	// Count sent emails
	s.db.Model(&models.Email{}).Where("created_at BETWEEN ? AND ?", timeRange.Start, timeRange.End).Count(&sentCount)

	// Count delivered emails
	s.db.Model(&models.Email{}).Where("status = ? AND created_at BETWEEN ? AND ?", models.EmailStatusDelivered, timeRange.Start, timeRange.End).Count(&deliveredCount)

	// Count opened emails
	s.db.Model(&models.Email{}).Where("status = ? AND created_at BETWEEN ? AND ?", models.EmailStatusOpened, timeRange.Start, timeRange.End).Count(&openedCount)

	// Count clicked emails
	s.db.Model(&models.Email{}).Where("status = ? AND created_at BETWEEN ? AND ?", models.EmailStatusClicked, timeRange.Start, timeRange.End).Count(&clickedCount)

	// Count bounced emails
	s.db.Model(&models.Email{}).Where("status = ? AND created_at BETWEEN ? AND ?", models.EmailStatusBounced, timeRange.Start, timeRange.End).Count(&bouncedCount)

	// Calculate rates
	var deliveryRate, openRate, clickRate float64
	if sentCount > 0 {
		deliveryRate = float64(deliveredCount) / float64(sentCount) * 100
		openRate = float64(openedCount) / float64(sentCount) * 100
		clickRate = float64(clickedCount) / float64(sentCount) * 100
	}

	return &EmailMetrics{
		SentCount:      int(sentCount),
		DeliveredCount: int(deliveredCount),
		OpenedCount:    int(openedCount),
		ClickedCount:   int(clickedCount),
		BouncedCount:   int(bouncedCount),
		DeliveryRate:   deliveryRate,
		OpenRate:       openRate,
		ClickRate:      clickRate,
	}, nil
}

// GetQueueSize retrieves the current size of the email queue
func (s *EmailServiceImplementation) GetQueueSize() (int64, error) {
	return s.queue.GetQueueSize()
}

// getSubjectFromData extracts subject from template data
func (s *EmailServiceImplementation) getSubjectFromData(data map[string]interface{}) string {
	if subject, ok := data["subject"].(string); ok {
		return subject
	}
	return "Algeria Market - Important Information"
}

// getTemplateNameForType returns the template name for a given email type
func (s *EmailServiceImplementation) getTemplateNameForType(emailType models.EmailType) string {
	switch emailType {
	case models.EmailTypePasswordReset:
		return "password_reset"
	case models.EmailTypeWelcome:
		return "welcome"
	case models.EmailTypeOrderConfirmation:
		return "order_confirmation"
	case models.EmailTypeOrderStatusUpdate:
		return "order_status_update"
	case models.EmailTypePaymentSuccess:
		return "payment_success"
	case models.EmailTypePaymentFailed:
		return "payment_failed"
	case models.EmailTypePromotional:
		return "promotional"
	case models.EmailTypeCartRecovery:
		return "cart_recovery"
	case models.EmailTypeSecurityAlert:
		return "security_alert"
	case models.EmailTypeAdminNotification:
		return "admin_notification"
	default:
		return ""
	}
}
