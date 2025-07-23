package payment

import (
	"context"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
)

// CustomerInfo represents customer information for payment processing
type CustomerInfo struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Phone    string `json:"phone,omitempty"`
	Address  string `json:"address,omitempty"`
	City     string `json:"city,omitempty"`
	Country  string `json:"country,omitempty"`
	PostCode string `json:"post_code,omitempty"`
}

// PaymentRequest represents a request to create a payment
type PaymentRequest struct {
	OrderID      uint              `json:"order_id"`
	Amount       float64           `json:"amount"`
	Currency     string            `json:"currency"`
	Description  string            `json:"description"`
	CustomerInfo *CustomerInfo     `json:"customer_info"`
	ReturnURL    string            `json:"return_url,omitempty"`
	CancelURL    string            `json:"cancel_url,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// PaymentResponse represents a response from creating a payment
type PaymentResponse struct {
	PaymentID     string     `json:"payment_id"`
	OrderID       string     `json:"order_id"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	Status        string     `json:"status"`
	CheckoutURL   string     `json:"checkout_url,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	PaymentMethod string     `json:"payment_method,omitempty"`
}

// RefundRequest represents a request to refund a payment
type RefundRequest struct {
	PaymentID string            `json:"payment_id"`
	Amount    float64           `json:"amount"`
	Reason    string            `json:"reason,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// RefundResponse represents a response from refunding a payment
type RefundResponse struct {
	RefundID  string    `json:"refund_id"`
	PaymentID string    `json:"payment_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Reason    string    `json:"reason,omitempty"`
}

// PaymentService defines the interface for payment operations
type PaymentService interface {
	// CreatePayment creates a new payment and returns payment details
	CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)

	// GetPaymentStatus retrieves the current status of a payment
	GetPaymentStatus(ctx context.Context, paymentID string) (string, error)

	// CapturePayment captures an authorized payment
	CapturePayment(ctx context.Context, paymentID string) error

	// RefundPayment refunds a payment
	RefundPayment(ctx context.Context, req *RefundRequest) (*RefundResponse, error)

	// CancelPayment cancels a pending payment
	CancelPayment(ctx context.Context, paymentID string) error

	// HandleWebhook processes webhook notifications from the payment provider
	HandleWebhook(ctx context.Context, payload []byte, signature string, timestamp string) error

	// GetPayment retrieves payment details by ID
	GetPayment(ctx context.Context, paymentID string) (*models.Payment, error)

	// ListPayments retrieves a list of payments with optional filtering
	ListPayments(ctx context.Context, orderID *uint, status *string, limit, offset int) ([]*models.Payment, int64, error)
}

// PaymentEvent represents a payment event for logging
type PaymentEvent struct {
	PaymentID string                 `json:"payment_id"`
	Event     string                 `json:"event"`
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// PaymentLogger defines the interface for payment event logging
type PaymentLogger interface {
	LogEvent(ctx context.Context, event *PaymentEvent) error
	GetPaymentEvents(ctx context.Context, paymentID string) ([]*PaymentEvent, error)
}
