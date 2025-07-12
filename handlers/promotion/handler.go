package promotion

import (
	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"gorm.io/gorm"
)

type PromotionHandler struct {
	db              *gorm.DB
	gcsService      *gcs.GCService
	appwriteService *aw.AppwriteService
}

func NewPromotionHandler(db *gorm.DB, gcsService *gcs.GCService, appwriteService *aw.AppwriteService) *PromotionHandler {
	return &PromotionHandler{db: db, gcsService: gcsService, appwriteService: appwriteService}
}
