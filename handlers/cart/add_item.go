package cart

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type AddCartItemRequest struct {
	ProductVariantID uint   `json:"product_variant_id" binding:"required"`
	Quantity         int    `json:"quantity" binding:"required,min=1"`
	PriceType        string `json:"price_type"` // "customer" or "b2b", defaults to "customer"
}

func (h *CartHandler) AddItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "cart/add_item", "Unauthorized")
		return
	}
	uid := userID.(uint)

	var req AddCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "cart/add_item", err.Error())
		return
	}

	// Default to customer pricing if not specified
	if req.PriceType == "" {
		req.PriceType = "customer"
	}

	// Validate price type
	if req.PriceType != "customer" && req.PriceType != "b2b" {
		response.GenerateBadRequestResponse(c, "cart/add_item", "Invalid price_type. Must be 'customer' or 'b2b'")
		return
	}

	// Get the product variant to validate and get pricing
	var variant models.ProductVariant
	if err := h.db.First(&variant, req.ProductVariantID).Error; err != nil {
		response.GenerateBadRequestResponse(c, "cart/add_item", "Product variant not found")
		return
	}

	// Check if variant is active
	if !variant.IsActive {
		response.GenerateBadRequestResponse(c, "cart/add_item", "Product variant is not available")
		return
	}

	// Calculate unit price based on price type
	var unitPrice float64
	if req.PriceType == "b2b" {
		unitPrice = variant.B2BPrice
	} else {
		unitPrice = variant.BasePrice
	}

	// Get or create cart
	var cart models.Cart
	h.db.Where("user_id = ?", uid).FirstOrCreate(&cart, models.Cart{UserID: &uid})

	// Check if item already exists in cart
	var item models.CartItem
	err := h.db.Where("cart_id = ? AND product_variant_id = ? AND price_type = ?", cart.ID, req.ProductVariantID, req.PriceType).First(&item).Error
	if err == nil {
		// Update existing item
		item.Quantity += req.Quantity
		item.TotalPrice = float64(item.Quantity) * item.UnitPrice
		h.db.Save(&item)
	} else {
		// Create new item
		totalPrice := float64(req.Quantity) * unitPrice
		item = models.CartItem{
			CartID:           cart.ID,
			ProductVariantID: req.ProductVariantID,
			Quantity:         req.Quantity,
			PriceType:        req.PriceType,
			UnitPrice:        unitPrice,
			TotalPrice:       totalPrice,
		}
		h.db.Create(&item)
	}

	// Preload variant and product data for response
	h.db.Preload("ProductVariant.Product").Preload("ProductVariant.Images").First(&item, item.ID)

	response.GenerateSuccessResponse(c, "cart/add_item", item)
}
