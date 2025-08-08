package support

import (
	"strconv"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

// CreateAbuseReportRequest represents the request to create an abuse report
type CreateAbuseReportRequest struct {
	ReportedUserID *uint                    `json:"reported_user_id,omitempty"`
	ProductID      *uint                    `json:"product_id,omitempty"`
	ReviewID       *uint                    `json:"review_id,omitempty"`
	OrderID        *uint                    `json:"order_id,omitempty"`
	Category       models.AbuseCategory     `json:"category" binding:"required"`
	Description    string                   `json:"description" binding:"required"`
	Severity       models.AbuseSeverity     `json:"severity"`
	Attachments    []AbuseAttachmentRequest `json:"attachments,omitempty"`
}

// AbuseAttachmentRequest represents an attachment for an abuse report
type AbuseAttachmentRequest struct {
	FileName string `json:"file_name" binding:"required"`
	FileURL  string `json:"file_url" binding:"required"`
	FileSize int64  `json:"file_size"`
	FileType string `json:"file_type"`
}

// UpdateAbuseReportRequest represents the request to update an abuse report
type UpdateAbuseReportRequest struct {
	Status        models.AbuseReportStatus `json:"status,omitempty"`
	Severity      models.AbuseSeverity     `json:"severity,omitempty"`
	Resolution    string                   `json:"resolution,omitempty"`
	InternalNotes string                   `json:"internal_notes,omitempty"`
}

// CreateAbuseReport creates a new abuse report
func (h *SupportHandler) CreateAbuseReport(c *gin.Context) {
	var request CreateAbuseReportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.GenerateBadRequestResponse(c, "support/create-abuse-report", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "support/create-abuse-report", "User not authenticated")
		return
	}

	// Create the abuse report
	abuseReport := models.AbuseReport{
		ReporterID:     userID.(uint),
		ReportedUserID: request.ReportedUserID,
		ProductID:      request.ProductID,
		ReviewID:       request.ReviewID,
		OrderID:        request.OrderID,
		Category:       request.Category,
		Description:    request.Description,
		Severity:       request.Severity,
		Status:         models.AbuseReportStatusPending,
	}

	if err := h.db.Create(&abuseReport).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/create-abuse-report", err.Error())
		return
	}

	// Handle attachments if provided
	if len(request.Attachments) > 0 {
		for _, attachment := range request.Attachments {
			abuseAttachment := models.AbuseReportAttachment{
				AbuseReportID: abuseReport.ID,
				FileName:      attachment.FileName,
				FileURL:       attachment.FileURL,
				FileSize:      attachment.FileSize,
				FileType:      attachment.FileType,
			}
			if err := h.db.Create(&abuseAttachment).Error; err != nil {
				response.GenerateInternalServerErrorResponse(c, "support/create-abuse-report", "Failed to create attachment")
				return
			}
		}
	}

	// Load the created abuse report with relationships
	if err := h.db.Preload("Reporter").Preload("ReportedUser").Preload("Product").Preload("Review").Preload("Order").Preload("Attachments").First(&abuseReport, abuseReport.ID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/create-abuse-report", "Failed to load created abuse report")
		return
	}

	response.GenerateSuccessResponse(c, "Abuse report created successfully", abuseReport)
}

// GetAbuseReport retrieves a specific abuse report
func (h *SupportHandler) GetAbuseReport(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "support/get-abuse-report", "Invalid report ID")
		return
	}

	var abuseReport models.AbuseReport
	if err := h.db.Preload("Reporter").Preload("ReportedUser").Preload("Product").Preload("Review").Preload("Order").Preload("Attachments").First(&abuseReport, reportID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "support/get-abuse-report", "Abuse report not found")
		return
	}

	// Check permissions
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "support/get-abuse-report", "User not authenticated")
		return
	}

	userType, _ := c.Get("user_type")
	if abuseReport.ReporterID != userID.(uint) && userType != "ADMIN" {
		response.GenerateForbiddenResponse(c, "support/get-abuse-report", "Access denied")
		return
	}

	response.GenerateSuccessResponse(c, "Abuse report retrieved successfully", abuseReport)
}

// GetUserAbuseReports retrieves all abuse reports for the current user
func (h *SupportHandler) GetUserAbuseReports(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "support/get-user-abuse-reports", "User not authenticated")
		return
	}

	var abuseReports []models.AbuseReport
	if err := h.db.Where("reporter_id = ?", userID).Preload("Reporter").Preload("ReportedUser").Preload("Product").Preload("Review").Preload("Order").Order("created_at DESC").Find(&abuseReports).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/get-user-abuse-reports", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "User abuse reports retrieved successfully", abuseReports)
}

// GetAllAbuseReports retrieves all abuse reports (admin only)
func (h *SupportHandler) GetAllAbuseReports(c *gin.Context) {
	userType, exists := c.Get("user_type")
	if !exists || userType != "ADMIN" {
		response.GenerateForbiddenResponse(c, "support/get-all-abuse-reports", "Admin access required")
		return
	}

	var abuseReports []models.AbuseReport
	if err := h.db.Preload("Reporter").Preload("ReportedUser").Preload("Product").Preload("Review").Preload("Order").Order("created_at DESC").Find(&abuseReports).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/get-all-abuse-reports", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "All abuse reports retrieved successfully", abuseReports)
}

// UpdateAbuseReport updates an abuse report
func (h *SupportHandler) UpdateAbuseReport(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "support/update-abuse-report", "Invalid report ID")
		return
	}

	var request UpdateAbuseReportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.GenerateBadRequestResponse(c, "support/update-abuse-report", err.Error())
		return
	}

	var abuseReport models.AbuseReport
	if err := h.db.First(&abuseReport, reportID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "support/update-abuse-report", "Abuse report not found")
		return
	}

	// Only admins can update abuse reports
	userType, exists := c.Get("user_type")
	if !exists || userType != "ADMIN" {
		response.GenerateForbiddenResponse(c, "support/update-abuse-report", "Admin access required")
		return
	}

	// Update fields
	updates := make(map[string]interface{})
	if request.Status != "" {
		updates["status"] = request.Status
		if request.Status == models.AbuseReportStatusResolved || request.Status == models.AbuseReportStatusDismissed {
			now := time.Now()
			updates["resolved_at"] = &now
			userID, _ := c.Get("user_id")
			updates["resolved_by"] = userID
		}
	}
	if request.Severity != "" {
		updates["severity"] = request.Severity
	}
	if request.Resolution != "" {
		updates["resolution"] = request.Resolution
	}
	if request.InternalNotes != "" {
		updates["internal_notes"] = request.InternalNotes
	}

	if err := h.db.Model(&abuseReport).Updates(updates).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/update-abuse-report", err.Error())
		return
	}

	// Load updated abuse report
	if err := h.db.Preload("Reporter").Preload("ReportedUser").Preload("Product").Preload("Review").Preload("Order").Preload("Attachments").First(&abuseReport, reportID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/update-abuse-report", "Failed to load updated abuse report")
		return
	}

	response.GenerateSuccessResponse(c, "Abuse report updated successfully", abuseReport)
}

// DeleteAbuseReport deletes an abuse report (admin only)
func (h *SupportHandler) DeleteAbuseReport(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "support/delete-abuse-report", "Invalid report ID")
		return
	}

	userType, exists := c.Get("user_type")
	if !exists || userType != "ADMIN" {
		response.GenerateForbiddenResponse(c, "support/delete-abuse-report", "Admin access required")
		return
	}

	var abuseReport models.AbuseReport
	if err := h.db.First(&abuseReport, reportID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "support/delete-abuse-report", "Abuse report not found")
		return
	}

	if err := h.db.Delete(&abuseReport).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "support/delete-abuse-report", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Abuse report deleted successfully", nil)
}
