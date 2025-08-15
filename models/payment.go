package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// RevolutPaymentStatus represents the status of a Revolut payment
type RevolutPaymentStatus string

const (
	RevolutPaymentStatusPending    RevolutPaymentStatus = "PENDING"
	RevolutPaymentStatusAuthorized RevolutPaymentStatus = "AUTHORIZED"
	RevolutPaymentStatusCompleted  RevolutPaymentStatus = "COMPLETED"
	RevolutPaymentStatusFailed     RevolutPaymentStatus = "FAILED"
	RevolutPaymentStatusCancelled  RevolutPaymentStatus = "CANCELLED"
	RevolutPaymentStatusRefunded   RevolutPaymentStatus = "REFUNDED"
	RevolutPaymentStatusDisputed   RevolutPaymentStatus = "DISPUTED"
)

// JSON is a custom type for storing JSON data
type JSON map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

// Payment represents a payment transaction
type Payment struct {
	gorm.Model
	OrderID          uint                 `json:"order_id" gorm:"not null"`
	Order            Order                `json:"order" gorm:"foreignKey:OrderID"`
	RevolutOrderID   string               `json:"revolut_order_id" gorm:"uniqueIndex"`
	RevolutPaymentID string               `json:"revolut_payment_id" gorm:"uniqueIndex"`
	Amount           float64              `json:"amount" gorm:"not null"`
	Currency         string               `json:"currency" gorm:"not null;default:'GBP'"`
	Status           RevolutPaymentStatus `json:"status" gorm:"type:varchar(20);not null;default:'PENDING'"`
	PaymentMethod    string               `json:"payment_method"`
	CustomerID       string               `json:"customer_id"`
	CheckoutURL      string               `json:"checkout_url"`
	CompletedAt      *time.Time           `json:"completed_at"`
	FailureReason    string               `json:"failure_reason"`
	RefundStatus     string               `json:"refund_status"`
	RefundedAmount   float64              `json:"refunded_amount" gorm:"default:0"`
	Metadata         JSON                 `json:"metadata" gorm:"type:json"`

	// Audit fields
	CreatedBy uint `json:"created_by"`
	UpdatedBy uint `json:"updated_by"`
}

// TableName specifies the table name for Payment
func (Payment) TableName() string {
	return "payments"
}

// BeforeCreate is a GORM hook that runs before creating a payment
func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.Status == "" {
		p.Status = RevolutPaymentStatusPending
	}
	if p.Currency == "" {
		p.Currency = "GBP"
	}
	if p.RefundedAmount == 0 {
		p.RefundedAmount = 0
	}
	return nil
}

// IsCompleted returns true if the payment is completed
func (p *Payment) IsCompleted() bool {
	return p.Status == RevolutPaymentStatusCompleted
}

// IsFailed returns true if the payment failed
func (p *Payment) IsFailed() bool {
	return p.Status == RevolutPaymentStatusFailed
}

// IsRefunded returns true if the payment is refunded
func (p *Payment) IsRefunded() bool {
	return p.Status == RevolutPaymentStatusRefunded
}

// CanRefund returns true if the payment can be refunded
func (p *Payment) CanRefund() bool {
	return p.IsCompleted() && !p.IsRefunded() && p.RefundedAmount < p.Amount
}

// GetRefundableAmount returns the amount that can be refunded
func (p *Payment) GetRefundableAmount() float64 {
	if !p.CanRefund() {
		return 0
	}
	return p.Amount - p.RefundedAmount
}

// PaymentLog represents a log entry for payment events
type PaymentLog struct {
	gorm.Model
	PaymentID uint                 `json:"payment_id" gorm:"not null"`
	Payment   Payment              `json:"payment" gorm:"foreignKey:PaymentID"`
	Event     string               `json:"event" gorm:"not null"`
	OldStatus RevolutPaymentStatus `json:"old_status"`
	NewStatus RevolutPaymentStatus `json:"new_status"`
	Message   string               `json:"message"`
	Metadata  JSON                 `json:"metadata" gorm:"type:json"`
	CreatedBy uint                 `json:"created_by"`
}

// TableName specifies the table name for PaymentLog
func (PaymentLog) TableName() string {
	return "payment_logs"
}
