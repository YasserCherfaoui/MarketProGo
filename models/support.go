package models

import (
	"time"

	"gorm.io/gorm"
)

// SupportTicket represents a help ticket created by users
type SupportTicket struct {
	gorm.Model
	UserID          uint           `json:"user_id"`
	User            *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	OrderID         *uint          `json:"order_id,omitempty"`
	Order           *Order         `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	Title           string         `json:"title" gorm:"not null"`
	Description     string         `json:"description" gorm:"type:text;not null"`
	Category        TicketCategory `json:"category" gorm:"type:varchar(50);not null"`
	Priority        TicketPriority `json:"priority" gorm:"type:varchar(20);default:'medium'"`
	Status          TicketStatus   `json:"status" gorm:"type:varchar(20);default:'open'"`
	AssignedTo      *uint          `json:"assigned_to,omitempty"`
	AssignedUser    *User          `json:"assigned_user,omitempty" gorm:"foreignKey:AssignedTo"`
	Resolution      string         `json:"resolution" gorm:"type:text"`
	ResolvedAt      *time.Time     `json:"resolved_at"`
	ResolvedBy      *uint          `json:"resolved_by,omitempty"`
	ResolvedByUser  *User          `json:"resolved_by_user,omitempty" gorm:"foreignKey:ResolvedBy"`
	InternalNotes   string         `json:"internal_notes" gorm:"type:text"`
	IsEscalated     bool           `json:"is_escalated" gorm:"default:false"`
	EscalatedAt     *time.Time     `json:"escalated_at"`
	EscalatedBy     *uint          `json:"escalated_by,omitempty"`
	EscalatedByUser *User          `json:"escalated_by_user,omitempty" gorm:"foreignKey:EscalatedBy"`

	// Attachments and responses
	Attachments []TicketAttachment `json:"attachments" gorm:"foreignKey:TicketID"`
	Responses   []TicketResponse   `json:"responses" gorm:"foreignKey:TicketID"`
}

// TicketCategory represents the category of a support ticket
type TicketCategory string

const (
	TicketCategoryGeneral   TicketCategory = "GENERAL"
	TicketCategoryOrder     TicketCategory = "ORDER"
	TicketCategoryPayment   TicketCategory = "PAYMENT"
	TicketCategoryProduct   TicketCategory = "PRODUCT"
	TicketCategoryShipping  TicketCategory = "SHIPPING"
	TicketCategoryReturn    TicketCategory = "RETURN"
	TicketCategoryTechnical TicketCategory = "TECHNICAL"
	TicketCategoryAccount   TicketCategory = "ACCOUNT"
	TicketCategoryBilling   TicketCategory = "BILLING"
	TicketCategoryOther     TicketCategory = "OTHER"
)

// TicketPriority represents the priority level of a ticket
type TicketPriority string

const (
	TicketPriorityLow    TicketPriority = "LOW"
	TicketPriorityMedium TicketPriority = "MEDIUM"
	TicketPriorityHigh   TicketPriority = "HIGH"
	TicketPriorityUrgent TicketPriority = "URGENT"
)

// TicketStatus represents the status of a ticket
type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "OPEN"
	TicketStatusInProgress TicketStatus = "IN_PROGRESS"
	TicketStatusWaiting    TicketStatus = "WAITING"
	TicketStatusResolved   TicketStatus = "RESOLVED"
	TicketStatusClosed     TicketStatus = "CLOSED"
)

// TicketAttachment represents files attached to a ticket
type TicketAttachment struct {
	gorm.Model
	TicketID uint           `json:"ticket_id"`
	Ticket   *SupportTicket `json:"-" gorm:"foreignKey:TicketID"`
	FileName string         `json:"file_name" gorm:"not null"`
	FileURL  string         `json:"file_url" gorm:"not null"`
	FileSize int64          `json:"file_size"`
	FileType string         `json:"file_type"`
}

// TicketResponse represents responses to a ticket
type TicketResponse struct {
	gorm.Model
	TicketID    uint           `json:"ticket_id"`
	Ticket      *SupportTicket `json:"-" gorm:"foreignKey:TicketID"`
	UserID      uint           `json:"user_id"`
	User        *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Message     string         `json:"message" gorm:"type:text;not null"`
	IsInternal  bool           `json:"is_internal" gorm:"default:false"`
	IsFromAdmin bool           `json:"is_from_admin" gorm:"default:false"`
}

// AbuseReport represents a report of abuse or inappropriate content
type AbuseReport struct {
	gorm.Model
	ReporterID     uint              `json:"reporter_id"`
	Reporter       *User             `json:"reporter,omitempty" gorm:"foreignKey:ReporterID"`
	ReportedUserID *uint             `json:"reported_user_id,omitempty"`
	ReportedUser   *User             `json:"reported_user,omitempty" gorm:"foreignKey:ReportedUserID"`
	ProductID      *uint             `json:"product_id,omitempty"`
	Product        *Product          `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	ReviewID       *uint             `json:"review_id,omitempty"`
	Review         *ProductReview    `json:"review,omitempty" gorm:"foreignKey:ReviewID"`
	OrderID        *uint             `json:"order_id,omitempty"`
	Order          *Order            `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	Category       AbuseCategory     `json:"category" gorm:"type:varchar(50);not null"`
	Description    string            `json:"description" gorm:"type:text;not null"`
	Status         AbuseReportStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Severity       AbuseSeverity     `json:"severity" gorm:"type:varchar(20);default:'medium'"`
	AssignedTo     *uint             `json:"assigned_to,omitempty"`
	AssignedUser   *User             `json:"assigned_user,omitempty" gorm:"foreignKey:AssignedTo"`
	Resolution     string            `json:"resolution" gorm:"type:text"`
	ResolvedAt     *time.Time        `json:"resolved_at"`
	ResolvedBy     *uint             `json:"resolved_by,omitempty"`
	ResolvedByUser *User             `json:"resolved_by_user,omitempty" gorm:"foreignKey:ResolvedBy"`
	InternalNotes  string            `json:"internal_notes" gorm:"type:text"`

	// Attachments
	Attachments []AbuseReportAttachment `json:"attachments" gorm:"foreignKey:AbuseReportID"`
}

// AbuseCategory represents the category of abuse
type AbuseCategory string

const (
	AbuseCategoryHarassment     AbuseCategory = "HARASSMENT"
	AbuseCategorySpam           AbuseCategory = "SPAM"
	AbuseCategoryInappropriate  AbuseCategory = "INAPPROPRIATE"
	AbuseCategoryFraud          AbuseCategory = "FRAUD"
	AbuseCategoryCopyright      AbuseCategory = "COPYRIGHT"
	AbuseCategoryViolence       AbuseCategory = "VIOLENCE"
	AbuseCategoryDiscrimination AbuseCategory = "DISCRIMINATION"
	AbuseCategoryOther          AbuseCategory = "OTHER"
)

// AbuseReportStatus represents the status of an abuse report
type AbuseReportStatus string

const (
	AbuseReportStatusPending   AbuseReportStatus = "PENDING"
	AbuseReportStatusReviewing AbuseReportStatus = "REVIEWING"
	AbuseReportStatusResolved  AbuseReportStatus = "RESOLVED"
	AbuseReportStatusDismissed AbuseReportStatus = "DISMISSED"
)

// AbuseSeverity represents the severity level of abuse
type AbuseSeverity string

const (
	AbuseSeverityLow      AbuseSeverity = "LOW"
	AbuseSeverityMedium   AbuseSeverity = "MEDIUM"
	AbuseSeverityHigh     AbuseSeverity = "HIGH"
	AbuseSeverityCritical AbuseSeverity = "CRITICAL"
)

// AbuseReportAttachment represents files attached to an abuse report
type AbuseReportAttachment struct {
	gorm.Model
	AbuseReportID uint         `json:"abuse_report_id"`
	AbuseReport   *AbuseReport `json:"-" gorm:"foreignKey:AbuseReportID"`
	FileName      string       `json:"file_name" gorm:"not null"`
	FileURL       string       `json:"file_url" gorm:"not null"`
	FileSize      int64        `json:"file_size"`
	FileType      string       `json:"file_type"`
}

// ContactInquiry represents a contact us form submission
type ContactInquiry struct {
	gorm.Model
	UserID          *uint           `json:"user_id,omitempty"`
	User            *User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Name            string          `json:"name" gorm:"not null"`
	Email           string          `json:"email" gorm:"not null"`
	Phone           string          `json:"phone"`
	Subject         string          `json:"subject" gorm:"not null"`
	Message         string          `json:"message" gorm:"type:text;not null"`
	Category        ContactCategory `json:"category" gorm:"type:varchar(50);not null"`
	Status          ContactStatus   `json:"status" gorm:"type:varchar(20);default:'new'"`
	Priority        ContactPriority `json:"priority" gorm:"type:varchar(20);default:'normal'"`
	AssignedTo      *uint           `json:"assigned_to,omitempty"`
	AssignedUser    *User           `json:"assigned_user,omitempty" gorm:"foreignKey:AssignedTo"`
	Response        string          `json:"response" gorm:"type:text"`
	RespondedAt     *time.Time      `json:"responded_at"`
	RespondedBy     *uint           `json:"responded_by,omitempty"`
	RespondedByUser *User           `json:"responded_by_user,omitempty" gorm:"foreignKey:RespondedBy"`
	InternalNotes   string          `json:"internal_notes" gorm:"type:text"`
}

// ContactCategory represents the category of contact inquiry
type ContactCategory string

const (
	ContactCategoryGeneral     ContactCategory = "GENERAL"
	ContactCategorySales       ContactCategory = "SALES"
	ContactCategorySupport     ContactCategory = "SUPPORT"
	ContactCategoryFeedback    ContactCategory = "FEEDBACK"
	ContactCategoryPartnership ContactCategory = "PARTNERSHIP"
	ContactCategoryPress       ContactCategory = "PRESS"
	ContactCategoryOther       ContactCategory = "OTHER"
)

// ContactStatus represents the status of a contact inquiry
type ContactStatus string

const (
	ContactStatusNew        ContactStatus = "NEW"
	ContactStatusInProgress ContactStatus = "IN_PROGRESS"
	ContactStatusResponded  ContactStatus = "RESPONDED"
	ContactStatusClosed     ContactStatus = "CLOSED"
)

// ContactPriority represents the priority of a contact inquiry
type ContactPriority string

const (
	ContactPriorityLow    ContactPriority = "LOW"
	ContactPriorityNormal ContactPriority = "NORMAL"
	ContactPriorityHigh   ContactPriority = "HIGH"
	ContactPriorityUrgent ContactPriority = "URGENT"
)

// Dispute represents a dispute submitted by users
type Dispute struct {
	gorm.Model
	UserID          uint            `json:"user_id"`
	User            *User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	OrderID         *uint           `json:"order_id,omitempty"`
	Order           *Order          `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	PaymentID       *uint           `json:"payment_id,omitempty"`
	Payment         *Payment        `json:"payment,omitempty" gorm:"foreignKey:PaymentID"`
	Title           string          `json:"title" gorm:"not null"`
	Description     string          `json:"description" gorm:"type:text;not null"`
	Category        DisputeCategory `json:"category" gorm:"type:varchar(50);not null"`
	Status          DisputeStatus   `json:"status" gorm:"type:varchar(20);default:'open'"`
	Priority        DisputePriority `json:"priority" gorm:"type:varchar(20);default:'medium'"`
	Amount          *float64        `json:"amount,omitempty"`
	Currency        string          `json:"currency" gorm:"default:'GBP'"`
	AssignedTo      *uint           `json:"assigned_to,omitempty"`
	AssignedUser    *User           `json:"assigned_user,omitempty" gorm:"foreignKey:AssignedTo"`
	Resolution      string          `json:"resolution" gorm:"type:text"`
	ResolvedAt      *time.Time      `json:"resolved_at"`
	ResolvedBy      *uint           `json:"resolved_by,omitempty"`
	ResolvedByUser  *User           `json:"resolved_by_user,omitempty" gorm:"foreignKey:ResolvedBy"`
	InternalNotes   string          `json:"internal_notes" gorm:"type:text"`
	IsEscalated     bool            `json:"is_escalated" gorm:"default:false"`
	EscalatedAt     *time.Time      `json:"escalated_at"`
	EscalatedBy     *uint           `json:"escalated_by,omitempty"`
	EscalatedByUser *User           `json:"escalated_by_user,omitempty" gorm:"foreignKey:EscalatedBy"`

	// Attachments and responses
	Attachments []DisputeAttachment `json:"attachments" gorm:"foreignKey:DisputeID"`
	Responses   []DisputeResponse   `json:"responses" gorm:"foreignKey:DisputeID"`
}

// DisputeCategory represents the category of a dispute
type DisputeCategory string

const (
	DisputeCategoryOrder    DisputeCategory = "ORDER"
	DisputeCategoryPayment  DisputeCategory = "PAYMENT"
	DisputeCategoryProduct  DisputeCategory = "PRODUCT"
	DisputeCategoryShipping DisputeCategory = "SHIPPING"
	DisputeCategoryRefund   DisputeCategory = "REFUND"
	DisputeCategoryBilling  DisputeCategory = "BILLING"
	DisputeCategoryService  DisputeCategory = "SERVICE"
	DisputeCategoryOther    DisputeCategory = "OTHER"
)

// DisputeStatus represents the status of a dispute
type DisputeStatus string

const (
	DisputeStatusOpen        DisputeStatus = "OPEN"
	DisputeStatusInProgress  DisputeStatus = "IN_PROGRESS"
	DisputeStatusUnderReview DisputeStatus = "UNDER_REVIEW"
	DisputeStatusResolved    DisputeStatus = "RESOLVED"
	DisputeStatusClosed      DisputeStatus = "CLOSED"
	DisputeStatusEscalated   DisputeStatus = "ESCALATED"
)

// DisputePriority represents the priority of a dispute
type DisputePriority string

const (
	DisputePriorityLow    DisputePriority = "LOW"
	DisputePriorityMedium DisputePriority = "MEDIUM"
	DisputePriorityHigh   DisputePriority = "HIGH"
	DisputePriorityUrgent DisputePriority = "URGENT"
)

// DisputeAttachment represents files attached to a dispute
type DisputeAttachment struct {
	gorm.Model
	DisputeID uint     `json:"dispute_id"`
	Dispute   *Dispute `json:"-" gorm:"foreignKey:DisputeID"`
	FileName  string   `json:"file_name" gorm:"not null"`
	FileURL   string   `json:"file_url" gorm:"not null"`
	FileSize  int64    `json:"file_size"`
	FileType  string   `json:"file_type"`
}

// DisputeResponse represents responses to a dispute
type DisputeResponse struct {
	gorm.Model
	DisputeID   uint     `json:"dispute_id"`
	Dispute     *Dispute `json:"-" gorm:"foreignKey:DisputeID"`
	UserID      uint     `json:"user_id"`
	User        *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Message     string   `json:"message" gorm:"type:text;not null"`
	IsInternal  bool     `json:"is_internal" gorm:"default:false"`
	IsFromAdmin bool     `json:"is_from_admin" gorm:"default:false"`
}
