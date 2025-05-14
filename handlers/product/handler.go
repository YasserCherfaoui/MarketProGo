package product

import (
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"gorm.io/gorm"
)

type ProductHandler struct {
	db         *gorm.DB
	gcsService *gcs.GCService
}

func NewProductHandler(db *gorm.DB, gcsService *gcs.GCService) *ProductHandler {
	return &ProductHandler{db: db, gcsService: gcsService}
}
