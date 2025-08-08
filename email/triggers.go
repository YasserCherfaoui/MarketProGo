package email

import (
	"fmt"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"gorm.io/gorm"
)

// EmailTriggerService handles automatic email triggers based on business events
type EmailTriggerService struct {
	emailService EmailService
	db           *gorm.DB
}

// NewEmailTriggerService creates a new email trigger service
func NewEmailTriggerService(emailService EmailService, db *gorm.DB) *EmailTriggerService {
	return &EmailTriggerService{
		emailService: emailService,
		db:           db,
	}
}

// TriggerPasswordReset sends a password reset email
func (t *EmailTriggerService) TriggerPasswordReset(userEmail, userName, resetToken string) error {
	data := map[string]interface{}{
		"UserName":     userName,
		"ResetLink":    fmt.Sprintf("%s/reset-password?token=%s", "https://algeriamarket.co.uk", resetToken),
		"ExpiryTime":   24, // 24 hours
		"UserEmail":    userEmail,
		"CompanyName":  "Algeria Market",
		"SiteURL":      "https://algeriamarket.co.uk",
		"SupportEmail": "enquirees@algeriamarket.co.uk",
	}

	recipient := models.EmailRecipient{
		Email: userEmail,
		Name:  userName,
	}

	return t.emailService.SendTransactionalEmail(models.EmailTypePasswordReset, data, recipient)
}

// TriggerWelcomeEmail sends a welcome email to new users
func (t *EmailTriggerService) TriggerWelcomeEmail(userEmail, userName string) error {
	data := map[string]interface{}{
		"UserName":       userName,
		"UserEmail":      userEmail,
		"CompanyName":    "Algeria Market",
		"SiteURL":        "https://algeriamarket.co.uk",
		"SupportEmail":   "enquirees@algeriamarket.co.uk",
		"ActivationLink": fmt.Sprintf("%s/activate?email=%s", "https://algeriamarket.co.uk", userEmail),
	}

	recipient := models.EmailRecipient{
		Email: userEmail,
		Name:  userName,
	}

	return t.emailService.SendTransactionalEmail(models.EmailTypeWelcome, data, recipient)
}

// TriggerOrderConfirmation sends an order confirmation email
func (t *EmailTriggerService) TriggerOrderConfirmation(orderID uint, userEmail, userName string, orderData map[string]interface{}) error {
	data := map[string]interface{}{
		"UserName":        userName,
		"UserEmail":       userEmail,
		"CompanyName":     "Algeria Market",
		"SiteURL":         "https://algeriamarket.co.uk",
		"SupportEmail":    "enquirees@algeriamarket.co.uk",
		"OrderNumber":     orderData["order_number"],
		"OrderDate":       orderData["order_date"],
		"TotalAmount":     orderData["total_amount"],
		"Currency":        orderData["currency"],
		"Items":           orderData["items"],
		"ShippingAddress": orderData["shipping_address"],
		"OrderStatusURL":  fmt.Sprintf("%s/orders/%d", "https://algeriamarket.co.uk", orderID),
	}

	recipient := models.EmailRecipient{
		Email: userEmail,
		Name:  userName,
	}

	return t.emailService.SendTransactionalEmail(models.EmailTypeOrderConfirmation, data, recipient)
}

// TriggerPaymentSuccess sends a payment success email
func (t *EmailTriggerService) TriggerPaymentSuccess(orderID uint, userEmail, userName string, paymentData map[string]interface{}) error {
	data := map[string]interface{}{
		"UserName":       userName,
		"UserEmail":      userEmail,
		"CompanyName":    "Algeria Market",
		"SiteURL":        "https://algeriamarket.co.uk",
		"SupportEmail":   "enquirees@algeriamarket.co.uk",
		"OrderNumber":    paymentData["order_number"],
		"OrderDate":      paymentData["order_date"],
		"TotalAmount":    paymentData["total_amount"],
		"Currency":       paymentData["currency"],
		"PaymentMethod":  paymentData["payment_method"],
		"OrderStatusURL": fmt.Sprintf("%s/orders/%d", "https://algeriamarket.co.uk", orderID),
	}

	recipient := models.EmailRecipient{
		Email: userEmail,
		Name:  userName,
	}

	return t.emailService.SendTransactionalEmail(models.EmailTypePaymentSuccess, data, recipient)
}

// TriggerPaymentFailed sends a payment failed email
func (t *EmailTriggerService) TriggerPaymentFailed(orderID uint, userEmail, userName string, paymentData map[string]interface{}) error {
	data := map[string]interface{}{
		"UserName":          userName,
		"UserEmail":         userEmail,
		"CompanyName":       "Algeria Market",
		"SiteURL":           "https://algeriamarket.co.uk",
		"SupportEmail":      "enquirees@algeriamarket.co.uk",
		"OrderNumber":       paymentData["order_number"],
		"OrderDate":         paymentData["order_date"],
		"TotalAmount":       paymentData["total_amount"],
		"Currency":          paymentData["currency"],
		"PaymentMethod":     paymentData["payment_method"],
		"ErrorMessage":      paymentData["error_message"],
		"RetryPaymentURL":   fmt.Sprintf("%s/orders/%d/payment/retry", "https://algeriamarket.co.uk", orderID),
		"UpdatePaymentURL":  fmt.Sprintf("%s/orders/%d/payment/update", "https://algeriamarket.co.uk", orderID),
		"ContactSupportURL": "https://algeriamarket.co.uk/support",
	}

	recipient := models.EmailRecipient{
		Email: userEmail,
		Name:  userName,
	}

	return t.emailService.SendTransactionalEmail(models.EmailTypePaymentFailed, data, recipient)
}

// TriggerOrderStatusUpdate sends an order status update email
func (t *EmailTriggerService) TriggerOrderStatusUpdate(orderID uint, userEmail, userName string, statusData map[string]interface{}) error {
	data := map[string]interface{}{
		"UserName":          userName,
		"UserEmail":         userEmail,
		"CompanyName":       "Algeria Market",
		"SiteURL":           "https://algeriamarket.co.uk",
		"SupportEmail":      "enquirees@algeriamarket.co.uk",
		"OrderNumber":       statusData["order_number"],
		"OrderDate":         statusData["order_date"],
		"Status":            statusData["status"],
		"StatusDisplay":     statusData["status_display"],
		"TotalAmount":       statusData["total_amount"],
		"Currency":          statusData["currency"],
		"TrackingNumber":    statusData["tracking_number"],
		"CarrierName":       statusData["carrier_name"],
		"TrackingURL":       statusData["tracking_url"],
		"EstimatedDelivery": statusData["estimated_delivery"],
		"Timeline":          statusData["timeline"],
		"OrderStatusURL":    fmt.Sprintf("%s/orders/%d", "https://algeriamarket.co.uk", orderID),
	}

	recipient := models.EmailRecipient{
		Email: userEmail,
		Name:  userName,
	}

	return t.emailService.SendTransactionalEmail(models.EmailTypeOrderStatusUpdate, data, recipient)
}

// TriggerSecurityAlert sends a security alert email
func (t *EmailTriggerService) TriggerSecurityAlert(userEmail, userName string, securityData map[string]interface{}) error {
	data := map[string]interface{}{
		"UserName":          userName,
		"UserEmail":         userEmail,
		"CompanyName":       "Algeria Market",
		"SiteURL":           "https://algeriamarket.co.uk",
		"SupportEmail":      "enquirees@algeriamarket.co.uk",
		"EventType":         securityData["event_type"],
		"EventDateTime":     securityData["event_datetime"],
		"Location":          securityData["location"],
		"Device":            securityData["device"],
		"IPAddress":         securityData["ip_address"],
		"SecureAccountURL":  "https://algeriamarket.co.uk/account/security",
		"ViewActivityURL":   "https://algeriamarket.co.uk/account/activity",
		"ResetPasswordURL":  "https://algeriamarket.co.uk/account/reset-password",
		"UnlockAccountURL":  "https://algeriamarket.co.uk/account/unlock",
		"ContactSupportURL": "https://algeriamarket.co.uk/support",
	}

	recipient := models.EmailRecipient{
		Email: userEmail,
		Name:  userName,
	}

	return t.emailService.SendTransactionalEmail(models.EmailTypeSecurityAlert, data, recipient)
}

// TriggerAdminNotification sends an admin notification email
func (t *EmailTriggerService) TriggerAdminNotification(adminEmail, adminName string, notificationData map[string]interface{}) error {
	data := map[string]interface{}{
		"AdminName":              adminName,
		"AdminEmail":             adminEmail,
		"CompanyName":            "Algeria Market",
		"SiteURL":                "https://algeriamarket.co.uk",
		"SupportEmail":           "enquirees@algeriamarket.co.uk",
		"NotificationType":       notificationData["notification_type"],
		"Priority":               notificationData["priority"],
		"DateTime":               notificationData["datetime"],
		"System":                 notificationData["system"],
		"ReferenceID":            notificationData["reference_id"],
		"OrderNumber":            notificationData["order_number"],
		"CustomerName":           notificationData["customer_name"],
		"TotalAmount":            notificationData["total_amount"],
		"Currency":               notificationData["currency"],
		"ItemCount":              notificationData["item_count"],
		"Amount":                 notificationData["amount"],
		"ErrorMessage":           notificationData["error_message"],
		"LowStockItems":          notificationData["low_stock_items"],
		"ErrorCode":              notificationData["error_code"],
		"Component":              notificationData["component"],
		"OrderManagementURL":     "https://algeriamarket.co.uk/admin/orders",
		"AdminDashboardURL":      "https://algeriamarket.co.uk/admin",
		"PaymentManagementURL":   "https://algeriamarket.co.uk/admin/payments",
		"CustomerSupportURL":     "https://algeriamarket.co.uk/admin/support",
		"InventoryManagementURL": "https://algeriamarket.co.uk/admin/inventory",
		"SystemLogsURL":          "https://algeriamarket.co.uk/admin/logs",
	}

	recipient := models.EmailRecipient{
		Email: adminEmail,
		Name:  adminName,
	}

	return t.emailService.SendTransactionalEmail(models.EmailTypeAdminNotification, data, recipient)
}

// TriggerNewOrderAdminNotification sends admin notification for new orders
func (t *EmailTriggerService) TriggerNewOrderAdminNotification(orderID uint, orderData map[string]interface{}) error {
	// Get admin users from database
	var adminUsers []models.User
	if err := t.db.Where("role = ?", "admin").Find(&adminUsers).Error; err != nil {
		return fmt.Errorf("failed to get admin users: %w", err)
	}

	for _, admin := range adminUsers {
		notificationData := map[string]interface{}{
			"notification_type": "new_order",
			"priority":          "medium",
			"datetime":          time.Now().Format("2006-01-02 15:04:05"),
			"system":            "order_management",
			"reference_id":      fmt.Sprintf("ORDER_%d", orderID),
			"order_number":      orderData["order_number"],
			"customer_name":     orderData["customer_name"],
			"total_amount":      orderData["total_amount"],
			"currency":          orderData["currency"],
			"item_count":        orderData["item_count"],
		}

		adminName := fmt.Sprintf("%s %s", admin.FirstName, admin.LastName)
		if err := t.TriggerAdminNotification(admin.Email, adminName, notificationData); err != nil {
			// Log error but continue with other admins
			fmt.Printf("Failed to send admin notification to %s: %v\n", admin.Email, err)
		}
	}

	return nil
}

// TriggerPaymentFailedAdminNotification sends admin notification for failed payments
func (t *EmailTriggerService) TriggerPaymentFailedAdminNotification(orderID uint, paymentData map[string]interface{}) error {
	// Get admin users from database
	var adminUsers []models.User
	if err := t.db.Where("role = ?", "admin").Find(&adminUsers).Error; err != nil {
		return fmt.Errorf("failed to get admin users: %w", err)
	}

	for _, admin := range adminUsers {
		notificationData := map[string]interface{}{
			"notification_type": "payment_failed",
			"priority":          "high",
			"datetime":          time.Now().Format("2006-01-02 15:04:05"),
			"system":            "payment_processing",
			"reference_id":      fmt.Sprintf("PAYMENT_%d", orderID),
			"order_number":      paymentData["order_number"],
			"customer_name":     paymentData["customer_name"],
			"amount":            paymentData["amount"],
			"currency":          paymentData["currency"],
			"error_message":     paymentData["error_message"],
		}

		adminName := fmt.Sprintf("%s %s", admin.FirstName, admin.LastName)
		if err := t.TriggerAdminNotification(admin.Email, adminName, notificationData); err != nil {
			// Log error but continue with other admins
			fmt.Printf("Failed to send admin notification to %s: %v\n", admin.Email, err)
		}
	}

	return nil
}

// Support notification helpers

// TriggerTicketResponse notifies user about a new response on their ticket
func (t *EmailTriggerService) TriggerTicketResponse(userEmail, userName string, data map[string]interface{}) error {
	recipient := models.EmailRecipient{Email: userEmail, Name: userName}
	return t.emailService.SendTransactionalEmail(models.EmailTypeTicketResponse, data, recipient)
}

// TriggerTicketStatusUpdated notifies user about ticket status change
func (t *EmailTriggerService) TriggerTicketStatusUpdated(userEmail, userName string, data map[string]interface{}) error {
	recipient := models.EmailRecipient{Email: userEmail, Name: userName}
	return t.emailService.SendTransactionalEmail(models.EmailTypeTicketStatusUpdated, data, recipient)
}

// TriggerDisputeResponse notifies user about a new response on their dispute
func (t *EmailTriggerService) TriggerDisputeResponse(userEmail, userName string, data map[string]interface{}) error {
	recipient := models.EmailRecipient{Email: userEmail, Name: userName}
	return t.emailService.SendTransactionalEmail(models.EmailTypeDisputeResponse, data, recipient)
}

// TriggerDisputeStatusUpdated notifies user about dispute status change
func (t *EmailTriggerService) TriggerDisputeStatusUpdated(userEmail, userName string, data map[string]interface{}) error {
	recipient := models.EmailRecipient{Email: userEmail, Name: userName}
	return t.emailService.SendTransactionalEmail(models.EmailTypeDisputeStatusUpdated, data, recipient)
}

// TriggerContactStatusUpdated notifies user about inquiry status change
func (t *EmailTriggerService) TriggerContactStatusUpdated(userEmail, userName string, data map[string]interface{}) error {
	recipient := models.EmailRecipient{Email: userEmail, Name: userName}
	return t.emailService.SendTransactionalEmail(models.EmailTypeContactStatusUpdated, data, recipient)
}

// TriggerAbuseStatusUpdated notifies reporter about abuse report status change
func (t *EmailTriggerService) TriggerAbuseStatusUpdated(userEmail, userName string, data map[string]interface{}) error {
	recipient := models.EmailRecipient{Email: userEmail, Name: userName}
	return t.emailService.SendTransactionalEmail(models.EmailTypeAbuseStatusUpdated, data, recipient)
}

// SendTemplateDirect renders and queues a specific template name with given recipient
func (t *EmailTriggerService) SendTemplateDirect(templateName string, data map[string]interface{}, recipient models.EmailRecipient, emailType models.EmailType) error {
	// Render and send via EmailService directly using transactional path
	// We temporarily set subject via data["subject"] in callers
	// Bypass type-to-template mapping by directly calling SendEmail with templateName
	return t.emailService.SendEmail(templateName, data, recipient)
}
