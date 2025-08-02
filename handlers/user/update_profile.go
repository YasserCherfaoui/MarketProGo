package user

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpdateProfileRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
	Avatar    string `json:"avatar"`
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		response.GenerateInternalServerErrorResponse(c, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	// Parse request body
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "INVALID_REQUEST", "Invalid request body: "+err.Error())
		return
	}

	// Validate required fields
	if req.FirstName == "" || req.LastName == "" {
		response.GenerateBadRequestResponse(c, "MISSING_REQUIRED_FIELDS", "First name and last name are required")
		return
	}

	// Get user from database
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateNotFoundResponse(c, "USER_NOT_FOUND", "User not found")
			return
		}
		response.GenerateInternalServerErrorResponse(c, "DB_ERROR", "Failed to fetch user")
		return
	}

	// Update user profile
	updates := map[string]interface{}{
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"phone":      req.Phone,
		"avatar":     req.Avatar,
	}

	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "UPDATE_FAILED", "Failed to update profile")
		return
	}

	// Return updated user data
	response.GenerateSuccessResponse(c, "Profile updated successfully", gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"phone":      user.Phone,
			"avatar":     user.Avatar,
			"user_type":  user.UserType,
			"is_active":  user.IsActive,
		},
	})
}
