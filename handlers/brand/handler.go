package brand

import (
	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"gorm.io/gorm"
)

type BrandHandler struct {
	db              *gorm.DB
	gcsService      *gcs.GCService
	appwriteService *aw.AppwriteService
}

func NewBrandHandler(db *gorm.DB, gcsService *gcs.GCService, appwriteService *aw.AppwriteService) *BrandHandler {
	return &BrandHandler{db: db, gcsService: gcsService, appwriteService: appwriteService}
}
