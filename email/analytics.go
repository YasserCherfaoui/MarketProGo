package email

import (
	"fmt"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"gorm.io/gorm"
)

// EmailAnalytics interface defines the contract for email analytics
type EmailAnalytics interface {
	TrackEmailSent(email *models.Email) error
	TrackEmailDelivered(emailID string) error
	TrackEmailOpened(emailID string) error
	TrackEmailClicked(emailID string, link string) error
	TrackEmailBounced(emailID string, reason string) error
	GetEmailMetrics(timeRange TimeRange) (*EmailMetrics, error)
}

// EmailAnalyticsImplementation implements EmailAnalytics
type EmailAnalyticsImplementation struct {
	db *gorm.DB
}

// NewEmailAnalytics creates a new email analytics instance
func NewEmailAnalytics(db *gorm.DB) *EmailAnalyticsImplementation {
	return &EmailAnalyticsImplementation{
		db: db,
	}
}

// TrackEmailSent tracks when an email is sent
func (a *EmailAnalyticsImplementation) TrackEmailSent(email *models.Email) error {
	// Update email status to sent
	now := time.Now()
	email.Status = models.EmailStatusSent
	email.SentAt = &now

	if err := a.db.Save(email).Error; err != nil {
		return fmt.Errorf("failed to track email sent: %w", err)
	}

	return nil
}

// TrackEmailDelivered tracks when an email is delivered
func (a *EmailAnalyticsImplementation) TrackEmailDelivered(emailID string) error {
	var email models.Email
	if err := a.db.Where("id = ?", emailID).First(&email).Error; err != nil {
		return fmt.Errorf("failed to get email: %w", err)
	}

	// Update email status to delivered
	now := time.Now()
	email.Status = models.EmailStatusDelivered
	email.DeliveredAt = &now

	if err := a.db.Save(&email).Error; err != nil {
		return fmt.Errorf("failed to track email delivered: %w", err)
	}

	return nil
}

// TrackEmailOpened tracks when an email is opened
func (a *EmailAnalyticsImplementation) TrackEmailOpened(emailID string) error {
	var email models.Email
	if err := a.db.Where("id = ?", emailID).First(&email).Error; err != nil {
		return fmt.Errorf("failed to get email: %w", err)
	}

	// Update email status to opened
	now := time.Now()
	email.Status = models.EmailStatusOpened
	email.OpenedAt = &now

	if err := a.db.Save(&email).Error; err != nil {
		return fmt.Errorf("failed to track email opened: %w", err)
	}

	return nil
}

// TrackEmailClicked tracks when an email link is clicked
func (a *EmailAnalyticsImplementation) TrackEmailClicked(emailID string, link string) error {
	var email models.Email
	if err := a.db.Where("id = ?", emailID).First(&email).Error; err != nil {
		return fmt.Errorf("failed to get email: %w", err)
	}

	// Update email status to clicked
	now := time.Now()
	email.Status = models.EmailStatusClicked
	email.ClickedAt = &now

	// Store clicked link in metadata
	if email.Metadata.IsNull() {
		email.Metadata = models.EmailJSON(`{"clicked_links": []}`)
	}

	// In a real implementation, you'd parse and update the metadata
	// For now, we'll just update the status

	if err := a.db.Save(&email).Error; err != nil {
		return fmt.Errorf("failed to track email clicked: %w", err)
	}

	return nil
}

// TrackEmailBounced tracks when an email bounces
func (a *EmailAnalyticsImplementation) TrackEmailBounced(emailID string, reason string) error {
	var email models.Email
	if err := a.db.Where("id = ?", emailID).First(&email).Error; err != nil {
		return fmt.Errorf("failed to get email: %w", err)
	}

	// Update email status to bounced
	now := time.Now()
	email.Status = models.EmailStatusBounced
	email.BouncedAt = &now
	email.BounceReason = reason

	if err := a.db.Save(&email).Error; err != nil {
		return fmt.Errorf("failed to track email bounced: %w", err)
	}

	return nil
}

// GetEmailMetrics retrieves email metrics for a time range
func (a *EmailAnalyticsImplementation) GetEmailMetrics(timeRange TimeRange) (*EmailMetrics, error) {
	// Query database for metrics
	var sentCount, deliveredCount, openedCount, clickedCount, bouncedCount int64

	// Count sent emails
	a.db.Model(&models.Email{}).Where("created_at BETWEEN ? AND ?", timeRange.Start, timeRange.End).Count(&sentCount)

	// Count delivered emails
	a.db.Model(&models.Email{}).Where("status = ? AND created_at BETWEEN ? AND ?", models.EmailStatusDelivered, timeRange.Start, timeRange.End).Count(&deliveredCount)

	// Count opened emails
	a.db.Model(&models.Email{}).Where("status = ? AND created_at BETWEEN ? AND ?", models.EmailStatusOpened, timeRange.Start, timeRange.End).Count(&openedCount)

	// Count clicked emails
	a.db.Model(&models.Email{}).Where("status = ? AND created_at BETWEEN ? AND ?", models.EmailStatusClicked, timeRange.Start, timeRange.End).Count(&clickedCount)

	// Count bounced emails
	a.db.Model(&models.Email{}).Where("status = ? AND created_at BETWEEN ? AND ?", models.EmailStatusBounced, timeRange.Start, timeRange.End).Count(&bouncedCount)

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
