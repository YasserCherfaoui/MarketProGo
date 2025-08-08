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

// CreateTicketRequest represents the request to create a support ticket
type CreateTicketRequest struct {
	Title       string                    `json:"title" binding:"required"`
	Description string                    `json:"description" binding:"required"`
	Category    models.TicketCategory     `json:"category" binding:"required"`
	Priority    models.TicketPriority     `json:"priority"`
	OrderID     *uint                     `json:"order_id,omitempty"`
	Attachments []TicketAttachmentRequest `json:"attachments,omitempty"`
}

// TicketAttachmentRequest represents an attachment for a ticket
type TicketAttachmentRequest struct {
	FileName string `json:"file_name" binding:"required"`
	FileURL  string `json:"file_url" binding:"required"`
	FileSize int64  `json:"file_size"`
	FileType string `json:"file_type"`
}

// UpdateTicketRequest represents the request to update a support ticket
type UpdateTicketRequest struct {
	Title         string                `json:"title,omitempty"`
	Description   string                `json:"description,omitempty"`
	Category      models.TicketCategory `json:"category,omitempty"`
	Priority      models.TicketPriority `json:"priority,omitempty"`
	Status        models.TicketStatus   `json:"status,omitempty"`
	Resolution    string                `json:"resolution,omitempty"`
	InternalNotes string                `json:"internal_notes,omitempty"`
}

// TicketResponseRequest represents a response to a ticket
type TicketResponseRequest struct {
	Message    string `json:"message" binding:"required"`
	IsInternal bool   `json:"is_internal"`
}

// applyTicketFilters applies filters/sort/pagination for tickets
func (h *SupportHandler) applyTicketFilters(c *gin.Context, query *gorm.DB) (*gorm.DB, int, int) {
	if status := c.Query("status"); status != "" {
		query = query.Where("status IN ?", strings.Split(strings.ToUpper(status), ","))
	}
	if category := c.Query("category"); category != "" {
		query = query.Where("category IN ?", strings.Split(strings.ToUpper(category), ","))
	}
	if priority := c.Query("priority"); priority != "" {
		query = query.Where("priority IN ?", strings.Split(strings.ToUpper(priority), ","))
	}
	if assigned := c.Query("assigned_to"); assigned != "" {
		if v, err := strconv.Atoi(assigned); err == nil {
			query = query.Where("assigned_to = ?", v)
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
	case "status":
		query = query.Order("status " + order).Order("created_at DESC")
	case "assigned":
		query = query.Order("assigned_to " + order).Order("created_at DESC")
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

// CreateTicket creates a new support ticket
func (h *SupportHandler) CreateTicket(c *gin.Context) {
	var request CreateTicketRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.GenerateBadRequestResponse(c, "support/create-ticket", err.Error())
		return
	}

	// Get user ID from context (assuming middleware sets it)
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "support/create-ticket", "User not authenticated")
		return
	}

	// Create the ticket
	ticket := models.SupportTicket{
		UserID:      userID.(uint),
		Title:       request.Title,
		Description: request.Description,
		Category:    request.Category,
		Priority:    request.Priority,
		Status:      models.TicketStatusOpen,
		OrderID:     request.OrderID,
	}

	if err := h.db.Create(&ticket).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/create-ticket", err.Error())
		return
	}

	// Handle attachments if provided
	if len(request.Attachments) > 0 {
		for _, attachment := range request.Attachments {
			ticketAttachment := models.TicketAttachment{
				TicketID: ticket.ID,
				FileName: attachment.FileName,
				FileURL:  attachment.FileURL,
				FileSize: attachment.FileSize,
				FileType: attachment.FileType,
			}
			if err := h.db.Create(&ticketAttachment).Error; err != nil {
				response.GenerateInternalServerErrorResponse(c, "support/create-ticket", "Failed to create attachment")
				return
			}
		}
	}

	// Load the created ticket with relationships
	if err := h.db.Preload("User").Preload("Order").Preload("Attachments").First(&ticket, ticket.ID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/create-ticket", "Failed to load created ticket")
		return
	}

	response.GenerateSuccessResponse(c, "Support ticket created successfully", ticket)
}

// GetTicket retrieves a specific support ticket
func (h *SupportHandler) GetTicket(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "support/get-ticket", "Invalid ticket ID")
		return
	}

	var ticket models.SupportTicket
	if err := h.db.Preload("User").Preload("Order").Preload("Attachments").Preload("Responses.User").First(&ticket, ticketID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "support/get-ticket", "Ticket not found")
		return
	}

	// Check if user has permission to view this ticket
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "support/get-ticket", "User not authenticated")
		return
	}

	// Only allow users to view their own tickets or admins to view any ticket
	if ticket.UserID != userID.(uint) {
		userType, _ := c.Get("user_type")
		if userType != models.Admin {
			response.GenerateForbiddenResponse(c, "support/get-ticket", "Access denied")
			return
		}
	}

	response.GenerateSuccessResponse(c, "Ticket retrieved successfully", ticket)
}

// GetUserTickets retrieves all tickets for the current user
func (h *SupportHandler) GetUserTickets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "support/get-user-tickets", "User not authenticated")
		return
	}

	var tickets []models.SupportTicket
	q := h.db.Where("user_id = ?", userID).Model(&models.SupportTicket{})
	q, _, _ = h.applyTicketFilters(c, q)
	if err := q.Preload("User").Preload("Order").Order("created_at DESC").Find(&tickets).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/get-user-tickets", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "User tickets retrieved successfully", tickets)
}

// GetAllTickets retrieves all tickets (admin only)
func (h *SupportHandler) GetAllTickets(c *gin.Context) {
	userType, exists := c.Get("user_type")
	if !exists || userType != models.Admin {
		response.GenerateForbiddenResponse(c, "support/get-all-tickets", "Admin access required")
		return
	}

	var tickets []models.SupportTicket
	q := h.db.Model(&models.SupportTicket{})
	q, _, _ = h.applyTicketFilters(c, q)
	if err := q.Preload("User").Preload("Order").Find(&tickets).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/get-all-tickets", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "All tickets retrieved successfully", tickets)
}

// UpdateTicket updates a support ticket
func (h *SupportHandler) UpdateTicket(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "support/update-ticket", "Invalid ticket ID")
		return
	}

	var request UpdateTicketRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.GenerateBadRequestResponse(c, "support/update-ticket", err.Error())
		return
	}

	var ticket models.SupportTicket
	if err := h.db.First(&ticket, ticketID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "support/update-ticket", "Ticket not found")
		return
	}

	// Check permissions
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "support/update-ticket", "User not authenticated")
		return
	}

	userType, _ := c.Get("user_type")
	if ticket.UserID != userID.(uint) && userType != models.Admin {
		response.GenerateForbiddenResponse(c, "support/update-ticket", "Access denied")
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
		if request.Status == models.TicketStatusResolved || request.Status == models.TicketStatusClosed {
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

	if err := h.db.Model(&ticket).Updates(updates).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/update-ticket", err.Error())
		return
	}

	// send status update email if status changed
	if _, ok := updates["status"]; ok && h.emailTriggerSvc != nil {
		// Load user for email
		var user models.User
		if err := h.db.First(&user, ticket.UserID).Error; err == nil {
			data := map[string]interface{}{
				"UserName":        strings.TrimSpace(user.FirstName + " " + user.LastName),
				"TicketID":        ticket.ID,
				"TicketTitle":     ticket.Title,
				"OldStatus":       "",
				"NewStatus":       updates["status"],
				"UserMessageHTML": template.HTML(ticket.Description),
				"AdminNoteHTML":   template.HTML(fmt.Sprintf("%v", updates["internal_notes"])),
				"subject":         fmt.Sprintf("Your ticket #%d status updated", ticket.ID),
			}
			_ = h.emailTriggerSvc.TriggerTicketStatusUpdated(user.Email, data["UserName"].(string), data)
		}
	}

	// Load updated ticket
	if err := h.db.Preload("User").Preload("Order").Preload("Attachments").Preload("Responses.User").First(&ticket, ticketID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/update-ticket", "Failed to load updated ticket")
		return
	}

	response.GenerateSuccessResponse(c, "Ticket updated successfully", ticket)
}

// AddTicketResponse adds a response to a support ticket
func (h *SupportHandler) AddTicketResponse(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "support/add-ticket-response", "Invalid ticket ID")
		return
	}

	var request TicketResponseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.GenerateBadRequestResponse(c, "support/add-ticket-response", err.Error())
		return
	}

	var ticket models.SupportTicket
	if err := h.db.First(&ticket, ticketID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "support/add-ticket-response", "Ticket not found")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "support/add-ticket-response", "User not authenticated")
		return
	}

	userType, _ := c.Get("user_type")
	isAdmin := userType == models.Admin

	// Check permissions
	if ticket.UserID != userID.(uint) && !isAdmin {
		response.GenerateForbiddenResponse(c, "support/add-ticket-response", "Access denied")
		return
	}

	// Create response
	ticketResponse := models.TicketResponse{
		TicketID:    uint(ticketID),
		UserID:      userID.(uint),
		Message:     request.Message,
		IsInternal:  request.IsInternal,
		IsFromAdmin: isAdmin,
	}

	if err := h.db.Create(&ticketResponse).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/add-ticket-response", err.Error())
		return
	}

	// Email notify ticket owner on non-internal responses
	if !request.IsInternal && h.emailTriggerSvc != nil {
		var owner models.User
		if err := h.db.First(&owner, ticket.UserID).Error; err == nil {
			// responder name
			responderName := "User"
			var responder models.User
			if err := h.db.First(&responder, userID.(uint)).Error; err == nil {
				responderName = strings.TrimSpace(responder.FirstName + " " + responder.LastName)
			}
			data := map[string]interface{}{
				"UserName":        strings.TrimSpace(owner.FirstName + " " + owner.LastName),
				"TicketID":        ticket.ID,
				"TicketTitle":     ticket.Title,
				"UserMessageHTML": template.HTML(ticket.Description),
				"ResponderName":   responderName,
				"RespondedAt":     time.Now().Format("2006-01-02 15:04:05"),
				"ResponseHTML":    template.HTML(request.Message),
				"subject":         fmt.Sprintf("New response on your ticket #%d", ticket.ID),
			}
			_ = h.emailTriggerSvc.TriggerTicketResponse(owner.Email, data["UserName"].(string), data)
		}
	}

	// Update ticket status if admin responded
	if isAdmin && ticket.Status == models.TicketStatusOpen {
		h.db.Model(&ticket).Update("status", models.TicketStatusInProgress)
	}

	response.GenerateSuccessResponse(c, "Response added successfully", ticketResponse)
}

// DeleteTicket deletes a support ticket (admin only)
func (h *SupportHandler) DeleteTicket(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "support/delete-ticket", "Invalid ticket ID")
		return
	}

	userType, exists := c.Get("user_type")
	if !exists || userType != models.Admin {
		response.GenerateForbiddenResponse(c, "support/delete-ticket", "Admin access required")
		return
	}

	var ticket models.SupportTicket
	if err := h.db.First(&ticket, ticketID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "support/delete-ticket", "Ticket not found")
		return
	}

	if err := h.db.Delete(&ticket).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/delete-ticket", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Ticket deleted successfully", nil)
}
