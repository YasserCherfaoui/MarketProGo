package category

import (
	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"gorm.io/gorm"
)

type CategoryHandler struct {
	db              *gorm.DB
	gcsService      *gcs.GCService
	appwriteService *aw.AppwriteService
}

func NewCategoryHandler(db *gorm.DB, gcsService *gcs.GCService, appwriteService *aw.AppwriteService) *CategoryHandler {
	return &CategoryHandler{db: db, gcsService: gcsService, appwriteService: appwriteService}
}
