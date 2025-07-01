package user

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *UserHandler) DeleteAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "user/delete_address", "User not authenticated")
		return
	}
	uid := userID.(uint)

	addressID := c.Param("id")
	if addressID == "" {
		response.GenerateBadRequestResponse(c, "user/delete_address", "Address ID is required")
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the address to check if it exists and belongs to user
	var address models.Address
	if err := tx.Where("id = ? AND user_id = ?", addressID, uid).
		First(&address).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			response.GenerateNotFoundResponse(c, "user/delete_address", "Address not found")
		} else {
			response.GenerateInternalServerErrorResponse(c, "user/delete_address", "Failed to get address")
		}
		return
	}

	// Check if this address is being used in any orders
	var orderCount int64
	if err := tx.Model(&models.Order{}).
		Where("shipping_address_id = ?", addressID).
		Count(&orderCount).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "user/delete_address", "Failed to check address usage")
		return
	}

	if orderCount > 0 {
		tx.Rollback()
		response.GenerateBadRequestResponse(c, "user/delete_address", "Cannot delete address that is used in orders")
		return
	}

	// Delete the address
	if err := tx.Delete(&address).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "user/delete_address", "Failed to delete address")
		return
	}

	// If this was the default address, set another address as default if available
	if address.IsDefault {
		var nextAddress models.Address
		if err := tx.Where("user_id = ?", uid).
			Order("created_at ASC").
			First(&nextAddress).Error; err == nil {
			// Set the oldest remaining address as default
			tx.Model(&nextAddress).Update("is_default", true)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "user/delete_address", "Failed to commit transaction")
		return
	}

	response.GenerateSuccessResponse(c, "Address deleted successfully", nil)
}
