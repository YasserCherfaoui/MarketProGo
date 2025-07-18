package review

import (
	"fmt"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"gorm.io/gorm"
)

// PurchaseVerificationResult represents the result of a purchase verification
type PurchaseVerificationResult struct {
	IsVerified     bool                   `json:"is_verified"`
	OrderItem      *models.OrderItem      `json:"order_item,omitempty"`
	Order          *models.Order          `json:"order,omitempty"`
	ProductVariant *models.ProductVariant `json:"product_variant,omitempty"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
}

// VerifyPurchase checks if a user has purchased a specific product variant
// and the order has been delivered successfully
func (h *ReviewHandler) VerifyPurchase(userID uint, productVariantID uint) (*PurchaseVerificationResult, error) {
	var orderItems []models.OrderItem
	// Query all delivered, paid, active order items for the user and product variant, ordered by delivered date descending
	err := h.db.
		Preload("Order").
		Preload("ProductVariant").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where(`
			orders.user_id = ? 
			AND order_items.product_variant_id = ? 
			AND orders.status = ? 
			AND orders.payment_status = ?
			AND order_items.status = ?
		`, userID, productVariantID, models.OrderStatusDelivered, models.PaymentStatusPaid, "active").
		Order("orders.delivered_date DESC").
		Find(&orderItems).Error

	if err != nil {
		return nil, fmt.Errorf("failed to verify purchase: %w", err)
	}

	twoYearsAgo := time.Now().AddDate(-2, 0, 0)
	for _, item := range orderItems {
		if item.Order.DeliveredDate != nil && !item.Order.DeliveredDate.Before(twoYearsAgo) {
			return &PurchaseVerificationResult{
				IsVerified:     true,
				OrderItem:      &item,
				Order:          &item.Order,
				ProductVariant: &item.ProductVariant,
			}, nil
		}
	}

	if len(orderItems) > 0 {
		// All purchases are too old
		return &PurchaseVerificationResult{
			IsVerified:   false,
			ErrorMessage: "Purchase is too old to review (more than 2 years)",
		}, nil
	}

	return &PurchaseVerificationResult{
		IsVerified:   false,
		ErrorMessage: "No verified purchase found for this product",
	}, nil
}

// VerifyPurchaseWithOrderItemID checks if a specific order item belongs to the user
// and can be used for review verification
func (h *ReviewHandler) VerifyPurchaseWithOrderItemID(userID uint, orderItemID uint) (*PurchaseVerificationResult, error) {
	var orderItem models.OrderItem

	err := h.db.
		Preload("Order").
		Preload("ProductVariant").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where(`
			orders.user_id = ? 
			AND order_items.id = ? 
			AND orders.status = ? 
			AND orders.payment_status = ?
			AND order_items.status = ?
		`, userID, orderItemID, models.OrderStatusDelivered, models.PaymentStatusPaid, "active").
		First(&orderItem).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &PurchaseVerificationResult{
				IsVerified:   false,
				ErrorMessage: "Order item not found or not eligible for review",
			}, nil
		}
		return nil, fmt.Errorf("failed to verify purchase with order item ID: %w", err)
	}

	// Check if the order was delivered within a reasonable timeframe
	twoYearsAgo := time.Now().AddDate(-2, 0, 0)
	if orderItem.Order.DeliveredDate != nil && orderItem.Order.DeliveredDate.Before(twoYearsAgo) {
		return &PurchaseVerificationResult{
			IsVerified:   false,
			ErrorMessage: "Purchase is too old to review (more than 2 years)",
		}, nil
	}

	return &PurchaseVerificationResult{
		IsVerified:     true,
		OrderItem:      &orderItem,
		Order:          &orderItem.Order,
		ProductVariant: &orderItem.ProductVariant,
	}, nil
}

// GetUserPurchasedProducts returns a list of product variants that the user has purchased
// and can potentially review
func (h *ReviewHandler) GetUserPurchasedProducts(userID uint, limit int) ([]models.OrderItem, error) {
	var orderItems []models.OrderItem

	query := h.db.
		Preload("Order").
		Preload("ProductVariant").
		Preload("ProductVariant.Product").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where(`
			orders.user_id = ? 
			AND orders.status = ? 
			AND orders.payment_status = ?
			AND order_items.status = ?
		`, userID, models.OrderStatusDelivered, models.PaymentStatusPaid, "active").
		Order("orders.delivered_date DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&orderItems).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user purchased products: %w", err)
	}

	return orderItems, nil
}

// CheckIfUserCanReview checks if a user can review a specific product variant
// This includes checking if they have purchased it and haven't already reviewed it
func (h *ReviewHandler) CheckIfUserCanReview(userID uint, productVariantID uint) (*PurchaseVerificationResult, error) {
	// First, verify the purchase
	purchaseResult, err := h.VerifyPurchase(userID, productVariantID)
	if err != nil {
		return nil, err
	}

	if !purchaseResult.IsVerified {
		return purchaseResult, nil
	}

	// Check if user has already reviewed this product variant
	var existingReview models.ProductReview
	err = h.db.Where("user_id = ? AND product_variant_id = ?", userID, productVariantID).
		First(&existingReview).Error

	if err == nil {
		// Review already exists
		return &PurchaseVerificationResult{
			IsVerified:   false,
			ErrorMessage: "You have already reviewed this product",
		}, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing review: %w", err)
	}

	// User can review this product
	return purchaseResult, nil
}

// GetReviewableProductsForUser returns a list of products the user can review
// (purchased but not yet reviewed)
func (h *ReviewHandler) GetReviewableProductsForUser(userID uint, limit int) ([]models.OrderItem, error) {
	// Get all purchased products
	purchasedItems, err := h.GetUserPurchasedProducts(userID, 0) // No limit to get all
	if err != nil {
		return nil, err
	}

	// Filter out products that have already been reviewed
	var reviewableItems []models.OrderItem
	for _, item := range purchasedItems {
		var existingReview models.ProductReview
		err := h.db.Where("user_id = ? AND product_variant_id = ?", userID, item.ProductVariantID).
			First(&existingReview).Error

		if err == gorm.ErrRecordNotFound {
			// No review exists, this product is reviewable
			reviewableItems = append(reviewableItems, item)
		} else if err != nil {
			return nil, fmt.Errorf("failed to check existing review: %w", err)
		}
		// If review exists, skip this product
	}

	// Apply limit if specified
	if limit > 0 && len(reviewableItems) > limit {
		reviewableItems = reviewableItems[:limit]
	}

	return reviewableItems, nil
}
