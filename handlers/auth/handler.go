package auth

import (
	"github.com/YasserCherfaoui/MarketProGo/email"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db              *gorm.DB
	emailTriggerSvc *email.EmailTriggerService
}

func NewAuthHandler(db *gorm.DB, emailTriggerSvc *email.EmailTriggerService) *AuthHandler {
	return &AuthHandler{
		db:              db,
		emailTriggerSvc: emailTriggerSvc,
	}
}
