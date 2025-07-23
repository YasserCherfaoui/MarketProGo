package payment

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/payment"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PaymentHandler handles payment-related HTTP requests
type PaymentHandler struct {
	paymentService payment.PaymentService
	db             *gorm.DB
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(db *gorm.DB, paymentService payment.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		db:             db,
	}
}

// CreatePaymentRequest represents the request body for creating a payment
type CreatePaymentRequest struct {
	OrderID     uint              `json:"order_id" binding:"required"`
	Amount      float64           `json:"amount" binding:"required,gt=0"`
	Currency    string            `json:"currency" binding:"required"`
	Description string            `json:"description" binding:"required"`
	ReturnURL   string            `json:"return_url"`
	CancelURL   string            `json:"cancel_url"`
	Metadata    map[string]string `json:"metadata"`
}

// RefundPaymentRequest represents the request body for refunding a payment
type RefundPaymentRequest struct {
	Amount   float64           `json:"amount" binding:"required,gt=0"`
	Reason   string            `json:"reason"`
	Metadata map[string]string `json:"metadata"`
}

// InitiatePayment handles POST /api/v1/payments
func (h *PaymentHandler) InitiatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// Get user from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Get user details
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.GenerateErrorResponse(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}

	// Verify order belongs to user
	var order models.Order
	if err := h.db.Where("id = ? AND user_id = ?", req.OrderID, userID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateErrorResponse(c, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found or does not belong to user")
			return
		}
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to verify order")
		return
	}

	// Create customer info
	customerInfo := &payment.CustomerInfo{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.FirstName + " " + user.LastName,
		Phone: user.Phone,
	}

	// Create payment request
	paymentReq := &payment.PaymentRequest{
		OrderID:      req.OrderID,
		Amount:       req.Amount,
		Currency:     req.Currency,
		Description:  req.Description,
		CustomerInfo: customerInfo,
		ReturnURL:    req.ReturnURL,
		CancelURL:    req.CancelURL,
		Metadata:     req.Metadata,
	}

	// Create payment
	paymentResp, err := h.paymentService.CreatePayment(c.Request.Context(), paymentReq)
	if err != nil {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "PAYMENT_CREATION_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    paymentResp,
	})
}

// GetPaymentStatus handles GET /api/v1/payments/:id/status
func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_PAYMENT_ID", "Payment ID is required")
		return
	}

	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Verify payment belongs to user
	var payment models.Payment
	if err := h.db.Joins("JOIN orders ON payments.order_id = orders.id").
		Where("payments.id = ? AND orders.user_id = ?", paymentID, userID).
		First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateErrorResponse(c, http.StatusNotFound, "PAYMENT_NOT_FOUND", "Payment not found or does not belong to user")
			return
		}
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to get payment")
		return
	}

	// Get payment status
	status, err := h.paymentService.GetPaymentStatus(c.Request.Context(), paymentID)
	if err != nil {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "STATUS_RETRIEVAL_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"payment_id": paymentID,
			"status":     status,
		},
	})
}

// GetPayment handles GET /api/v1/payments/:id
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_PAYMENT_ID", "Payment ID is required")
		return
	}

	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Get payment with order details
	var payment models.Payment
	if err := h.db.Joins("JOIN orders ON payments.order_id = orders.id").
		Where("payments.id = ? AND orders.user_id = ?", paymentID, userID).
		Preload("Order").
		First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateErrorResponse(c, http.StatusNotFound, "PAYMENT_NOT_FOUND", "Payment not found or does not belong to user")
			return
		}
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to get payment")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    payment,
	})
}

// ListPayments handles GET /api/v1/payments
func (h *PaymentHandler) ListPayments(c *gin.Context) {
	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	orderIDStr := c.Query("order_id")

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Build query for user's payments
	query := h.db.Joins("JOIN orders ON payments.order_id = orders.id").
		Where("orders.user_id = ?", userID)

	// Apply filters
	if status != "" {
		query = query.Where("payments.status = ?", status)
	}
	if orderIDStr != "" {
		if orderID, err := strconv.ParseUint(orderIDStr, 10, 32); err == nil {
			query = query.Where("payments.order_id = ?", orderID)
		}
	}

	// Get total count
	var total int64
	if err := query.Model(&models.Payment{}).Count(&total).Error; err != nil {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "COUNT_ERROR", "Failed to count payments")
		return
	}

	// Get payments with pagination
	var payments []models.Payment
	if err := query.Preload("Order").
		Offset(offset).
		Limit(limit).
		Order("payments.created_at DESC").
		Find(&payments).Error; err != nil {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "RETRIEVAL_ERROR", "Failed to retrieve payments")
		return
	}

	// Calculate pagination info
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"payments": payments,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
				"has_next":    hasNext,
				"has_prev":    hasPrev,
			},
		},
	})
}

// RefundPayment handles POST /api/v1/payments/:id/refund (Admin only)
func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_PAYMENT_ID", "Payment ID is required")
		return
	}

	// Check if user is admin
	userType, exists := c.Get("user_type")
	if !exists || userType != models.Admin {
		response.GenerateErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Admin access required")
		return
	}

	var req RefundPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// Create refund request
	refundReq := &payment.RefundRequest{
		PaymentID: paymentID,
		Amount:    req.Amount,
		Reason:    req.Reason,
		Metadata:  req.Metadata,
	}

	// Process refund
	refundResp, err := h.paymentService.RefundPayment(c.Request.Context(), refundReq)
	if err != nil {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "REFUND_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    refundResp,
	})
}

// CancelPayment handles POST /api/v1/payments/:id/cancel
func (h *PaymentHandler) CancelPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_PAYMENT_ID", "Payment ID is required")
		return
	}

	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Verify payment belongs to user
	var payment models.Payment
	if err := h.db.Joins("JOIN orders ON payments.order_id = orders.id").
		Where("payments.id = ? AND orders.user_id = ?", paymentID, userID).
		First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateErrorResponse(c, http.StatusNotFound, "PAYMENT_NOT_FOUND", "Payment not found or does not belong to user")
			return
		}
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to get payment")
		return
	}

	// Cancel payment
	if err := h.paymentService.CancelPayment(c.Request.Context(), paymentID); err != nil {
		response.GenerateErrorResponse(c, http.StatusInternalServerError, "CANCELLATION_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Payment cancelled successfully",
	})
}

// HandleWebhook handles POST /api/v1/payments/webhook
func (h *PaymentHandler) HandleWebhook(c *gin.Context) {
	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[DEBUG] Failed to read webhook request body: %v", err)
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}
	log.Printf("[DEBUG] Webhook request body read successfully: %d bytes", len(body))

	// Get webhook signature from header
	signature := c.GetHeader("Revolut-Signature")
	if signature == "" {
		log.Printf("[DEBUG] Missing Revolut-Signature header in webhook request")
		response.GenerateErrorResponse(c, http.StatusBadRequest, "MISSING_SIGNATURE", "Webhook signature is required")
		return
	}
	log.Printf("[DEBUG] Revolut-Signature header received: %s", signature)

	// Get webhook timestamp from header for additional security
	timestamp := c.GetHeader("Revolut-Request-Timestamp")
	if timestamp == "" {
		log.Printf("[DEBUG] Missing Revolut-Request-Timestamp header in webhook request")
		response.GenerateErrorResponse(c, http.StatusBadRequest, "MISSING_TIMESTAMP", "Webhook timestamp is required")
		return
	}
	log.Printf("[DEBUG] Revolut-Request-Timestamp header received: %s", timestamp)

	// Validate timestamp (optional but recommended for security)
	if err := h.validateWebhookTimestamp(timestamp); err != nil {
		log.Printf("[DEBUG] Webhook timestamp validation failed: %v", err)
		response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_TIMESTAMP", err.Error())
		return
	}

	// Process webhook
	if err := h.paymentService.HandleWebhook(c.Request.Context(), body, signature, timestamp); err != nil {
		log.Printf("[DEBUG] Error processing webhook: %v", err)
		response.GenerateErrorResponse(c, http.StatusBadRequest, "WEBHOOK_PROCESSING_FAILED", err.Error())
		return
	}

	log.Printf("[DEBUG] Webhook processed successfully")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Webhook processed successfully",
	})
}

// validateWebhookTimestamp validates the webhook timestamp to prevent replay attacks
// Revolut sends timestamps in milliseconds since epoch
func (h *PaymentHandler) validateWebhookTimestamp(timestamp string) error {
	// Parse timestamp (milliseconds since epoch)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp format: %w", err)
	}

	// Convert to seconds
	webhookTime := time.Unix(ts/1000, (ts%1000)*1000000)
	now := time.Now()

	// Allow webhook to be up to 5 minutes old (to account for network delays)
	maxAge := 5 * time.Minute
	if now.Sub(webhookTime) > maxAge {
		return fmt.Errorf("webhook timestamp is too old: %v (max age: %v)", webhookTime, maxAge)
	}

	// Allow webhook to be up to 1 minute in the future (to account for clock skew)
	maxFuture := 1 * time.Minute
	if webhookTime.Sub(now) > maxFuture {
		return fmt.Errorf("webhook timestamp is too far in the future: %v (max future: %v)", webhookTime, maxFuture)
	}

	return nil
}
