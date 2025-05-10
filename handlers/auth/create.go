package auth

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/password"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Email     string          `json:"email" binding:"required,email"`
	Password  string          `json:"password" binding:"required,min=8"`
	FirstName string          `json:"first_name" binding:"required"`
	LastName  string          `json:"last_name" binding:"required"`
	Phone     string          `json:"phone"`
	UserType  models.UserType `json:"user_type" binding:"required"`
}

func (h *AuthHandler) CreateUser(c *gin.Context) {
	var request RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.GenerateBadRequestResponse(c, "auth/create-user", err.Error())
		return
	}

	hashedPassword, err := password.Hash(request.Password)
	if err != nil {
		response.GenerateInternalServerErrorResponse(c, "auth/create-user", err.Error())
		return
	}

	user := models.User{
		Email:     request.Email,
		Password:  hashedPassword,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Phone:     request.Phone,
		UserType:  request.UserType,
	}

	if err := h.db.Create(&user).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "auth/create-user", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "User created successfully", user)
}
