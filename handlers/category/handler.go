package category

import (
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"gorm.io/gorm"
)

type CategoryHandler struct {
	db         *gorm.DB
	gcsService *gcs.GCService
}

func NewCategoryHandler(db *gorm.DB, gcsService *gcs.GCService) *CategoryHandler {
	return &CategoryHandler{db: db, gcsService: gcsService}
}
