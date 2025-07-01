package user

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpdateAddressRequest struct {
	StreetAddress1 *string `json:"street_address1"`
	StreetAddress2 *string `json:"street_address2"`
	City           *string `json:"city"`
	State          *string `json:"state"`
	PostalCode     *string `json:"postal_code"`
	Country        *string `json:"country"`
	IsDefault      *bool   `json:"is_default"`
}

func (h *UserHandler) UpdateAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "user/update_address", "User not authenticated")
		return
	}
	uid := userID.(uint)

	addressID := c.Param("id")
	if addressID == "" {
		response.GenerateBadRequestResponse(c, "user/update_address", "Address ID is required")
		return
	}

	var req UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "user/update_address", err.Error())
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the existing address
	var address models.Address
	if err := tx.Where("id = ? AND user_id = ?", addressID, uid).
		First(&address).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			response.GenerateNotFoundResponse(c, "user/update_address", "Address not found")
		} else {
			response.GenerateInternalServerErrorResponse(c, "user/update_address", "Failed to get address")
		}
		return
	}

	// If setting this as default, unset other default addresses
	if req.IsDefault != nil && *req.IsDefault {
		if err := tx.Model(&models.Address{}).
			Where("user_id = ? AND id != ? AND is_default = ?", uid, addressID, true).
			Update("is_default", false).Error; err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "user/update_address", "Failed to update existing default addresses")
			return
		}
	}

	// Update fields if provided
	updates := make(map[string]interface{})
	if req.StreetAddress1 != nil {
		updates["street_address1"] = *req.StreetAddress1
	}
	if req.StreetAddress2 != nil {
		updates["street_address2"] = *req.StreetAddress2
	}
	if req.City != nil {
		updates["city"] = *req.City
	}
	if req.State != nil {
		updates["state"] = *req.State
	}
	if req.PostalCode != nil {
		updates["postal_code"] = *req.PostalCode
	}
	if req.Country != nil {
		updates["country"] = *req.Country
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}

	if len(updates) == 0 {
		tx.Rollback()
		response.GenerateBadRequestResponse(c, "user/update_address", "No fields to update")
		return
	}

	// Update the address
	if err := tx.Model(&address).Updates(updates).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "user/update_address", "Failed to update address")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "user/update_address", "Failed to commit transaction")
		return
	}

	// Get updated address
	if err := h.db.First(&address, address.ID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "user/update_address", "Address updated but failed to load details")
		return
	}

	response.GenerateSuccessResponse(c, "Address updated successfully", address)
}
