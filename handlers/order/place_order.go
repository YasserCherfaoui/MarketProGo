package order

import (
	"fmt"
	"strconv"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PlaceOrderRequest struct {
	ShippingAddressID uint    `json:"shipping_address_id" binding:"required"`
	PaymentMethod     string  `json:"payment_method" binding:"required"`
	CustomerNotes     string  `json:"customer_notes"`
	ShippingMethod    string  `json:"shipping_method"`
	TaxAmount         float64 `json:"tax_amount"`
	ShippingAmount    float64 `json:"shipping_amount"`
	DiscountAmount    float64 `json:"discount_amount"`
}

func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "order/place_order", "User not authenticated")
		return
	}
	uid := userID.(uint)

	var req PlaceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "order/place_order", err.Error())
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get user's cart with items
	var cart models.Cart
	if err := tx.Preload("Items.ProductVariant.Product").
		Preload("Items.ProductVariant.Product.Images").
		Preload("Items.ProductVariant.OptionValues").
		Preload("Items.Product"). // Legacy support
		Where("user_id = ?", uid).
		First(&cart).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			response.GenerateNotFoundResponse(c, "order/place_order", "Cart not found")
		} else {
			response.GenerateInternalServerErrorResponse(c, "order/place_order", "Failed to get cart")
		}
		return
	}

	// Check if cart has items
	if len(cart.Items) == 0 {
		tx.Rollback()
		response.GenerateBadRequestResponse(c, "order/place_order", "Cart is empty")
		return
	}

	// Verify shipping address belongs to user
	var address models.Address
	if err := tx.Where("id = ? AND user_id = ?", req.ShippingAddressID, uid).
		First(&address).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			response.GenerateNotFoundResponse(c, "order/place_order", "Shipping address not found")
		} else {
			response.GenerateInternalServerErrorResponse(c, "order/place_order", "Failed to verify shipping address")
		}
		return
	}

	// Calculate total amount
	var totalAmount float64
	for _, item := range cart.Items {
		// Fetch latest variant with price tiers
		var variant models.ProductVariant
		h.db.Model(&models.ProductVariant{}).Preload("PriceTiers").First(&variant, item.ProductVariantID)
		if item.Quantity < variant.MinQuantity {
			tx.Rollback()
			response.GenerateBadRequestResponse(c, "order/place_order", "Minimum quantity for variant '"+variant.Name+"' is "+strconv.Itoa(variant.MinQuantity))
			return
		}
		// Dynamic pricing: select price tier
		unitPrice := variant.BasePrice
		if len(variant.PriceTiers) > 0 {
			tiers := variant.PriceTiers
			for i := range tiers {
				for j := i + 1; j < len(tiers); j++ {
					if tiers[j].MinQuantity > tiers[i].MinQuantity {
						tiers[i], tiers[j] = tiers[j], tiers[i]
					}
				}
			}
			for _, tier := range tiers {
				if item.Quantity >= tier.MinQuantity {
					unitPrice = tier.Price
					break
				}
			}
		}
		item.UnitPrice = unitPrice
		item.TotalPrice = float64(item.Quantity) * item.UnitPrice
		totalAmount += item.TotalPrice
	}

	// Calculate final amount
	finalAmount := totalAmount + req.TaxAmount + req.ShippingAmount - req.DiscountAmount

	// Generate order number
	orderNumber := generateOrderNumber()

	// Create order
	order := models.Order{
		OrderNumber:       orderNumber,
		UserID:            uid,
		Status:            models.OrderStatusPending,
		PaymentStatus:     models.PaymentStatusPending,
		TotalAmount:       totalAmount,
		TaxAmount:         req.TaxAmount,
		ShippingAmount:    req.ShippingAmount,
		DiscountAmount:    req.DiscountAmount,
		FinalAmount:       finalAmount,
		ShippingAddressID: req.ShippingAddressID,
		ShippingMethod:    req.ShippingMethod,
		PaymentMethod:     req.PaymentMethod,
		CustomerNotes:     req.CustomerNotes,
		OrderDate:         time.Now(),
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "order/place_order", "Failed to create order")
		return
	}

	// Create order items from cart items
	var orderItems []models.OrderItem
	for _, cartItem := range cart.Items {
		orderItem := models.OrderItem{
			OrderID:          order.ID,
			ProductVariantID: cartItem.ProductVariantID,
			ProductID:        cartItem.ProductID, // Legacy support
			Quantity:         cartItem.Quantity,
			UnitPrice:        cartItem.UnitPrice,
			TotalAmount:      cartItem.TotalPrice,
			Status:           "active",
		}
		orderItems = append(orderItems, orderItem)
	}

	if err := tx.Create(&orderItems).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "order/place_order", "Failed to create order items")
		return
	}

	// Clear cart items
	if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "order/place_order", "Failed to clear cart")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "order/place_order", "Failed to commit transaction")
		return
	}

	// Load the complete order with relationships for response
	var completeOrder models.Order
	if err := h.db.Preload("User").
		Preload("ShippingAddress").
		Preload("Items.ProductVariant.Product").
		Preload("Items.ProductVariant.Product.Images").
		Preload("Items.ProductVariant.OptionValues").
		Preload("Items.Product"). // Legacy support
		First(&completeOrder, order.ID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "order/place_order", "Order created but failed to load details")
		return
	}

	// Send order confirmation email asynchronously
	go func() {
		// Prepare order data for email
		orderData := map[string]interface{}{
			"order_number":     completeOrder.OrderNumber,
			"order_date":       completeOrder.OrderDate,
			"total_amount":     completeOrder.FinalAmount,
			"currency":         "GBP",
			"items":            completeOrder.Items,
			"shipping_address": completeOrder.ShippingAddress,
		}

		// Send order confirmation to customer
		if err := h.emailTriggerSvc.TriggerOrderConfirmation(
			completeOrder.ID,
			completeOrder.User.Email,
			fmt.Sprintf("%s %s", completeOrder.User.FirstName, completeOrder.User.LastName),
			orderData,
		); err != nil {
			fmt.Printf("Failed to send order confirmation email: %v\n", err)
		}

		// Send admin notification
		if err := h.emailTriggerSvc.TriggerNewOrderAdminNotification(completeOrder.ID, orderData); err != nil {
			fmt.Printf("Failed to send admin notification: %v\n", err)
		}
	}()

	response.GenerateCreatedResponse(c, "Order placed successfully", completeOrder)
}

func generateOrderNumber() string {
	// Generate order number with timestamp
	now := time.Now()
	return fmt.Sprintf("ORD-%d%02d%02d-%d",
		now.Year(), now.Month(), now.Day(),
		now.Unix()%10000) // Last 4 digits of timestamp for uniqueness
}
