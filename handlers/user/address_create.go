package user

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type CreateAddressRequest struct {
	StreetAddress1 string `json:"street_address1" binding:"required"`
	StreetAddress2 string `json:"street_address2"`
	City           string `json:"city" binding:"required"`
	State          string `json:"state"`
	PostalCode     string `json:"postal_code" binding:"required"`
	Country        string `json:"country" binding:"required"`
	IsDefault      bool   `json:"is_default"`
}

func (h *UserHandler) CreateAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "user/create_address", "User not authenticated")
		return
	}
	uid := userID.(uint)

	var req CreateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "user/create_address", err.Error())
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// If this is set as default, unset other default addresses for this user
	if req.IsDefault {
		if err := tx.Model(&models.Address{}).
			Where("user_id = ? AND is_default = ?", uid, true).
			Update("is_default", false).Error; err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "user/create_address", "Failed to update existing default addresses")
			return
		}
	}

	// Create new address
	address := models.Address{
		StreetAddress1: req.StreetAddress1,
		StreetAddress2: req.StreetAddress2,
		City:           req.City,
		State:          req.State,
		PostalCode:     req.PostalCode,
		Country:        req.Country,
		IsDefault:      req.IsDefault,
		UserID:         &uid,
	}

	if err := tx.Create(&address).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "user/create_address", "Failed to create address")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "user/create_address", "Failed to commit transaction")
		return
	}

	response.GenerateCreatedResponse(c, "Address created successfully", address)
}
