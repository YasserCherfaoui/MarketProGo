package support

import (
	"strconv"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

// CreateDisputeRequest represents the request to create a dispute
type CreateDisputeRequest struct {
	OrderID     *uint                      `json:"order_id,omitempty"`
	PaymentID   *uint                      `json:"payment_id,omitempty"`
	Title       string                     `json:"title" binding:"required"`
	Description string                     `json:"description" binding:"required"`
	Category    models.DisputeCategory     `json:"category" binding:"required"`
	Priority    models.DisputePriority     `json:"priority"`
	Amount      *float64                   `json:"amount,omitempty"`
	Currency    string                     `json:"currency"`
	Attachments []DisputeAttachmentRequest `json:"attachments,omitempty"`
}

// DisputeAttachmentRequest represents an attachment for a dispute
type DisputeAttachmentRequest struct {
	FileName string `json:"file_name" binding:"required"`
	FileURL  string `json:"file_url" binding:"required"`
	FileSize int64  `json:"file_size"`
	FileType string `json:"file_type"`
}

// UpdateDisputeRequest represents the request to update a dispute
type UpdateDisputeRequest struct {
	Title         string                 `json:"title,omitempty"`
	Description   string                 `json:"description,omitempty"`
	Category      models.DisputeCategory `json:"category,omitempty"`
	Status        models.DisputeStatus   `json:"status,omitempty"`
	Priority      models.DisputePriority `json:"priority,omitempty"`
	Resolution    string                 `json:"resolution,omitempty"`
	InternalNotes string                 `json:"internal_notes,omitempty"`
}

// DisputeResponseRequest represents a response to a dispute
type DisputeResponseRequest struct {
	Message    string `json:"message" binding:"required"`
	IsInternal bool   `json:"is_internal"`
}

// CreateDispute creates a new dispute
func (h *SupportHandler) CreateDispute(c *gin.Context) {
	var request CreateDisputeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.GenerateBadRequestResponse(c, "support/create-dispute", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "support/create-dispute", "User not authenticated")
		return
	}

	// Create the dispute
	dispute := models.Dispute{
		UserID:      userID.(uint),
		OrderID:     request.OrderID,
		PaymentID:   request.PaymentID,
		Title:       request.Title,
		Description: request.Description,
		Category:    request.Category,
		Priority:    request.Priority,
		Status:      models.DisputeStatusOpen,
		Amount:      request.Amount,
		Currency:    request.Currency,
	}

	if err := h.db.Create(&dispute).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/create-dispute", err.Error())
		return
	}

	// Handle attachments if provided
	if len(request.Attachments) > 0 {
		for _, attachment := range request.Attachments {
			disputeAttachment := models.DisputeAttachment{
				DisputeID: dispute.ID,
				FileName:  attachment.FileName,
				FileURL:   attachment.FileURL,
				FileSize:  attachment.FileSize,
				FileType:  attachment.FileType,
			}
			if err := h.db.Create(&disputeAttachment).Error; err != nil {
				response.GenerateInternalServerErrorResponse(c, "support/create-dispute", "Failed to create attachment")
				return
			}
		}
	}

	// Load the created dispute with relationships
	if err := h.db.Preload("User").Preload("Order").Preload("Payment").Preload("Attachments").First(&dispute, dispute.ID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/create-dispute", "Failed to load created dispute")
		return
	}

	response.GenerateSuccessResponse(c, "Dispute created successfully", dispute)
}

// GetDispute retrieves a specific dispute
func (h *SupportHandler) GetDispute(c *gin.Context) {
	disputeID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "support/get-dispute", "Invalid dispute ID")
		return
	}

	var dispute models.Dispute
	if err := h.db.Preload("User").Preload("Order").Preload("Payment").Preload("Attachments").Preload("Responses.User").First(&dispute, disputeID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "support/get-dispute", "Dispute not found")
		return
	}

	// Check if user has permission to view this dispute
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "support/get-dispute", "User not authenticated")
		return
	}

	// Only allow users to view their own disputes or admins to view any dispute
	if dispute.UserID != userID.(uint) {
		userType, _ := c.Get("user_type")
		if userType != "ADMIN" {
			response.GenerateForbiddenResponse(c, "support/get-dispute", "Access denied")
			return
		}
	}

	response.GenerateSuccessResponse(c, "Dispute retrieved successfully", dispute)
}

// GetUserDisputes retrieves all disputes for the current user
func (h *SupportHandler) GetUserDisputes(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "support/get-user-disputes", "User not authenticated")
		return
	}

	var disputes []models.Dispute
	if err := h.db.Where("user_id = ?", userID).Preload("User").Preload("Order").Preload("Payment").Order("created_at DESC").Find(&disputes).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/get-user-disputes", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "User disputes retrieved successfully", disputes)
}

// GetAllDisputes retrieves all disputes (admin only)
func (h *SupportHandler) GetAllDisputes(c *gin.Context) {
	userType, exists := c.Get("user_type")
	if !exists || userType != "ADMIN" {
		response.GenerateForbiddenResponse(c, "support/get-all-disputes", "Admin access required")
		return
	}

	var disputes []models.Dispute
	if err := h.db.Preload("User").Preload("Order").Preload("Payment").Order("created_at DESC").Find(&disputes).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/get-all-disputes", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "All disputes retrieved successfully", disputes)
}

// UpdateDispute updates a dispute
func (h *SupportHandler) UpdateDispute(c *gin.Context) {
	disputeID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "support/update-dispute", "Invalid dispute ID")
		return
	}

	var request UpdateDisputeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.GenerateBadRequestResponse(c, "support/update-dispute", err.Error())
		return
	}

	var dispute models.Dispute
	if err := h.db.First(&dispute, disputeID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "support/update-dispute", "Dispute not found")
		return
	}

	// Check permissions
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "support/update-dispute", "User not authenticated")
		return
	}

	userType, _ := c.Get("user_type")
	if dispute.UserID != userID.(uint) && userType != "ADMIN" {
		response.GenerateForbiddenResponse(c, "support/update-dispute", "Access denied")
		return
	}

	// Update fields
	updates := make(map[string]interface{})
	if request.Title != "" {
		updates["title"] = request.Title
	}
	if request.Description != "" {
		updates["description"] = request.Description
	}
	if request.Category != "" {
		updates["category"] = request.Category
	}
	if request.Priority != "" {
		updates["priority"] = request.Priority
	}
	if request.Status != "" {
		updates["status"] = request.Status
		if request.Status == models.DisputeStatusResolved || request.Status == models.DisputeStatusClosed {
			now := time.Now()
			updates["resolved_at"] = &now
			updates["resolved_by"] = userID
		}
	}
	if request.Resolution != "" {
		updates["resolution"] = request.Resolution
	}
	if request.InternalNotes != "" {
		updates["internal_notes"] = request.InternalNotes
	}

	if err := h.db.Model(&dispute).Updates(updates).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/update-dispute", err.Error())
		return
	}

	// Load updated dispute
	if err := h.db.Preload("User").Preload("Order").Preload("Payment").Preload("Attachments").Preload("Responses.User").First(&dispute, disputeID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/update-dispute", "Failed to load updated dispute")
		return
	}

	response.GenerateSuccessResponse(c, "Dispute updated successfully", dispute)
}

// AddDisputeResponse adds a response to a dispute
func (h *SupportHandler) AddDisputeResponse(c *gin.Context) {
	disputeID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "support/add-dispute-response", "Invalid dispute ID")
		return
	}

	var request DisputeResponseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.GenerateBadRequestResponse(c, "support/add-dispute-response", err.Error())
		return
	}

	var dispute models.Dispute
	if err := h.db.First(&dispute, disputeID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "support/add-dispute-response", "Dispute not found")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "support/add-dispute-response", "User not authenticated")
		return
	}

	userType, _ := c.Get("user_type")
	isAdmin := userType == "ADMIN"

	// Check permissions
	if dispute.UserID != userID.(uint) && !isAdmin {
		response.GenerateForbiddenResponse(c, "support/add-dispute-response", "Access denied")
		return
	}

	// Create response
	disputeResponse := models.DisputeResponse{
		DisputeID:   uint(disputeID),
		UserID:      userID.(uint),
		Message:     request.Message,
		IsInternal:  request.IsInternal,
		IsFromAdmin: isAdmin,
	}

	if err := h.db.Create(&disputeResponse).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/add-dispute-response", err.Error())
		return
	}

	// Update dispute status if admin responded
	if isAdmin && dispute.Status == models.DisputeStatusOpen {
		h.db.Model(&dispute).Update("status", models.DisputeStatusInProgress)
	}

	response.GenerateSuccessResponse(c, "Response added successfully", disputeResponse)
}

// DeleteDispute deletes a dispute (admin only)
func (h *SupportHandler) DeleteDispute(c *gin.Context) {
	disputeID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "support/delete-dispute", "Invalid dispute ID")
		return
	}

	userType, exists := c.Get("user_type")
	if !exists || userType != "ADMIN" {
		response.GenerateForbiddenResponse(c, "support/delete-dispute", "Admin access required")
		return
	}

	var dispute models.Dispute
	if err := h.db.First(&dispute, disputeID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "support/delete-dispute", "Dispute not found")
		return
	}

	if err := h.db.Delete(&dispute).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/delete-dispute", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Dispute deleted successfully", nil)
}
