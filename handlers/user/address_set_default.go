package user

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *UserHandler) SetDefaultAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "user/set_default_address", "User not authenticated")
		return
	}
	uid := userID.(uint)

	addressID := c.Param("id")
	if addressID == "" {
		response.GenerateBadRequestResponse(c, "user/set_default_address", "Address ID is required")
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if the address exists and belongs to user
	var address models.Address
	if err := tx.Where("id = ? AND user_id = ?", addressID, uid).
		First(&address).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			response.GenerateNotFoundResponse(c, "user/set_default_address", "Address not found")
		} else {
			response.GenerateInternalServerErrorResponse(c, "user/set_default_address", "Failed to get address")
		}
		return
	}

	// If already default, no need to update
	if address.IsDefault {
		tx.Rollback()
		response.GenerateSuccessResponse(c, "Address is already set as default", address)
		return
	}

	// Unset other default addresses for this user
	if err := tx.Model(&models.Address{}).
		Where("user_id = ? AND is_default = ?", uid, true).
		Update("is_default", false).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "user/set_default_address", "Failed to update existing default addresses")
		return
	}

	// Set this address as default
	if err := tx.Model(&address).Update("is_default", true).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "user/set_default_address", "Failed to set address as default")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "user/set_default_address", "Failed to commit transaction")
		return
	}

	// Get updated address
	if err := h.db.First(&address, address.ID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "user/set_default_address", "Address updated but failed to load details")
		return
	}

	response.GenerateSuccessResponse(c, "Default address set successfully", address)
}
