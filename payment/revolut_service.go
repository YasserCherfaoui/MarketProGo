package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/cfg"
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/payment/revolut"
	"gorm.io/gorm"
)

// RevolutPaymentService implements PaymentService for Revolut
type RevolutPaymentService struct {
	client        *revolut.Client
	db            *gorm.DB
	webhookSecret string
	config        *cfg.RevolutConfig
}

// NewRevolutPaymentService creates a new Revolut payment service
func NewRevolutPaymentService(db *gorm.DB, config *cfg.RevolutConfig) *RevolutPaymentService {
	client := revolut.NewClient(config)

	return &RevolutPaymentService{
		client:        client,
		db:            db,
		webhookSecret: config.WebhookSecret,
		config:        config,
	}
}

// CreatePayment creates a new payment using Revolut
func (s *RevolutPaymentService) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if req.Amount <= 0 {
		return nil, fmt.Errorf("invalid amount: must be greater than 0")
	}
	if req.CustomerInfo == nil {
		return nil, fmt.Errorf("customer info is required")
	}

	// Validate Revolut configuration
	if s.config.APIKey == "" {
		return nil, fmt.Errorf("Revolut API key is not configured")
	}
	if s.config.BaseURL == "" {
		return nil, fmt.Errorf("Revolut base URL is not configured")
	}

	// Debug logging
	log.Printf("Creating Revolut payment for order %d, amount: %.2f %s", req.OrderID, req.Amount, req.Currency)
	log.Printf("Revolut config - BaseURL: %s, IsSandbox: %t", s.config.BaseURL, s.config.IsSandbox)
	log.Printf("API Key length: %d", len(s.config.APIKey))
	log.Printf("API Key prefix: %s", s.config.APIKey[:10]+"...")

	// Get order details
	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, req.OrderID).Error; err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Convert amount to minor units (cents) as required by Revolut API
	amountInMinorUnits := int64(req.Amount * 100)
	log.Printf("Converted amount: %d minor units", amountInMinorUnits)

	// Validate minimum amount (Revolut requires at least 1 cent)
	if amountInMinorUnits < 1 {
		return nil, fmt.Errorf("amount must be at least 0.01")
	}

	// Validate and normalize currency
	currency := req.Currency
	if currency == "" {
		currency = "GBP" // Default to GBP
	}
	// Ensure currency is uppercase
	currency = strings.ToUpper(currency)
	log.Printf("Using currency: %s", currency)

	// Validate description
	description := req.Description
	if description == "" {
		description = fmt.Sprintf("Order #%d", req.OrderID)
	}
	// Limit description length (Revolut might have limits)
	if len(description) > 255 {
		description = description[:252] + "..."
	}
	log.Printf("Using description: %s", description)

	// Create customer info for Revolut
	customer := &revolut.Customer{
		FullName: req.CustomerInfo.Name,
		Email:    req.CustomerInfo.Email,
		Phone:    req.CustomerInfo.Phone,
	}

	// Validate customer data
	if customer.FullName == "" {
		return nil, fmt.Errorf("customer full name is required")
	}
	if customer.Email == "" {
		return nil, fmt.Errorf("customer email is required")
	}

	log.Printf("Customer data: ID=%s, Name=%s, Email=%s, Phone=%s",
		customer.ID, customer.FullName, customer.Email, customer.Phone)

	// Create Revolut order request - simplified to avoid internal server errors
	revolutReq := &revolut.OrderRequest{
		Amount:           amountInMinorUnits,
		Currency:         currency,
		Description:      description,
		Customer:         customer,
		CaptureMode:      "automatic",
		EnforceChallenge: "automatic",
	}

	// Only add redirect URL if it's provided
	if req.ReturnURL != "" {
		revolutReq.RedirectURL = req.ReturnURL
	}

	// Only add metadata if it's not empty
	if req.Metadata != nil && len(req.Metadata) > 0 {
		revolutReq.Metadata = req.Metadata
	}

	// Debug: Log the request as JSON to see exactly what's being sent
	reqJSON, _ := json.MarshalIndent(revolutReq, "", "  ")
	log.Printf("Revolut order request JSON:\n%s", string(reqJSON))

	// Create order in Revolut
	revolutResp, err := s.client.CreateOrder(revolutReq)
	if err != nil {
		log.Printf("Revolut API error: %v", err)
		return nil, fmt.Errorf("failed to create Revolut order: %w", err)
	}

	log.Printf("Revolut order created successfully: %s", revolutResp.ID)

	// Create payment record in database
	payment := &models.Payment{
		OrderID:          req.OrderID,
		RevolutOrderID:   revolutResp.ID,
		RevolutPaymentID: revolutResp.ID, // The order ID from Revolut is actually the payment ID
		Amount:           req.Amount,
		Currency:         req.Currency,
		Status:           models.RevolutPaymentStatusPending,
		CustomerID:       strconv.FormatUint(uint64(req.CustomerInfo.ID), 10),
		CheckoutURL:      revolutResp.CheckoutURL,
		Metadata:         models.JSON(map[string]interface{}{}),
		CreatedBy:        req.CustomerInfo.ID,
	}

	if err := s.db.WithContext(ctx).Create(payment).Error; err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// Update order with Revolut information
	order.RevolutOrderID = revolutResp.ID
	order.CheckoutURL = revolutResp.CheckoutURL
	order.PaymentProvider = "revolut"

	if err := s.db.WithContext(ctx).Save(&order).Error; err != nil {
		log.Printf("Warning: failed to update order with Revolut info: %v", err)
	}

	// Log payment creation
	s.logPaymentEvent(ctx, payment.ID, "payment_created", "Payment created successfully", map[string]interface{}{
		"revolut_order_id": revolutResp.ID,
		"checkout_url":     revolutResp.CheckoutURL,
		"token":            revolutResp.Token,
	})

	return &PaymentResponse{
		PaymentID:   strconv.FormatUint(uint64(payment.ID), 10),
		OrderID:     revolutResp.ID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Status:      string(payment.Status),
		CheckoutURL: revolutResp.CheckoutURL,
		CreatedAt:   payment.CreatedAt,
	}, nil
}

// GetPaymentStatus retrieves the current status of a payment
func (s *RevolutPaymentService) GetPaymentStatus(ctx context.Context, paymentID string) (string, error) {
	// Get payment from database
	var payment models.Payment
	if err := s.db.WithContext(ctx).First(&payment, paymentID).Error; err != nil {
		return "", fmt.Errorf("payment not found: %w", err)
	}

	// If we have a Revolut order ID, check with Revolut API
	if payment.RevolutOrderID != "" {
		revolutOrder, err := s.client.GetOrder(payment.RevolutOrderID)
		if err != nil {
			log.Printf("Warning: failed to get Revolut order status: %v", err)
			// Return database status if API call fails
			return string(payment.Status), nil
		}

		// Update payment status if it has changed
		newStatus := s.mapRevolutStatusToPaymentStatus(revolutOrder.State)
		if newStatus != payment.Status {
			oldStatus := payment.Status
			payment.Status = newStatus

			if newStatus == models.RevolutPaymentStatusCompleted {
				now := time.Now()
				payment.CompletedAt = &now

				// The RevolutPaymentID is already set during order creation
				// No need to set it again here
			}

			if err := s.db.WithContext(ctx).Save(&payment).Error; err != nil {
				log.Printf("Warning: failed to update payment status: %v", err)
			} else {
				// Log status change
				s.logPaymentEvent(ctx, payment.ID, "status_changed", "Payment status updated", map[string]interface{}{
					"old_status":         oldStatus,
					"new_status":         newStatus,
					"revolut_state":      revolutOrder.State,
					"revolut_payment_id": payment.RevolutPaymentID,
				})
			}
		}
	}

	return string(payment.Status), nil
}

// CapturePayment captures an authorized payment
func (s *RevolutPaymentService) CapturePayment(ctx context.Context, paymentID string) error {
	// Get payment from database
	var payment models.Payment
	if err := s.db.WithContext(ctx).First(&payment, paymentID).Error; err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	if payment.RevolutPaymentID == "" {
		return fmt.Errorf("no Revolut payment ID available for capture")
	}

	// Capture payment in Revolut
	_, err := s.client.CaptureOrder(payment.RevolutOrderID)
	if err != nil {
		return fmt.Errorf("failed to capture payment: %w", err)
	}

	// Update payment status
	payment.Status = models.RevolutPaymentStatusCompleted
	now := time.Now()
	payment.CompletedAt = &now

	if err := s.db.WithContext(ctx).Save(&payment).Error; err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Log capture event
	s.logPaymentEvent(ctx, payment.ID, "payment_captured", "Payment captured successfully", nil)

	return nil
}

// RefundPayment refunds a payment
func (s *RevolutPaymentService) RefundPayment(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	// Get payment from database
	var payment models.Payment
	if err := s.db.WithContext(ctx).First(&payment, req.PaymentID).Error; err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	if !payment.CanRefund() {
		return nil, fmt.Errorf("payment cannot be refunded")
	}

	if req.Amount > payment.GetRefundableAmount() {
		return nil, fmt.Errorf("refund amount exceeds refundable amount")
	}

	// Create refund request
	revolutRefundReq := &revolut.RefundRequest{
		Amount:   req.Amount,
		Currency: payment.Currency,
		Reason:   req.Reason,
		Metadata: req.Metadata,
	}

	// Process refund in Revolut
	revolutResp, err := s.client.RefundPayment(payment.RevolutPaymentID, revolutRefundReq)
	if err != nil {
		return nil, fmt.Errorf("failed to process refund: %w", err)
	}

	// Update payment record
	payment.RefundedAmount += req.Amount
	if payment.RefundedAmount >= payment.Amount {
		payment.Status = models.RevolutPaymentStatusRefunded
	}
	payment.RefundStatus = revolutResp.State

	if err := s.db.WithContext(ctx).Save(&payment).Error; err != nil {
		return nil, fmt.Errorf("failed to update payment record: %w", err)
	}

	// Log refund event
	s.logPaymentEvent(ctx, payment.ID, "payment_refunded", "Payment refunded", map[string]interface{}{
		"refund_amount":     req.Amount,
		"refund_reason":     req.Reason,
		"revolut_refund_id": revolutResp.ID,
	})

	return &RefundResponse{
		RefundID:  revolutResp.ID,
		PaymentID: req.PaymentID,
		Amount:    req.Amount,
		Status:    revolutResp.State,
		CreatedAt: time.Now(),
		Reason:    req.Reason,
	}, nil
}

// CancelPayment cancels a pending payment
func (s *RevolutPaymentService) CancelPayment(ctx context.Context, paymentID string) error {
	// Get payment from database
	var payment models.Payment
	if err := s.db.WithContext(ctx).First(&payment, paymentID).Error; err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	if payment.Status != models.RevolutPaymentStatusPending {
		return fmt.Errorf("only pending payments can be cancelled")
	}

	// Update payment status
	payment.Status = models.RevolutPaymentStatusCancelled

	if err := s.db.WithContext(ctx).Save(&payment).Error; err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Log cancellation event
	s.logPaymentEvent(ctx, payment.ID, "payment_cancelled", "Payment cancelled", nil)

	return nil
}

// HandleWebhook processes webhook notifications from Revolut
// Headers expected:
// - Revolut-Signature: v1=signature (hex-encoded HMAC-SHA256)
// - Revolut-Request-Timestamp: UNIX timestamp of the webhook event
func (s *RevolutPaymentService) HandleWebhook(ctx context.Context, payload []byte, signature string, timestamp string) error {
	// Validate webhook signature
	if !s.validateWebhookSignature(payload, signature, timestamp) {
		return fmt.Errorf("invalid webhook signature")
	}

	// Parse webhook payload
	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Extract order ID from webhook
	orderID, ok := webhookData["order_id"].(string)
	if !ok {
		return fmt.Errorf("invalid webhook payload: missing order_id")
	}

	// Get payment by Revolut order ID
	var payment models.Payment
	if err := s.db.WithContext(ctx).Where("revolut_order_id = ?", orderID).First(&payment).Error; err != nil {
		return fmt.Errorf("payment not found for order ID %s: %w", orderID, err)
	}

	// Update payment based on webhook event
	if err := s.processWebhookEvent(ctx, &payment, webhookData); err != nil {
		return fmt.Errorf("failed to process webhook event: %w", err)
	}

	return nil
}

// GetPayment retrieves payment details by ID
func (s *RevolutPaymentService) GetPayment(ctx context.Context, paymentID string) (*models.Payment, error) {
	var payment models.Payment
	if err := s.db.WithContext(ctx).First(&payment, paymentID).Error; err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}
	return &payment, nil
}

// ListPayments retrieves a list of payments with optional filtering
func (s *RevolutPaymentService) ListPayments(ctx context.Context, orderID *uint, status *string, limit, offset int) ([]*models.Payment, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.Payment{})

	if orderID != nil {
		query = query.Where("order_id = ?", *orderID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payments: %w", err)
	}

	// Get payments with pagination
	var payments []*models.Payment
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get payments: %w", err)
	}

	return payments, total, nil
}

// Helper methods

// mapRevolutStatusToPaymentStatus maps Revolut order states to our payment status
// This is used when polling the Revolut API for order status
func (s *RevolutPaymentService) mapRevolutStatusToPaymentStatus(revolutState string) models.RevolutPaymentStatus {
	switch revolutState {
	case "PENDING":
		return models.RevolutPaymentStatusPending
	case "AUTHORIZED":
		return models.RevolutPaymentStatusAuthorized
	case "COMPLETED":
		return models.RevolutPaymentStatusCompleted
	case "FAILED":
		return models.RevolutPaymentStatusFailed
	case "CANCELLED":
		return models.RevolutPaymentStatusCancelled
	case "REFUNDED":
		return models.RevolutPaymentStatusRefunded
	default:
		return models.RevolutPaymentStatusPending
	}
}

// validateWebhookSignature validates the webhook signature according to Revolut's security requirements
// Based on: https://developer.revolut.com/docs/guides/accept-payments/tutorials/work-with-webhooks/verify-the-payload-signature
func (s *RevolutPaymentService) validateWebhookSignature(payload []byte, signature string, timestamp string) bool {
	if s.webhookSecret == "" {
		log.Printf("Warning: webhook secret not configured, skipping signature validation")
		return true
	}

	// Parse the signature format: v1=signature
	if len(signature) < 3 || signature[:2] != "v1" || signature[2] != '=' {
		log.Printf("Invalid signature format: %s", signature)
		return false
	}

	// Extract the actual signature (remove "v1=" prefix)
	actualSignature := signature[3:]

	// Step 1: Prepare the payload to sign
	// payload_to_sign = v1.{timestamp}.{raw-payload}
	payloadToSign := fmt.Sprintf("v1.%s.%s", timestamp, string(payload))
	log.Printf("[DEBUG] Payload to sign: %s", payloadToSign)

	// Step 2: Compute the expected signature using HMAC-SHA256
	// Use the webhook secret as the key and the prepared payload as the message
	h := hmac.New(sha256.New, []byte(s.webhookSecret))
	h.Write([]byte(payloadToSign))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Step 3: Compare signatures using constant-time comparison
	isValid := hmac.Equal([]byte(actualSignature), []byte(expectedSignature))

	if !isValid {
		log.Printf("Signature validation failed. Expected: %s, Received: %s", expectedSignature, actualSignature)
		log.Printf("[DEBUG] Webhook secret length: %d", len(s.webhookSecret))
		log.Printf("[DEBUG] Payload length: %d", len(payload))
		log.Printf("[DEBUG] Timestamp: %s", timestamp)
	} else {
		log.Printf("[DEBUG] Signature validation successful")
	}

	return isValid
}

// processWebhookEvent processes a webhook event and updates payment status
// Based on Revolut webhook documentation: https://developer.revolut.com/docs/guides/accept-payments/tutorials/work-with-webhooks/using-webhooks
func (s *RevolutPaymentService) processWebhookEvent(ctx context.Context, payment *models.Payment, webhookData map[string]interface{}) error {
	// Extract event type and order ID from webhook
	event, _ := webhookData["event"].(string)
	orderID, _ := webhookData["order_id"].(string)

	// The order_id from webhook is the same as the RevolutPaymentID we stored during order creation
	// No need to extract payment_id since it doesn't exist in webhook payload
	// The order_id in webhook corresponds to the id from order creation response

	// Log the webhook event first
	s.logPaymentEvent(ctx, payment.ID, "webhook_received", fmt.Sprintf("Webhook event: %s", event), map[string]interface{}{
		"webhook_event":      event,
		"revolut_order_id":   orderID,
		"revolut_payment_id": payment.RevolutPaymentID,
		"webhook_data":       webhookData,
	})

	// Handle different webhook events
	switch event {
	case "ORDER_COMPLETED":
		return s.handleOrderCompleted(ctx, payment, webhookData)
	case "ORDER_PAYMENT_FAILED":
		return s.handleOrderPaymentFailed(ctx, payment, webhookData)
	case "ORDER_AUTHORIZED":
		return s.handleOrderAuthorized(ctx, payment, webhookData)
	case "ORDER_CANCELLED":
		return s.handleOrderCancelled(ctx, payment, webhookData)
	default:
		// Log unknown event but don't fail
		log.Printf("Unknown webhook event: %s", event)
		return nil
	}
}

// handleOrderCompleted processes ORDER_COMPLETED webhook event
func (s *RevolutPaymentService) handleOrderCompleted(ctx context.Context, payment *models.Payment, webhookData map[string]interface{}) error {
	oldStatus := payment.Status
	payment.Status = models.RevolutPaymentStatusCompleted
	now := time.Now()
	payment.CompletedAt = &now

	// The RevolutPaymentID is already set during order creation
	// No need to extract from webhook since payment_id doesn't exist in webhook payload
	// The order_id in webhook corresponds to the RevolutPaymentID we already have

	// Update order status to PAID
	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, payment.OrderID).Error; err != nil {
		log.Printf("Warning: failed to get order for payment %d: %v", payment.ID, err)
	} else {
		order.PaymentStatus = models.PaymentStatusPaid
		order.PaymentDate = &now
		if err := s.db.WithContext(ctx).Save(&order).Error; err != nil {
			log.Printf("Warning: failed to update order payment status: %v", err)
		}
	}

	// Save payment changes
	if err := s.db.WithContext(ctx).Save(payment).Error; err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// Log the status change
	s.logPaymentEvent(ctx, payment.ID, "payment_completed", "Payment completed successfully", map[string]interface{}{
		"old_status":         oldStatus,
		"new_status":         payment.Status,
		"completed_at":       now,
		"revolut_payment_id": payment.RevolutPaymentID,
	})

	return nil
}

// handleOrderPaymentFailed processes ORDER_PAYMENT_FAILED webhook event
func (s *RevolutPaymentService) handleOrderPaymentFailed(ctx context.Context, payment *models.Payment, webhookData map[string]interface{}) error {
	oldStatus := payment.Status
	payment.Status = models.RevolutPaymentStatusFailed

	// Extract failure reason if available
	if failureReason, ok := webhookData["failure_reason"].(string); ok {
		payment.FailureReason = failureReason
	}

	// Update order status to FAILED
	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, payment.OrderID).Error; err != nil {
		log.Printf("Warning: failed to get order for payment %d: %v", payment.ID, err)
	} else {
		order.PaymentStatus = models.PaymentStatusFailed
		if err := s.db.WithContext(ctx).Save(&order).Error; err != nil {
			log.Printf("Warning: failed to update order payment status: %v", err)
		}
	}

	// Save payment changes
	if err := s.db.WithContext(ctx).Save(payment).Error; err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// Log the status change
	s.logPaymentEvent(ctx, payment.ID, "payment_failed", "Payment failed", map[string]interface{}{
		"old_status":     oldStatus,
		"new_status":     payment.Status,
		"failure_reason": payment.FailureReason,
	})

	return nil
}

// handleOrderAuthorized processes ORDER_AUTHORIZED webhook event
func (s *RevolutPaymentService) handleOrderAuthorized(ctx context.Context, payment *models.Payment, webhookData map[string]interface{}) error {
	oldStatus := payment.Status
	payment.Status = models.RevolutPaymentStatusAuthorized

	// Save payment changes
	if err := s.db.WithContext(ctx).Save(payment).Error; err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// Log the status change
	s.logPaymentEvent(ctx, payment.ID, "payment_authorized", "Payment authorized", map[string]interface{}{
		"old_status": oldStatus,
		"new_status": payment.Status,
	})

	return nil
}

// handleOrderCancelled processes ORDER_CANCELLED webhook event
func (s *RevolutPaymentService) handleOrderCancelled(ctx context.Context, payment *models.Payment, webhookData map[string]interface{}) error {
	oldStatus := payment.Status
	payment.Status = models.RevolutPaymentStatusCancelled

	// Update order status to CANCELLED
	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, payment.OrderID).Error; err != nil {
		log.Printf("Warning: failed to get order for payment %d: %v", payment.ID, err)
	} else {
		order.Status = models.OrderStatusCancelled
		if err := s.db.WithContext(ctx).Save(&order).Error; err != nil {
			log.Printf("Warning: failed to update order status: %v", err)
		}
	}

	// Save payment changes
	if err := s.db.WithContext(ctx).Save(payment).Error; err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// Log the status change
	s.logPaymentEvent(ctx, payment.ID, "payment_cancelled", "Payment cancelled", map[string]interface{}{
		"old_status": oldStatus,
		"new_status": payment.Status,
	})

	return nil
}

// logPaymentEvent logs a payment event
func (s *RevolutPaymentService) logPaymentEvent(ctx context.Context, paymentID uint, event, message string, metadata map[string]interface{}) {
	paymentLog := &models.PaymentLog{
		PaymentID: paymentID,
		Event:     event,
		Message:   message,
		Metadata:  models.JSON(metadata),
		CreatedBy: 0, // System event
	}

	if err := s.db.WithContext(ctx).Create(paymentLog).Error; err != nil {
		log.Printf("Warning: failed to log payment event: %v", err)
	}
}

// UpdateRevolutPaymentID manually updates the RevolutPaymentID for a payment
// This can be used when the payment ID is obtained from other sources
func (s *RevolutPaymentService) UpdateRevolutPaymentID(ctx context.Context, paymentID string, revolutPaymentID string) error {
	var payment models.Payment
	if err := s.db.WithContext(ctx).First(&payment, paymentID).Error; err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	oldPaymentID := payment.RevolutPaymentID
	payment.RevolutPaymentID = revolutPaymentID

	if err := s.db.WithContext(ctx).Save(&payment).Error; err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// Log the update
	s.logPaymentEvent(ctx, payment.ID, "payment_id_updated", "RevolutPaymentID updated", map[string]interface{}{
		"old_revolut_payment_id": oldPaymentID,
		"new_revolut_payment_id": revolutPaymentID,
	})

	return nil
}
