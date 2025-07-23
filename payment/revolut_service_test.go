package payment

import (
	"encoding/json"
	"testing"

	"github.com/YasserCherfaoui/MarketProGo/cfg"
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/payment/revolut"
	"github.com/stretchr/testify/assert"
)

func TestRevolutPaymentService_MapRevolutStatusToPaymentStatus(t *testing.T) {
	// Setup
	config := &cfg.RevolutConfig{
		APIKey:    "test_api_key",
		BaseURL:   "https://sandbox-merchant.revolut.com",
		IsSandbox: true,
	}

	service := &RevolutPaymentService{
		config: config,
	}

	// Test cases
	testCases := []struct {
		revolutState string
		expected     models.RevolutPaymentStatus
	}{
		{"pending", models.RevolutPaymentStatusPending},
		{"authorized", models.RevolutPaymentStatusAuthorized},
		{"completed", models.RevolutPaymentStatusCompleted},
		{"failed", models.RevolutPaymentStatusFailed},
		{"cancelled", models.RevolutPaymentStatusCancelled},
		{"refunded", models.RevolutPaymentStatusRefunded},
		{"unknown", models.RevolutPaymentStatusPending}, // default case
	}

	for _, tc := range testCases {
		t.Run(tc.revolutState, func(t *testing.T) {
			result := service.mapRevolutStatusToPaymentStatus(tc.revolutState)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRevolutPaymentService_ValidateWebhookSignature(t *testing.T) {
	// Setup
	config := &cfg.RevolutConfig{
		APIKey:        "test_api_key",
		BaseURL:       "https://sandbox-merchant.revolut.com",
		IsSandbox:     true,
		WebhookSecret: "test_webhook_secret",
	}

	service := &RevolutPaymentService{
		config: config,
	}

	// Test data
	payload := []byte(`{"test": "data"}`)

	// Test with empty webhook secret (should return true when webhook secret is not configured)
	service.webhookSecret = ""
	result := service.validateWebhookSignature(payload, "")
	assert.True(t, result)

	// Test with webhook secret configured but invalid signature
	service.webhookSecret = "test_secret"
	result = service.validateWebhookSignature(payload, "invalid_signature")
	assert.False(t, result)

	// Test with valid signature (this would require calculating the actual HMAC)
	// For now, we just test that it returns false for invalid signatures
	service.webhookSecret = "test_secret"
	result = service.validateWebhookSignature(payload, "valid_signature_that_doesnt_match")
	assert.False(t, result)
}

func TestRevolutPaymentService_WebhookSignatureValidation(t *testing.T) {
	// Setup
	config := &cfg.RevolutConfig{
		APIKey:        "test_api_key",
		BaseURL:       "https://sandbox-merchant.revolut.com",
		IsSandbox:     true,
		WebhookSecret: "test_webhook_secret",
	}

	service := &RevolutPaymentService{
		config:        config,
		webhookSecret: config.WebhookSecret, // Set the webhook secret directly
	}

	// Test data
	payload := []byte(`{"event":"ORDER_COMPLETED","order_id":"test_order_123"}`)

	// Test with valid signature format but wrong signature
	validSignature := "v1=09a9989dd8d9282c1d34974fc730f5cbfc4f4296941247e90ae5256590a11e8c"
	result := service.validateWebhookSignature(payload, validSignature)
	assert.False(t, result) // Should be false because signature doesn't match

	// Test with invalid signature format
	invalidSignature := "invalid_signature_format"
	result = service.validateWebhookSignature(payload, invalidSignature)
	assert.False(t, result)

	// Test with empty signature
	result = service.validateWebhookSignature(payload, "")
	assert.False(t, result)

	// Test with missing v1= prefix
	noPrefixSignature := "09a9989dd8d9282c1d34974fc730f5cbfc4f4296941247e90ae5256590a11e8c"
	result = service.validateWebhookSignature(payload, noPrefixSignature)
	assert.False(t, result)

	// Test with empty webhook secret (should return true)
	service.webhookSecret = ""
	result = service.validateWebhookSignature(payload, validSignature)
	assert.True(t, result)
}

func TestRevolutOrderRequest_JSONStructure(t *testing.T) {
	// Test the JSON structure of a minimal order request
	req := &revolut.OrderRequest{
		Amount:      100, // £1.00
		Currency:    "GBP",
		Description: "Test order",
		Customer: &revolut.Customer{
			ID:       "1",
			FullName: "Test User",
			Email:    "test@example.com",
			Phone:    "+1234567890",
		},
		CaptureMode:      "automatic",
		EnforceChallenge: "automatic",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(req)
	assert.NoError(t, err)

	// Log the JSON for inspection
	t.Logf("Order request JSON: %s", string(jsonData))

	// Verify required fields are present
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)

	// Check required fields
	assert.Contains(t, jsonMap, "amount")
	assert.Contains(t, jsonMap, "currency")
	assert.Contains(t, jsonMap, "description")
	assert.Contains(t, jsonMap, "customer")
	assert.Contains(t, jsonMap, "capture_mode")
	assert.Contains(t, jsonMap, "enforce_challenge")

	// Check customer fields
	customer, ok := jsonMap["customer"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, customer, "id")
	assert.Contains(t, customer, "full_name")
	assert.Contains(t, customer, "email")

	// Verify values
	assert.Equal(t, float64(100), jsonMap["amount"])
	assert.Equal(t, "GBP", jsonMap["currency"])
	assert.Equal(t, "Test order", jsonMap["description"])
	assert.Equal(t, "automatic", jsonMap["capture_mode"])
	assert.Equal(t, "automatic", jsonMap["enforce_challenge"])
}
