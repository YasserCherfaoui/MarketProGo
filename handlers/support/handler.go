package support

import (
	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/email"
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"gorm.io/gorm"
)

// SupportHandler handles all support-related operations
type SupportHandler struct {
	db              *gorm.DB
	gcsService      *gcs.GCService
	appwriteService *aw.AppwriteService
	emailTriggerSvc *email.EmailTriggerService
}

// NewSupportHandler creates a new support handler
func NewSupportHandler(db *gorm.DB, gcsService *gcs.GCService, appwriteService *aw.AppwriteService, emailTriggerSvc *email.EmailTriggerService) *SupportHandler {
	return &SupportHandler{
		db:              db,
		gcsService:      gcsService,
		appwriteService: appwriteService,
		emailTriggerSvc: emailTriggerSvc,
	}
}
