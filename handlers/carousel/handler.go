package carousel

import (
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"gorm.io/gorm"
)

type CarouselHandler struct {
	db         *gorm.DB
	gcsService *gcs.GCService
}

func NewCarouselHandler(db *gorm.DB, gcsService *gcs.GCService) *CarouselHandler {
	return &CarouselHandler{db: db, gcsService: gcsService}
}
