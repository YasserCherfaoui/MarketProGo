package user

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// CreateSellerRequest holds all data needed to create a seller and their company
type CreateSellerRequest struct {
	// User fields
	Email     string          `json:"email" binding:"required,email"`
	Password  string          `json:"password" binding:"required,min=8"`
	FirstName string          `json:"first_name" binding:"required"`
	LastName  string          `json:"last_name" binding:"required"`
	Phone     string          `json:"phone" binding:"required"`
	UserType  models.UserType `json:"user_type" binding:"required,oneof=WHOLESALER VENDOR"`
	Role      string          `json:"role" binding:"required"`

	// Company fields
	Company struct {
		Name               string  `json:"name" binding:"required"`
		VATNumber          string  `json:"vat_number" binding:"required"`
		RegistrationNumber string  `json:"registration_number" binding:"required"`
		Phone              string  `json:"phone" binding:"required"`
		Email              string  `json:"email" binding:"required,email"`
		Website            string  `json:"website"`
		CreditLimit        float64 `json:"credit_limit"`
		PaymentTerms       int     `json:"payment_terms"`

		// Company address
		Address struct {
			StreetAddress1 string `json:"street_address1" binding:"required"`
			StreetAddress2 string `json:"street_address2"`
			City           string `json:"city" binding:"required"`
			State          string `json:"state"`
			PostalCode     string `json:"postal_code" binding:"required"`
			Country        string `json:"country" binding:"required"`
		} `json:"address" binding:"required"`
	} `json:"company" binding:"required"`
}

// CreateSeller handles creating a seller (WHOLESALER or VENDOR) with associated company
func (h *UserHandler) CreateSeller(c *gin.Context) {
	var req CreateSellerRequest

	// Validate input
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "user/create_seller", err.Error())
		return
	}

	// Verify user type is valid for a seller
	if req.UserType != models.Wholesaler && req.UserType != models.Vendor {
		response.GenerateBadRequestResponse(c, "user/create_seller", "Invalid user type")
		return
	}

	// Begin transaction
	tx := h.db.Begin()

	// Create company first (without address)
	company := models.Company{
		Name:               req.Company.Name,
		VATNumber:          req.Company.VATNumber,
		RegistrationNumber: req.Company.RegistrationNumber,
		Phone:              req.Company.Phone,
		Email:              req.Company.Email,
		Website:            req.Company.Website,
		CreditLimit:        req.Company.CreditLimit,
		PaymentTerms:       req.Company.PaymentTerms,
		IsVerified:         false, // Default to false, admin will verify later
	}

	if err := tx.Create(&company).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "user/create_seller", err.Error())
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "user/create_seller", "Failed to secure password")
		return
	}

	// Create user
	user := models.User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		UserType:  req.UserType,
		CompanyID: &company.ID,
		Role:      req.Role,
		IsActive:  true,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "user/create_seller", "Failed to create user")
		return
	}

	// Now create company address with proper foreign keys
	companyAddress := models.Address{
		StreetAddress1: req.Company.Address.StreetAddress1,
		StreetAddress2: req.Company.Address.StreetAddress2,
		City:           req.Company.Address.City,
		State:          req.Company.Address.State,
		PostalCode:     req.Company.Address.PostalCode,
		Country:        req.Company.Address.Country,
		IsDefault:      true,
		UserID:         &user.ID, // Link address to the user
	}

	if err := tx.Create(&companyAddress).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "user/create_seller", err.Error())
		return
	}

	// Update company with the address ID
	if err := tx.Model(&company).Update("address_id", companyAddress.ID).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "user/create_seller", "Failed to update company with address")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "user/create_seller", "Failed to complete registration")
		return
	}

	// Return success with user ID
	response.GenerateSuccessResponse(c, "Seller account created successfully", gin.H{
		"user_id":    user.ID,
		"company_id": company.ID,
	})
}
