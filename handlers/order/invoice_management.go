package order

import (
	"fmt"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateInvoiceRequest struct {
	OrderID          uint   `json:"order_id" binding:"required"`
	DueDate          string `json:"due_date"`
	Notes            string `json:"notes"`
	PaymentMethod    string `json:"payment_method"`
	PaymentReference string `json:"payment_reference"`
}

type UpdateInvoiceRequest struct {
	Status           string `json:"status"`
	PaymentMethod    string `json:"payment_method"`
	PaymentReference string `json:"payment_reference"`
	Notes            string `json:"notes"`
}

// CreateInvoice - Admin endpoint to create invoice for an order
func (h *OrderHandler) CreateInvoice(c *gin.Context) {
	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "invoice/create", err.Error())
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the order
	var order models.Order
	if err := tx.First(&order, req.OrderID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			response.GenerateNotFoundResponse(c, "invoice/create", "Order not found")
		} else {
			response.GenerateInternalServerErrorResponse(c, "invoice/create", "Failed to get order")
		}
		return
	}

	// Check if invoice already exists for this order
	var existingInvoice models.Invoice
	if err := tx.Where("order_id = ?", req.OrderID).First(&existingInvoice).Error; err == nil {
		tx.Rollback()
		response.GenerateBadRequestResponse(c, "invoice/create", "Invoice already exists for this order")
		return
	}

	// Parse due date
	var dueDate time.Time
	var err error
	if req.DueDate != "" {
		dueDate, err = time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			tx.Rollback()
			response.GenerateBadRequestResponse(c, "invoice/create", "Invalid due date format. Use YYYY-MM-DD")
			return
		}
	} else {
		// Default due date is 30 days from now
		dueDate = time.Now().AddDate(0, 0, 30)
	}

	// Generate invoice number
	invoiceNumber := generateInvoiceNumber()

	// Create invoice
	invoice := models.Invoice{
		OrderID:          req.OrderID,
		InvoiceNumber:    invoiceNumber,
		IssueDate:        time.Now(),
		DueDate:          dueDate,
		Amount:           order.FinalAmount,
		TaxAmount:        order.TaxAmount,
		Status:           "pending",
		PaymentMethod:    req.PaymentMethod,
		PaymentReference: req.PaymentReference,
		Notes:            req.Notes,
	}

	if err := tx.Create(&invoice).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "invoice/create", "Failed to create invoice")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "invoice/create", "Failed to commit transaction")
		return
	}

	// Load complete invoice with order details
	var completeInvoice models.Invoice
	if err := h.db.Preload("Order.User").
		Preload("Order.Items.ProductVariant.Product").
		First(&completeInvoice, invoice.ID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "invoice/create", "Invoice created but failed to load details")
		return
	}

	response.GenerateCreatedResponse(c, "Invoice created successfully", completeInvoice)
}

// GetInvoices - Admin endpoint to get all invoices with filtering
func (h *OrderHandler) GetInvoices(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")
	status := c.Query("status")
	orderID := c.Query("order_id")

	// Build query
	query := h.db.Model(&models.Invoice{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if orderID != "" {
		query = query.Where("order_id = ?", orderID)
	}

	// Get total count
	var totalCount int64
	query.Count(&totalCount)

	// Parse pagination
	pageInt := 1
	limitInt := 20
	fmt.Sscanf(page, "%d", &pageInt)
	fmt.Sscanf(limit, "%d", &limitInt)

	if pageInt < 1 {
		pageInt = 1
	}
	if limitInt < 1 || limitInt > 100 {
		limitInt = 20
	}

	offset := (pageInt - 1) * limitInt

	// Get invoices
	var invoices []models.Invoice
	if err := query.
		Preload("Order.User").
		Order("created_at DESC").
		Limit(limitInt).
		Offset(offset).
		Find(&invoices).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "invoice/get_all", "Failed to get invoices")
		return
	}

	responseData := map[string]interface{}{
		"invoices":    invoices,
		"page":        pageInt,
		"limit":       limitInt,
		"total_count": totalCount,
		"total_pages": (totalCount + int64(limitInt) - 1) / int64(limitInt),
	}

	response.GenerateSuccessResponse(c, "Invoices retrieved successfully", responseData)
}

// GetInvoice - Admin endpoint to get single invoice
func (h *OrderHandler) GetInvoice(c *gin.Context) {
	invoiceID := c.Param("id")

	var invoice models.Invoice
	if err := h.db.
		Preload("Order.User").
		Preload("Order.ShippingAddress").
		Preload("Order.Items.ProductVariant.Product").
		First(&invoice, invoiceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateNotFoundResponse(c, "invoice/get", "Invoice not found")
		} else {
			response.GenerateInternalServerErrorResponse(c, "invoice/get", "Failed to get invoice")
		}
		return
	}

	response.GenerateSuccessResponse(c, "Invoice retrieved successfully", invoice)
}

// UpdateInvoice - Admin endpoint to update invoice
func (h *OrderHandler) UpdateInvoice(c *gin.Context) {
	invoiceID := c.Param("id")

	var req UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "invoice/update", err.Error())
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the invoice
	var invoice models.Invoice
	if err := tx.First(&invoice, invoiceID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			response.GenerateNotFoundResponse(c, "invoice/update", "Invoice not found")
		} else {
			response.GenerateInternalServerErrorResponse(c, "invoice/update", "Failed to get invoice")
		}
		return
	}

	// Update fields
	if req.Status != "" {
		validStatuses := []string{"pending", "paid", "overdue", "cancelled"}
		isValid := false
		for _, status := range validStatuses {
			if status == req.Status {
				isValid = true
				break
			}
		}
		if !isValid {
			tx.Rollback()
			response.GenerateBadRequestResponse(c, "invoice/update", "Invalid status")
			return
		}
		invoice.Status = req.Status

		// Set payment date if status is paid
		if req.Status == "paid" && invoice.PaymentDate == nil {
			now := time.Now()
			invoice.PaymentDate = &now
		}
	}

	if req.PaymentMethod != "" {
		invoice.PaymentMethod = req.PaymentMethod
	}
	if req.PaymentReference != "" {
		invoice.PaymentReference = req.PaymentReference
	}
	if req.Notes != "" {
		invoice.Notes = req.Notes
	}

	if err := tx.Save(&invoice).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "invoice/update", "Failed to update invoice")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "invoice/update", "Failed to commit transaction")
		return
	}

	// Load complete invoice
	var completeInvoice models.Invoice
	if err := h.db.Preload("Order.User").
		First(&completeInvoice, invoice.ID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "invoice/update", "Invoice updated but failed to load details")
		return
	}

	response.GenerateSuccessResponse(c, "Invoice updated successfully", completeInvoice)
}

// generateInvoiceNumber generates a unique invoice number
func generateInvoiceNumber() string {
	now := time.Now()
	return fmt.Sprintf("INV-%d%02d%02d-%d",
		now.Year(), now.Month(), now.Day(),
		now.Unix()%10000)
}
