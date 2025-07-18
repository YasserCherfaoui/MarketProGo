package product

import (
	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct {
	db              *gorm.DB
	gcsService      *gcs.GCService
	appwriteService *aw.AppwriteService
	reviewService   *ReviewIntegrationService
}

func NewProductHandler(db *gorm.DB, gcsService *gcs.GCService, appwriteService *aw.AppwriteService) *ProductHandler {
	return &ProductHandler{
		db:              db,
		gcsService:      gcsService,
		appwriteService: appwriteService,
		reviewService:   NewReviewIntegrationService(db),
	}
}

// GetProductReviewStats handles the request for product review statistics
func (h *ProductHandler) GetProductReviewStats(c *gin.Context) {
	h.reviewService.GetProductReviewStatsHandler(c)
}
