package auth

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/auth"
	"github.com/YasserCherfaoui/MarketProGo/utils/password"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.GenerateBadRequestResponse(c, "auth/login", err.Error())
		return
	}

	user := models.User{}
	if err := h.db.Where("email = ?", request.Email).First(&user).Error; err != nil {
		response.GenerateNotFoundResponse(c, "auth/login", "User not found")
		return
	}

	if !password.Validate(request.Password, user.Password) {
		response.GenerateUnauthorizedResponse(c, "auth/login", "Invalid password")
		return
	}

	token, err := auth.GenerateToken(user.ID, user.UserType, user.CompanyID)
	if err != nil {
		response.GenerateInternalServerErrorResponse(c, "auth/login", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Login successful", LoginResponse{
		Token: token,
		User:  user,
	})
}
