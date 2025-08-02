package order

import (
	"github.com/YasserCherfaoui/MarketProGo/email"
	"gorm.io/gorm"
)

type OrderHandler struct {
	db              *gorm.DB
	emailTriggerSvc *email.EmailTriggerService
}

func NewOrderHandler(db *gorm.DB, emailTriggerSvc *email.EmailTriggerService) *OrderHandler {
	return &OrderHandler{
		db:              db,
		emailTriggerSvc: emailTriggerSvc,
	}
}
