package support

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

// applyDisputeFilters applies filters/sort/pagination on disputes
func (h *SupportHandler) applyDisputeFilters(c *gin.Context, query *gorm.DB) (*gorm.DB, int, int) {
	if status := c.Query("status"); status != "" {
		query = query.Where("status IN ?", strings.Split(strings.ToUpper(status), ","))
	}
	if category := c.Query("category"); category != "" {
		query = query.Where("category IN ?", strings.Split(strings.ToUpper(category), ","))
	}
	if priority := c.Query("priority"); priority != "" {
		query = query.Where("priority IN ?", strings.Split(strings.ToUpper(priority), ","))
	}
	if min := c.Query("amount_min"); min != "" {
		if v, err := strconv.ParseFloat(min, 64); err == nil {
			query = query.Where("amount >= ?", v)
		}
	}
	if max := c.Query("amount_max"); max != "" {
		if v, err := strconv.ParseFloat(max, 64); err == nil {
			query = query.Where("amount <= ?", v)
		}
	}
	if start := c.Query("start_date"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			query = query.Where("created_at >= ?", t)
		} else if t2, err2 := time.Parse("2006-01-02", start); err2 == nil {
			query = query.Where("created_at >= ?", t2)
		}
	}
	if end := c.Query("end_date"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			query = query.Where("created_at <= ?", t)
		} else if t2, err2 := time.Parse("2006-01-02", end); err2 == nil {
			query = query.Where("created_at < ?", t2.Add(24*time.Hour))
		}
	}
	// sort
	sort := strings.ToLower(c.DefaultQuery("sort_by", "date"))
	order := strings.ToUpper(c.DefaultQuery("sort_order", "DESC"))
	if order != "ASC" {
		order = "DESC"
	}
	switch sort {
	case "date":
		query = query.Order("created_at " + order)
	case "priority":
		caseExpr := fmt.Sprintf("CASE priority WHEN 'LOW' THEN 1 WHEN 'MEDIUM' THEN 2 WHEN 'HIGH' THEN 3 WHEN 'URGENT' THEN 4 END %s", order)
		query = query.Order(caseExpr).Order("created_at DESC")
	case "amount":
		query = query.Order("amount " + order).Order("created_at DESC")
	case "status":
		query = query.Order("status " + order).Order("created_at DESC")
	default:
		query = query.Order("created_at DESC")
	}
	// pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	return query, page, pageSize
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
		if userType != models.Admin {
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
	q := h.db.Where("user_id = ?", userID).Model(&models.Dispute{})
	q, _, _ = h.applyDisputeFilters(c, q)
	if err := q.Preload("User").Preload("Order").Preload("Payment").Order("created_at DESC").Find(&disputes).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/get-user-disputes", err.Error())
		return
	}
	response.GenerateSuccessResponse(c, "User disputes retrieved successfully", disputes)
}

// GetAllDisputes retrieves all disputes (admin only)
func (h *SupportHandler) GetAllDisputes(c *gin.Context) {
	userType, exists := c.Get("user_type")
	if !exists || userType != models.Admin {
		response.GenerateForbiddenResponse(c, "support/get-all-disputes", "Admin access required")
		return
	}
	var disputes []models.Dispute
	q := h.db.Model(&models.Dispute{})
	q, _, _ = h.applyDisputeFilters(c, q)
	if err := q.Preload("User").Preload("Order").Preload("Payment").Find(&disputes).Error; err != nil {
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
	if dispute.UserID != userID.(uint) && userType != models.Admin {
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

	// send status update email if status changed
	if _, ok := updates["status"]; ok && h.emailTriggerSvc != nil {
		var user models.User
		if err := h.db.First(&user, dispute.UserID).Error; err == nil {
			data := map[string]interface{}{
				"UserName":        strings.TrimSpace(user.FirstName + " " + user.LastName),
				"DisputeID":       dispute.ID,
				"DisputeTitle":    dispute.Title,
				"OldStatus":       "",
				"NewStatus":       updates["status"],
				"UserMessageHTML": template.HTML(dispute.Description),
				"AdminNoteHTML":   template.HTML(fmt.Sprintf("%v", updates["internal_notes"])),
				"subject":         fmt.Sprintf("Your dispute #%d status updated", dispute.ID),
			}
			_ = h.emailTriggerSvc.TriggerDisputeStatusUpdated(user.Email, data["UserName"].(string), data)
		}
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
	isAdmin := userType == models.Admin

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

	// Email notify dispute owner on non-internal responses
	if !request.IsInternal && h.emailTriggerSvc != nil {
		var owner models.User
		if err := h.db.First(&owner, dispute.UserID).Error; err == nil {
			responderName := "User"
			var responder models.User
			if err := h.db.First(&responder, userID.(uint)).Error; err == nil {
				responderName = strings.TrimSpace(responder.FirstName + " " + responder.LastName)
			}
			data := map[string]interface{}{
				"UserName":        strings.TrimSpace(owner.FirstName + " " + owner.LastName),
				"DisputeID":       dispute.ID,
				"DisputeTitle":    dispute.Title,
				"UserMessageHTML": template.HTML(dispute.Description),
				"ResponderName":   responderName,
				"RespondedAt":     time.Now().Format("2006-01-02 15:04:05"),
				"ResponseHTML":    template.HTML(request.Message),
				"subject":         fmt.Sprintf("New response on your dispute #%d", dispute.ID),
			}
			_ = h.emailTriggerSvc.TriggerDisputeResponse(owner.Email, data["UserName"].(string), data)
		}
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
	if !exists || userType != models.Admin {
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
