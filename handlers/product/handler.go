package product

import (
	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"gorm.io/gorm"
)

type ProductHandler struct {
	db              *gorm.DB
	gcsService      *gcs.GCService
	appwriteService *aw.AppwriteService
}

func NewProductHandler(db *gorm.DB, gcsService *gcs.GCService, appwriteService *aw.AppwriteService) *ProductHandler {
	return &ProductHandler{db: db, gcsService: gcsService, appwriteService: appwriteService}
}
