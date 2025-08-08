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

// CreateContactInquiryRequest represents the request to create a contact inquiry
type CreateContactInquiryRequest struct {
	Name     string                 `json:"name" binding:"required"`
	Email    string                 `json:"email" binding:"required,email"`
	Phone    string                 `json:"phone"`
	Subject  string                 `json:"subject" binding:"required"`
	Message  string                 `json:"message" binding:"required"`
	Category models.ContactCategory `json:"category" binding:"required"`
	Priority models.ContactPriority `json:"priority"`
}

// UpdateContactInquiryRequest represents the request to update a contact inquiry
type UpdateContactInquiryRequest struct {
	Status        models.ContactStatus   `json:"status,omitempty"`
	Priority      models.ContactPriority `json:"priority,omitempty"`
	Response      string                 `json:"response,omitempty"`
	InternalNotes string                 `json:"internal_notes,omitempty"`
}

// ReplyToContactInquiryRequest represents admin reply payload
type ReplyToContactInquiryRequest struct {
	ResponseHTML string `json:"response_html" binding:"required"`
}

// CreateContactInquiry creates a new contact inquiry
func (h *SupportHandler) CreateContactInquiry(c *gin.Context) {
	var request CreateContactInquiryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.GenerateBadRequestResponse(c, "support/create-contact-inquiry", err.Error())
		return
	}

	// Get user ID from context (optional - contact inquiries can be from non-authenticated users)
	var userID *uint
	if userIDVal, exists := c.Get("user_id"); exists {
		userIDUint := userIDVal.(uint)
		userID = &userIDUint
	}

	// Create the contact inquiry
	contactInquiry := models.ContactInquiry{
		UserID:   userID,
		Name:     request.Name,
		Email:    request.Email,
		Phone:    request.Phone,
		Subject:  request.Subject,
		Message:  request.Message,
		Category: request.Category,
		Priority: request.Priority,
		Status:   models.ContactStatusNew,
	}

	if err := h.db.Create(&contactInquiry).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/create-contact-inquiry", err.Error())
		return
	}

	// Load the created contact inquiry with relationships
	if err := h.db.Preload("User").First(&contactInquiry, contactInquiry.ID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/create-contact-inquiry", "Failed to load created contact inquiry")
		return
	}

	response.GenerateSuccessResponse(c, "Contact inquiry submitted successfully", contactInquiry)
}

// GetContactInquiry retrieves a specific contact inquiry
func (h *SupportHandler) GetContactInquiry(c *gin.Context) {
	inquiryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "support/get-contact-inquiry", "Invalid inquiry ID")
		return
	}

	var contactInquiry models.ContactInquiry
	if err := h.db.Preload("User").First(&contactInquiry, inquiryID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "support/get-contact-inquiry", "Contact inquiry not found")
		return
	}

	// Check permissions
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "support/get-contact-inquiry", "User not authenticated")
		return
	}

	userType, _ := c.Get("user_type")
	if (contactInquiry.UserID == nil || *contactInquiry.UserID != userID.(uint)) && userType != models.Admin {
		response.GenerateForbiddenResponse(c, "support/get-contact-inquiry", "Access denied")
		return
	}

	response.GenerateSuccessResponse(c, "Contact inquiry retrieved successfully", contactInquiry)
}

// applyContactFilters applies filters/sort/pagination to the contact queries
func (h *SupportHandler) applyContactFilters(c *gin.Context, query *gorm.DB) (*gorm.DB, int, int) {
	if status := c.Query("status"); status != "" {
		query = query.Where("status IN ?", strings.Split(strings.ToUpper(status), ","))
	}
	if category := c.Query("category"); category != "" {
		query = query.Where("category IN ?", strings.Split(strings.ToUpper(category), ","))
	}
	if priority := c.Query("priority"); priority != "" {
		query = query.Where("priority IN ?", strings.Split(strings.ToUpper(priority), ","))
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
	// Sorting
	sort := strings.ToLower(c.DefaultQuery("sort_by", "date"))
	order := strings.ToUpper(c.DefaultQuery("sort_order", "DESC"))
	if order != "ASC" {
		order = "DESC"
	}
	switch sort {
	case "date":
		query = query.Order("created_at " + order)
	case "priority":
		caseExpr := fmt.Sprintf("CASE priority WHEN 'LOW' THEN 1 WHEN 'NORMAL' THEN 2 WHEN 'HIGH' THEN 3 WHEN 'URGENT' THEN 4 END %s", order)
		query = query.Order(caseExpr).Order("created_at DESC")
	case "status":
		query = query.Order("status " + order).Order("created_at DESC")
	default:
		query = query.Order("created_at DESC")
	}
	// Pagination
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

// GetUserContactInquiries retrieves all contact inquiries for the current user
func (h *SupportHandler) GetUserContactInquiries(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "support/get-user-contact-inquiries", "User not authenticated")
		return
	}
	var contactInquiries []models.ContactInquiry
	q := h.db.Where("user_id = ?", userID).Model(&models.ContactInquiry{})
	q, _, _ = h.applyContactFilters(c, q)
	if err := q.Preload("User").Order("created_at DESC").Find(&contactInquiries).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/get-user-contact-inquiries", err.Error())
		return
	}
	response.GenerateSuccessResponse(c, "User contact inquiries retrieved successfully", contactInquiries)
}

// GetAllContactInquiries retrieves all contact inquiries (admin only)
func (h *SupportHandler) GetAllContactInquiries(c *gin.Context) {
	userType, exists := c.Get("user_type")
	if !exists || userType != models.Admin {
		response.GenerateForbiddenResponse(c, "support/get-all-contact-inquiries", "Admin access required")
		return
	}
	var contactInquiries []models.ContactInquiry
	q := h.db.Model(&models.ContactInquiry{})
	q, _, _ = h.applyContactFilters(c, q)
	if err := q.Preload("User").Find(&contactInquiries).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/get-all-contact-inquiries", err.Error())
		return
	}
	response.GenerateSuccessResponse(c, "All contact inquiries retrieved successfully", contactInquiries)
}

// UpdateContactInquiry updates a contact inquiry
func (h *SupportHandler) UpdateContactInquiry(c *gin.Context) {
	inquiryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "support/update-contact-inquiry", "Invalid inquiry ID")
		return
	}

	var request UpdateContactInquiryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.GenerateBadRequestResponse(c, "support/update-contact-inquiry", err.Error())
		return
	}

	var contactInquiry models.ContactInquiry
	if err := h.db.First(&contactInquiry, inquiryID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "support/update-contact-inquiry", "Contact inquiry not found")
		return
	}

	// Only admins can update contact inquiries
	userType, exists := c.Get("user_type")
	if !exists || userType != models.Admin {
		response.GenerateForbiddenResponse(c, "support/update-contact-inquiry", "Admin access required")
		return
	}

	// Update fields
	updates := make(map[string]interface{})
	if request.Status != "" {
		updates["status"] = request.Status
		if request.Status == models.ContactStatusResponded {
			now := time.Now()
			updates["responded_at"] = &now
			userID, _ := c.Get("user_id")
			updates["responded_by"] = userID
		}
	}
	if request.Priority != "" {
		updates["priority"] = request.Priority
	}
	if request.Response != "" {
		updates["response"] = request.Response
	}
	if request.InternalNotes != "" {
		updates["internal_notes"] = request.InternalNotes
	}

	if err := h.db.Model(&contactInquiry).Updates(updates).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/update-contact-inquiry", err.Error())
		return
	}

	// Email notify user on status change
	if _, ok := updates["status"]; ok && h.emailTriggerSvc != nil {
		name := contactInquiry.Name
		emailAddr := contactInquiry.Email
		if contactInquiry.UserID != nil {
			var u models.User
			if err := h.db.First(&u, *contactInquiry.UserID).Error; err == nil {
				if name == "" {
					name = strings.TrimSpace(u.FirstName + " " + u.LastName)
				}
				if emailAddr == "" {
					emailAddr = u.Email
				}
			}
		}
		data := map[string]interface{}{
			"Name":            name,
			"InquiryID":       contactInquiry.ID,
			"Subject":         contactInquiry.Subject,
			"OldStatus":       "",
			"NewStatus":       updates["status"],
			"UserMessageHTML": template.HTML(contactInquiry.Message),
			"Category":        string(contactInquiry.Category),
			"Priority":        string(contactInquiry.Priority),
			"AdminNoteHTML":   template.HTML(fmt.Sprintf("%v", updates["internal_notes"])),
			"subject":         fmt.Sprintf("Your inquiry #%d status updated", contactInquiry.ID),
		}
		_ = h.emailTriggerSvc.TriggerContactStatusUpdated(emailAddr, name, data)
	}

	// Load updated contact inquiry
	if err := h.db.Preload("User").First(&contactInquiry, inquiryID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/update-contact-inquiry", "Failed to load updated contact inquiry")
		return
	}

	response.GenerateSuccessResponse(c, "Contact inquiry updated successfully", contactInquiry)
}

// DeleteContactInquiry deletes a contact inquiry (admin only)
func (h *SupportHandler) DeleteContactInquiry(c *gin.Context) {
	inquiryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "support/delete-contact-inquiry", "Invalid inquiry ID")
		return
	}

	userType, exists := c.Get("user_type")
	if !exists || userType != models.Admin {
		response.GenerateForbiddenResponse(c, "support/delete-contact-inquiry", "Admin access required")
		return
	}

	var contactInquiry models.ContactInquiry
	if err := h.db.First(&contactInquiry, inquiryID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "support/delete-contact-inquiry", "Contact inquiry not found")
		return
	}

	if err := h.db.Delete(&contactInquiry).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/delete-contact-inquiry", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Contact inquiry deleted successfully", nil)
}

// ReplyToContactInquiry allows admin to reply and sends an email to the inquirer
func (h *SupportHandler) ReplyToContactInquiry(c *gin.Context) {
	// Admin check
	userType, exists := c.Get("user_type")
	if !exists || userType != models.Admin {
		response.GenerateForbiddenResponse(c, "support/reply-contact-inquiry", "Admin access required")
		return
	}

	inquiryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "support/reply-contact-inquiry", "Invalid inquiry ID")
		return
	}

	var req ReplyToContactInquiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "support/reply-contact-inquiry", err.Error())
		return
	}

	var inquiry models.ContactInquiry
	if err := h.db.Preload("User").First(&inquiry, inquiryID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "support/reply-contact-inquiry", "Contact inquiry not found")
		return
	}

	adminIDVal, _ := c.Get("user_id")
	adminID := adminIDVal.(uint)

	// Persist reply to inquiry
	now := time.Now()
	updates := map[string]interface{}{
		"response":     req.ResponseHTML,
		"responded_at": &now,
		"responded_by": adminID,
		"status":       models.ContactStatusResponded,
	}
	if err := h.db.Model(&inquiry).Updates(updates).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/reply-contact-inquiry", err.Error())
		return
	}

	// Prepare email to inquirer
	recipientEmail := inquiry.Email
	recipientName := inquiry.Name
	if recipientEmail == "" && inquiry.User != nil {
		recipientEmail = inquiry.User.Email
		if recipientName == "" {
			recipientName = inquiry.User.FirstName + " " + inquiry.User.LastName
		}
	}

	// Build template data
	adminName := "Admin"
	// Try to load admin user name if exists
	var admin models.User
	if err := h.db.First(&admin, adminID).Error; err == nil {
		if admin.FirstName != "" || admin.LastName != "" {
			adminName = strings.TrimSpace(admin.FirstName + " " + admin.LastName)
		}
	}

	data := map[string]interface{}{
		"Name":              recipientName,
		"Subject":           inquiry.Subject,
		"Category":          string(inquiry.Category),
		"UserMessageHTML":   template.HTML(inquiry.Message),
		"AdminResponseHTML": template.HTML(req.ResponseHTML),
		"AdminName":         adminName,
		"RespondedAt":       now.Format("2006-01-02 15:04:05"),
		"SupportEmail":      "enquirees@algeriamarket.co.uk",
		"InquiryID":         inquiry.ID,
		"subject":           "Response to your inquiry: " + inquiry.Subject,
	}

	recipient := models.EmailRecipient{Email: recipientEmail, Name: recipientName}

	// Use the email trigger service available via app routes - reuse EmailService through triggers if added to handler
	// Fallback: render and send via emailService directly is not accessible here; so we store through triggers in AppRoutes by exposing a method on handler
	if h.emailTriggerSvc != nil {
		if err := h.emailTriggerSvc.SendTemplateDirect("contact_inquiry_response", data, recipient, models.EmailTypeContactInquiryResponse); err != nil {
			// Soft fail email, but keep reply
			fmt.Printf("Failed to send inquiry response email: %v\n", err)
		}
	}

	// Return updated inquiry
	if err := h.db.Preload("User").First(&inquiry, inquiryID).Error; err != nil {
		response.GenerateSuccessResponse(c, "Reply saved but failed to reload inquiry", nil)
		return
	}

	response.GenerateSuccessResponse(c, "Inquiry replied and email sent", inquiry)
}
