package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/password"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// POST /auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "auth/forgot-password", err.Error())
		return
	}

	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Do not reveal user existence
		response.GenerateSuccessResponse(c, "If that email is registered, you will receive a reset email shortly", nil)
		return
	}

	// generate token (random)
	raw := fmt.Sprintf("%d:%d", user.ID, time.Now().UnixNano())
	hash := hashToken(raw)
	expires := time.Now().Add(24 * time.Hour)
	record := models.PasswordResetToken{UserID: user.ID, TokenHash: hash, ExpiresAt: expires}
	if err := h.db.Create(&record).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "auth/forgot-password", "Failed to create reset token")
		return
	}

	// send email
	if h.emailTriggerSvc != nil {
		name := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
		_ = h.emailTriggerSvc.TriggerPasswordReset(user.Email, name, raw)
	}

	response.GenerateSuccessResponse(c, "If that email is registered, you will receive a reset email shortly", nil)
}

// GET /auth/verify-reset-token?token=...
func (h *AuthHandler) VerifyResetToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.GenerateBadRequestResponse(c, "auth/verify-reset-token", "Missing token")
		return
	}
	hash := hashToken(token)
	var rec models.PasswordResetToken
	if err := h.db.Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", hash, time.Now()).First(&rec).Error; err != nil {
		response.GenerateUnauthorizedResponse(c, "auth/verify-reset-token", "Invalid or expired token")
		return
	}
	response.GenerateSuccessResponse(c, "Token is valid", nil)
}

// POST /auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "auth/reset-password", err.Error())
		return
	}
	hash := hashToken(req.Token)
	var rec models.PasswordResetToken
	if err := h.db.Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", hash, time.Now()).First(&rec).Error; err != nil {
		response.GenerateUnauthorizedResponse(c, "auth/reset-password", "Invalid or expired token")
		return
	}

	// update password
	hashed, err := password.Hash(req.NewPassword)
	if err != nil {
		response.GenerateInternalServerErrorResponse(c, "auth/reset-password", "Failed to hash password")
		return
	}
	if err := h.db.Model(&models.User{}).Where("id = ?", rec.UserID).Update("password", hashed).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "auth/reset-password", "Failed to update password")
		return
	}
	// mark token used
	now := time.Now()
	_ = h.db.Model(&rec).Update("used_at", &now).Error

	response.GenerateSuccessResponse(c, "Password reset successful", nil)
}
