package email

import (
	"testing"

	"github.com/YasserCherfaoui/MarketProGo/cfg"
	"github.com/YasserCherfaoui/MarketProGo/models"
)

func TestNewOutlookProvider(t *testing.T) {
	// Test with missing configuration
	config := &cfg.OutlookConfig{}
	_, err := NewOutlookProvider(config)
	if err == nil {
		t.Error("Expected error when configuration is missing")
	}

	// Test with partial configuration
	config = &cfg.OutlookConfig{
		TenantID: "test-tenant",
		// Missing ClientID and ClientSecret
	}
	_, err = NewOutlookProvider(config)
	if err == nil {
		t.Error("Expected error when configuration is incomplete")
	}

	// Skip the actual provider creation test since it requires valid Azure credentials
	// In a real environment, this would be tested with actual credentials
	t.Skip("Skipping actual provider creation test - requires valid Azure credentials")
}

func TestMockEmailProvider(t *testing.T) {
	provider := NewMockEmailProvider("test@example.com", "Test Sender")

	email := &models.Email{
		Subject:     "Test Subject",
		HTMLContent: "<h1>Test Email</h1>",
		Recipients: []models.EmailRecipient{
			{
				Email: "recipient@example.com",
				Name:  "Test Recipient",
			},
		},
	}

	err := provider.SendEmail(email)
	if err != nil {
		t.Errorf("Mock provider should not return error: %v", err)
	}

	// Test bulk email
	emails := []*models.Email{email, email}
	err = provider.SendBulkEmail(emails)
	if err != nil {
		t.Errorf("Mock provider should not return error for bulk email: %v", err)
	}

	// Test delivery status
	status, err := provider.GetDeliveryStatus("test-id")
	if err != nil {
		t.Errorf("Mock provider should not return error for delivery status: %v", err)
	}
	if status != DeliveryStatusDelivered {
		t.Errorf("Expected DeliveryStatusDelivered, got %s", status)
	}

	// Test bounce list
	bounces, err := provider.GetBounceList()
	if err != nil {
		t.Errorf("Mock provider should not return error for bounce list: %v", err)
	}
	if len(bounces) != 0 {
		t.Errorf("Expected empty bounce list, got %d items", len(bounces))
	}

	// Test complaint list
	complaints, err := provider.GetComplaintList()
	if err != nil {
		t.Errorf("Mock provider should not return error for complaint list: %v", err)
	}
	if len(complaints) != 0 {
		t.Errorf("Expected empty complaint list, got %d items", len(complaints))
	}
}

func TestOutlookProviderValidation(t *testing.T) {
	provider := &OutlookProvider{
		senderEmail: "test@example.com",
		senderName:  "Test Sender",
	}

	// Test with no recipients
	email := &models.Email{
		Subject:     "Test Subject",
		HTMLContent: "<h1>Test Email</h1>",
		Recipients:  []models.EmailRecipient{},
	}

	err := provider.SendEmail(email)
	if err == nil {
		t.Error("Expected error when no recipients specified")
	}
}

// TestEmailProviderInterface tests that both providers implement the EmailProvider interface
func TestEmailProviderInterface(t *testing.T) {
	// Test that MockEmailProvider implements EmailProvider
	var _ EmailProvider = NewMockEmailProvider("test@example.com", "Test Sender")

	// Test that OutlookProvider implements EmailProvider by checking the struct directly
	// We can't create a real instance due to Azure credential requirements
	var _ EmailProvider = (*OutlookProvider)(nil)
}
