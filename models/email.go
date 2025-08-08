package models

import (
	"time"

	"gorm.io/gorm"
)

// Email represents an email record in the database
type Email struct {
	gorm.Model
	Type         EmailType        `json:"type"`
	Template     string           `json:"template"`
	Recipients   []EmailRecipient `json:"recipients" gorm:"-"`
	SenderEmail  string           `json:"sender_email" gorm:"default:'enquirees@algeriamarket.co.uk'"`
	SenderName   string           `json:"sender_name" gorm:"default:'Algeria Market'"`
	Subject      string           `json:"subject"`
	HTMLContent  string           `json:"html_content"`
	TextContent  string           `json:"text_content"`
	Status       EmailStatus      `json:"status"`
	ProviderID   string           `json:"provider_id"`
	SentAt       *time.Time       `json:"sent_at"`
	DeliveredAt  *time.Time       `json:"delivered_at"`
	OpenedAt     *time.Time       `json:"opened_at"`
	ClickedAt    *time.Time       `json:"clicked_at"`
	BouncedAt    *time.Time       `json:"bounced_at"`
	BounceReason string           `json:"bounce_reason"`
	RetryCount   int              `json:"retry_count"`
	Metadata     EmailJSON        `json:"metadata"`
}

// EmailRecipient represents an email recipient
type EmailRecipient struct {
	Email  string `json:"email"`
	Name   string `json:"name"`
	UserID *uint  `json:"user_id"`
	User   *User  `json:"user,omitempty"`
}

// EmailType represents the type of email
type EmailType string

const (
	EmailTypePasswordReset          EmailType = "password_reset"
	EmailTypeWelcome                EmailType = "welcome"
	EmailTypeOrderConfirmation      EmailType = "order_confirmation"
	EmailTypeOrderStatusUpdate      EmailType = "order_status_update"
	EmailTypePaymentSuccess         EmailType = "payment_success"
	EmailTypePaymentFailed          EmailType = "payment_failed"
	EmailTypePromotional            EmailType = "promotional"
	EmailTypeCartRecovery           EmailType = "cart_recovery"
	EmailTypeSecurityAlert          EmailType = "security_alert"
	EmailTypeAdminNotification      EmailType = "admin_notification"
	EmailTypeContactInquiryResponse EmailType = "contact_inquiry_response"
	EmailTypeContactStatusUpdated   EmailType = "contact_status_updated"
	EmailTypeTicketResponse         EmailType = "ticket_response"
	EmailTypeTicketStatusUpdated    EmailType = "ticket_status_updated"
	EmailTypeDisputeResponse        EmailType = "dispute_response"
	EmailTypeDisputeStatusUpdated   EmailType = "dispute_status_updated"
	EmailTypeAbuseStatusUpdated     EmailType = "abuse_status_updated"
)

// EmailStatus represents the status of an email
type EmailStatus string

const (
	EmailStatusPending   EmailStatus = "pending"
	EmailStatusSent      EmailStatus = "sent"
	EmailStatusDelivered EmailStatus = "delivered"
	EmailStatusOpened    EmailStatus = "opened"
	EmailStatusClicked   EmailStatus = "clicked"
	EmailStatusBounced   EmailStatus = "bounced"
	EmailStatusFailed    EmailStatus = "failed"
)

// EmailTemplate represents an email template
type EmailTemplate struct {
	gorm.Model
	Name        string    `json:"name"`
	Type        EmailType `json:"type"`
	Subject     string    `json:"subject"`
	HTMLContent string    `json:"html_content"`
	TextContent string    `json:"text_content"`
	Version     int       `json:"version"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	Metadata    EmailJSON `json:"metadata"`
}

// EmailJSON is a custom type for storing JSON data in GORM for email models
type EmailJSON []byte

// Value implements the driver.Valuer interface
func (j EmailJSON) Value() (interface{}, error) {
	if j.IsNull() {
		return nil, nil
	}
	return string(j), nil
}

// Scan implements the sql.Scanner interface
func (j *EmailJSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	s, ok := value.([]byte)
	if !ok {
		return nil
	}
	*j = append((*j)[0:0], s...)
	return nil
}

// MarshalJSON implements json.Marshaler interface
func (j EmailJSON) MarshalJSON() ([]byte, error) {
	if j.IsNull() {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON implements json.Unmarshaler interface
func (j *EmailJSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return nil
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// IsNull checks if the JSON is null
func (j EmailJSON) IsNull() bool {
	return len(j) == 0 || string(j) == "null"
}

// String returns the string representation of the JSON
func (j EmailJSON) String() string {
	return string(j)
}
