package support

import (
	"strconv"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
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
	if (contactInquiry.UserID == nil || *contactInquiry.UserID != userID.(uint)) && userType != "ADMIN" {
		response.GenerateForbiddenResponse(c, "support/get-contact-inquiry", "Access denied")
		return
	}

	response.GenerateSuccessResponse(c, "Contact inquiry retrieved successfully", contactInquiry)
}

// GetUserContactInquiries retrieves all contact inquiries for the current user
func (h *SupportHandler) GetUserContactInquiries(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "support/get-user-contact-inquiries", "User not authenticated")
		return
	}

	var contactInquiries []models.ContactInquiry
	if err := h.db.Where("user_id = ?", userID).Preload("User").Order("created_at DESC").Find(&contactInquiries).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/get-user-contact-inquiries", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "User contact inquiries retrieved successfully", contactInquiries)
}

// GetAllContactInquiries retrieves all contact inquiries (admin only)
func (h *SupportHandler) GetAllContactInquiries(c *gin.Context) {
	userType, exists := c.Get("user_type")
	if !exists || userType != "ADMIN" {
		response.GenerateForbiddenResponse(c, "support/get-all-contact-inquiries", "Admin access required")
		return
	}

	var contactInquiries []models.ContactInquiry
	if err := h.db.Preload("User").Order("created_at DESC").Find(&contactInquiries).Error; err != nil {
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
	if !exists || userType != "ADMIN" {
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
	if !exists || userType != "ADMIN" {
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
