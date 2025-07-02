package inventory

import (
	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"gorm.io/gorm"
)

type InventoryHandler struct {
	db              *gorm.DB
	gcsService      *gcs.GCService
	appwriteService *aw.AppwriteService
}

func NewInventoryHandler(db *gorm.DB, gcsService *gcs.GCService, appwriteService *aw.AppwriteService) *InventoryHandler {
	return &InventoryHandler{
		db:              db,
		gcsService:      gcsService,
		appwriteService: appwriteService,
	}
}
