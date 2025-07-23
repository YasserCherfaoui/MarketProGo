package payment

import (
	"encoding/json"
	"fmt"
	"testing"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

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

func TestRevolutPaymentService_WebhookSignatureValidation(t *testing.T) {
	config := &cfg.RevolutConfig{
		APIKey:        "test_api_key",
		BaseURL:       "https://sandbox-merchant.revolut.com",
		WebhookSecret: "test_webhook_secret",
		IsSandbox:     true,
	}

	service := &RevolutPaymentService{
		webhookSecret: config.WebhookSecret,
		config:        config,
	}

	// Test with valid signature
	payload := []byte(`{"event": "ORDER_COMPLETED", "order_id": "test-order-id"}`)
	timestamp := "1683650202360"

	// Create a valid signature
	payloadToSign := fmt.Sprintf("v1.%s.%s", timestamp, string(payload))
	h := hmac.New(sha256.New, []byte(config.WebhookSecret))
	h.Write([]byte(payloadToSign))
	validSignature := "v1=" + hex.EncodeToString(h.Sum(nil))

	result := service.validateWebhookSignature(payload, validSignature, timestamp)
	assert.True(t, result)

	// Test with invalid signature
	invalidSignature := "v1=invalid_signature"
	result = service.validateWebhookSignature(payload, invalidSignature, timestamp)
	assert.False(t, result)

	// Test with empty signature
	result = service.validateWebhookSignature(payload, "", timestamp)
	assert.False(t, result)

	// Test with malformed signature
	result = service.validateWebhookSignature(payload, "invalid_format", timestamp)
	assert.False(t, result)

	// Test with empty webhook secret (should skip validation)
	service.webhookSecret = ""
	result = service.validateWebhookSignature(payload, validSignature, timestamp)
	assert.True(t, result, "Should skip validation when webhook secret is empty")

	// Test with different timestamp
	differentTimestamp := "1683650202361"
	service.webhookSecret = config.WebhookSecret // Restore the secret
	result = service.validateWebhookSignature(payload, validSignature, differentTimestamp)
	assert.False(t, result, "Should fail with different timestamp")

	// Test with different payload
	differentPayload := []byte(`{"event": "ORDER_COMPLETED", "order_id": "different-order-id"}`)
	result = service.validateWebhookSignature(differentPayload, validSignature, timestamp)
	assert.False(t, result, "Should fail with different payload")
}

func TestRevolutPaymentService_WebhookSignatureValidationWithTimestamp(t *testing.T) {
	config := &cfg.RevolutConfig{
		APIKey:        "test_api_key",
		BaseURL:       "https://sandbox-merchant.revolut.com",
		WebhookSecret: "wsk_r59a4HfWVAKycbCaNO1RvgCJec02gRd8",
		IsSandbox:     true,
	}

	service := &RevolutPaymentService{
		webhookSecret: config.WebhookSecret,
		config:        config,
	}

	// Test data from Revolut documentation example
	rawPayload := `{"event": "ORDER_COMPLETED","order_id": "9fc01989-3f61-4484-a5d9-ffe768531be9","merchant_order_ext_ref": "Test #3928"}`
	timestamp := "1683650202360"

	// Create the signature as Revolut would
	payloadToSign := fmt.Sprintf("v1.%s.%s", timestamp, rawPayload)
	h := hmac.New(sha256.New, []byte(config.WebhookSecret))
	h.Write([]byte(payloadToSign))
	computedSignature := hex.EncodeToString(h.Sum(nil))

	// Test the validation function with our computed signature
	isValid := service.validateWebhookSignature([]byte(rawPayload), "v1="+computedSignature, timestamp)
	assert.True(t, isValid, "Signature validation should pass with correct data")

	// Test with wrong signature
	isValid = service.validateWebhookSignature([]byte(rawPayload), "v1=wrongsignature", timestamp)
	assert.False(t, isValid, "Signature validation should fail with wrong signature")

	// Test with wrong timestamp
	isValid = service.validateWebhookSignature([]byte(rawPayload), "v1="+computedSignature, "1683650202361")
	assert.False(t, isValid, "Signature validation should fail with wrong timestamp")

	// Test with wrong payload
	wrongPayload := `{"event": "ORDER_COMPLETED","order_id": "different-id","merchant_order_ext_ref": "Test #3928"}`
	isValid = service.validateWebhookSignature([]byte(wrongPayload), "v1="+computedSignature, timestamp)
	assert.False(t, isValid, "Signature validation should fail with wrong payload")
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
