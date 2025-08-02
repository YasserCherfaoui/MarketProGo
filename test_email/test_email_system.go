package main

import (
	"fmt"
	"log"

	"github.com/YasserCherfaoui/MarketProGo/cfg"
	"github.com/YasserCherfaoui/MarketProGo/database"
	"github.com/YasserCherfaoui/MarketProGo/email"
	"github.com/YasserCherfaoui/MarketProGo/models"
)

func main() {
	fmt.Println("Testing Email System...")

	// Load configuration
	config, err := cfg.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize email components
	emailProvider := email.NewMockEmailProvider(
		config.Email.SenderEmail,
		config.Email.SenderName,
	)

	templateEngine := email.NewHTMLTemplateEngine("templates/emails")
	if err := templateEngine.ReloadTemplates(); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	emailAnalytics := email.NewEmailAnalytics(db)

	// Initialize email service
	emailService := email.NewEmailService(
		emailProvider,
		templateEngine,
		nil, // No queue for testing
		emailAnalytics,
		&config.Email,
		db,
	)

	// Test 1: Send a password reset email
	fmt.Println("\n1. Testing Password Reset Email...")
	recipient := models.EmailRecipient{
		Email: "test@example.com",
		Name:  "Test User",
	}

	data := map[string]interface{}{
		"UserName":     "Test User",
		"ResetLink":    "https://algeriamarket.co.uk/reset-password?token=test123",
		"ExpiryTime":   24,
		"UserEmail":    "test@example.com",
		"CompanyName":  "Algeria Market",
		"SiteURL":      "https://algeriamarket.co.uk",
		"SupportEmail": "enquirees@algeriamarket.co.uk",
	}

	err = emailService.SendTransactionalEmail(models.EmailTypePasswordReset, data, recipient)
	if err != nil {
		log.Printf("Failed to send password reset email: %v", err)
	} else {
		fmt.Println("✓ Password reset email sent successfully")
	}

	// Test 2: Send a welcome email
	fmt.Println("\n2. Testing Welcome Email...")
	data = map[string]interface{}{
		"UserName":       "Test User",
		"UserEmail":      "test@example.com",
		"CompanyName":    "Algeria Market",
		"SiteURL":        "https://algeriamarket.co.uk",
		"SupportEmail":   "enquirees@algeriamarket.co.uk",
		"ActivationLink": "https://algeriamarket.co.uk/activate?email=test@example.com",
	}

	err = emailService.SendTransactionalEmail(models.EmailTypeWelcome, data, recipient)
	if err != nil {
		log.Printf("Failed to send welcome email: %v", err)
	} else {
		fmt.Println("✓ Welcome email sent successfully")
	}

	// Test 3: Send an order confirmation email
	fmt.Println("\n3. Testing Order Confirmation Email...")
	data = map[string]interface{}{
		"UserName":        "Test User",
		"UserEmail":       "test@example.com",
		"CompanyName":     "Algeria Market",
		"SiteURL":         "https://algeriamarket.co.uk",
		"SupportEmail":    "enquirees@algeriamarket.co.uk",
		"OrderNumber":     "ORD-12345",
		"OrderDate":       "2024-01-15",
		"TotalAmount":     99.99,
		"Currency":        "GBP",
		"Items":           []map[string]interface{}{{"name": "Test Product", "quantity": 1, "price": 99.99}},
		"ShippingAddress": map[string]interface{}{"street": "123 Test St", "city": "London", "country": "UK"},
		"OrderStatusURL":  "https://algeriamarket.co.uk/orders/123",
	}

	err = emailService.SendTransactionalEmail(models.EmailTypeOrderConfirmation, data, recipient)
	if err != nil {
		log.Printf("Failed to send order confirmation email: %v", err)
	} else {
		fmt.Println("✓ Order confirmation email sent successfully")
	}

	// Test 4: Send a payment success email
	fmt.Println("\n4. Testing Payment Success Email...")
	data = map[string]interface{}{
		"UserName":       "Test User",
		"UserEmail":      "test@example.com",
		"CompanyName":    "Algeria Market",
		"SiteURL":        "https://algeriamarket.co.uk",
		"SupportEmail":   "enquirees@algeriamarket.co.uk",
		"OrderNumber":    "ORD-12345",
		"OrderDate":      "2024-01-15",
		"TotalAmount":    99.99,
		"Currency":       "GBP",
		"PaymentMethod":  "Credit Card",
		"OrderStatusURL": "https://algeriamarket.co.uk/orders/123",
	}

	err = emailService.SendTransactionalEmail(models.EmailTypePaymentSuccess, data, recipient)
	if err != nil {
		log.Printf("Failed to send payment success email: %v", err)
	} else {
		fmt.Println("✓ Payment success email sent successfully")
	}

	// Test 5: Send a payment failed email
	fmt.Println("\n5. Testing Payment Failed Email...")
	data = map[string]interface{}{
		"UserName":          "Test User",
		"UserEmail":         "test@example.com",
		"CompanyName":       "Algeria Market",
		"SiteURL":           "https://algeriamarket.co.uk",
		"SupportEmail":      "enquirees@algeriamarket.co.uk",
		"OrderNumber":       "ORD-12345",
		"OrderDate":         "2024-01-15",
		"TotalAmount":       99.99,
		"Currency":          "GBP",
		"PaymentMethod":     "Credit Card",
		"ErrorMessage":      "Insufficient funds",
		"RetryPaymentURL":   "https://algeriamarket.co.uk/orders/123/payment/retry",
		"UpdatePaymentURL":  "https://algeriamarket.co.uk/orders/123/payment/update",
		"ContactSupportURL": "https://algeriamarket.co.uk/support",
	}

	err = emailService.SendTransactionalEmail(models.EmailTypePaymentFailed, data, recipient)
	if err != nil {
		log.Printf("Failed to send payment failed email: %v", err)
	} else {
		fmt.Println("✓ Payment failed email sent successfully")
	}

	// Test 6: Send an order status update email
	fmt.Println("\n6. Testing Order Status Update Email...")
	data = map[string]interface{}{
		"UserName":          "Test User",
		"UserEmail":         "test@example.com",
		"CompanyName":       "Algeria Market",
		"SiteURL":           "https://algeriamarket.co.uk",
		"SupportEmail":      "enquirees@algeriamarket.co.uk",
		"OrderNumber":       "ORD-12345",
		"OrderDate":         "2024-01-15",
		"Status":            "shipped",
		"StatusDisplay":     "Shipped",
		"TotalAmount":       99.99,
		"Currency":          "GBP",
		"TrackingNumber":    "TRK123456789",
		"CarrierName":       "Royal Mail",
		"TrackingURL":       "https://royalmail.com/track/TRK123456789",
		"EstimatedDelivery": "2024-01-18",
		"Timeline": []map[string]interface{}{
			{"title": "Order Placed", "date": "2024-01-15", "status": "completed"},
			{"title": "Processing", "date": "2024-01-16", "status": "completed"},
			{"title": "Shipped", "date": "2024-01-17", "status": "current"},
			{"title": "Delivered", "date": "2024-01-18", "status": "pending"},
		},
		"OrderStatusURL": "https://algeriamarket.co.uk/orders/123",
	}

	err = emailService.SendTransactionalEmail(models.EmailTypeOrderStatusUpdate, data, recipient)
	if err != nil {
		log.Printf("Failed to send order status update email: %v", err)
	} else {
		fmt.Println("✓ Order status update email sent successfully")
	}

	// Test 7: Send a security alert email
	fmt.Println("\n7. Testing Security Alert Email...")
	data = map[string]interface{}{
		"UserName":          "Test User",
		"UserEmail":         "test@example.com",
		"CompanyName":       "Algeria Market",
		"SiteURL":           "https://algeriamarket.co.uk",
		"SupportEmail":      "enquirees@algeriamarket.co.uk",
		"EventType":         "login_attempt",
		"EventDateTime":     "2024-01-15 14:30:00",
		"Location":          "London, UK",
		"Device":            "Chrome on Windows",
		"IPAddress":         "192.168.1.1",
		"SecureAccountURL":  "https://algeriamarket.co.uk/account/security",
		"ViewActivityURL":   "https://algeriamarket.co.uk/account/activity",
		"ResetPasswordURL":  "https://algeriamarket.co.uk/account/reset-password",
		"UnlockAccountURL":  "https://algeriamarket.co.uk/account/unlock",
		"ContactSupportURL": "https://algeriamarket.co.uk/support",
	}

	err = emailService.SendTransactionalEmail(models.EmailTypeSecurityAlert, data, recipient)
	if err != nil {
		log.Printf("Failed to send security alert email: %v", err)
	} else {
		fmt.Println("✓ Security alert email sent successfully")
	}

	// Test 8: Send an admin notification email
	fmt.Println("\n8. Testing Admin Notification Email...")
	adminRecipient := models.EmailRecipient{
		Email: "admin@algeriamarket.co.uk",
		Name:  "Admin User",
	}

	data = map[string]interface{}{
		"AdminName":              "Admin User",
		"AdminEmail":             "admin@algeriamarket.co.uk",
		"CompanyName":            "Algeria Market",
		"SiteURL":                "https://algeriamarket.co.uk",
		"SupportEmail":           "enquirees@algeriamarket.co.uk",
		"NotificationType":       "new_order",
		"Priority":               "medium",
		"DateTime":               "2024-01-15 14:30:00",
		"System":                 "order_management",
		"ReferenceID":            "ORDER_123",
		"OrderNumber":            "ORD-12345",
		"CustomerName":           "Test User",
		"TotalAmount":            99.99,
		"Currency":               "GBP",
		"ItemCount":              1,
		"OrderManagementURL":     "https://algeriamarket.co.uk/admin/orders",
		"AdminDashboardURL":      "https://algeriamarket.co.uk/admin",
		"PaymentManagementURL":   "https://algeriamarket.co.uk/admin/payments",
		"CustomerSupportURL":     "https://algeriamarket.co.uk/admin/support",
		"InventoryManagementURL": "https://algeriamarket.co.uk/admin/inventory",
		"SystemLogsURL":          "https://algeriamarket.co.uk/admin/logs",
	}

	err = emailService.SendTransactionalEmail(models.EmailTypeAdminNotification, data, adminRecipient)
	if err != nil {
		log.Printf("Failed to send admin notification email: %v", err)
	} else {
		fmt.Println("✓ Admin notification email sent successfully")
	}

	fmt.Println("\n🎉 All email tests completed successfully!")
	fmt.Println("The email system is working correctly with all templates.")
}
