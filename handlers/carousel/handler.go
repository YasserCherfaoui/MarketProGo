package carousel

import (
	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"gorm.io/gorm"
)

type CarouselHandler struct {
	db              *gorm.DB
	gcsService      *gcs.GCService
	appwriteService *aw.AppwriteService
}

func NewCarouselHandler(db *gorm.DB, gcsService *gcs.GCService, appwriteService *aw.AppwriteService) *CarouselHandler {
	return &CarouselHandler{db: db, gcsService: gcsService, appwriteService: appwriteService}
}
