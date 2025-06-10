package cart

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type AddCartItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
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

	var cart models.Cart
	h.db.Where("user_id = ?", uid).FirstOrCreate(&cart, models.Cart{UserID: &uid})

	var item models.CartItem
	err := h.db.Where("cart_id = ? AND product_id = ?", cart.ID, req.ProductID).First(&item).Error
	if err == nil {
		item.Quantity += req.Quantity
		h.db.Save(&item)
	} else {
		item = models.CartItem{
			CartID:    cart.ID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		}
		h.db.Create(&item)
	}

	response.GenerateSuccessResponse(c, "cart/add_item", item)
}
