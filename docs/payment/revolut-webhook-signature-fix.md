# Revolut Webhook Signature Verification Fix

## Issue Identified

The webhook signature validation was failing with the error:
```
Signature validation failed. Expected: 68a32bd61485a29815f3ef0473e23f8cd0c30d96a8f16960a22e6ff922d29366, Received: 947dcf5884e75311294e15324b923b1d7230c4bbe91b6938969fe5a48efd83a9
```

## Root Cause

The signature validation was not following the correct process outlined in the [Revolut webhook documentation](https://developer.revolut.com/docs/guides/accept-payments/tutorials/work-with-webhooks/verify-the-payload-signature).

### **Incorrect Implementation**
The original code was only using the raw payload for signature validation:
```go
h := hmac.New(sha256.New, []byte(s.webhookSecret))
h.Write(payload)  // ❌ Only using raw payload
expectedSignature := hex.EncodeToString(h.Sum(nil))
```

### **Correct Implementation Required**
According to Revolut's documentation, the signature must be computed using a concatenated payload that includes:
1. The version (`v1`)
2. The timestamp from `Revolut-Request-Timestamp` header
3. The raw webhook payload

## Fix Implementation

### **1. Updated Signature Validation Method**
```go
func (s *RevolutPaymentService) validateWebhookSignature(payload []byte, signature string, timestamp string) bool {
    // Step 1: Prepare the payload to sign
    // payload_to_sign = v1.{timestamp}.{raw-payload}
    payloadToSign := fmt.Sprintf("v1.%s.%s", timestamp, string(payload))
    
    // Step 2: Compute the expected signature using HMAC-SHA256
    h := hmac.New(sha256.New, []byte(s.webhookSecret))
    h.Write([]byte(payloadToSign))  // ✅ Using concatenated payload
    expectedSignature := hex.EncodeToString(h.Sum(nil))
    
    // Step 3: Compare signatures
    return hmac.Equal([]byte(actualSignature), []byte(expectedSignature))
}
```

### **2. Updated Service Interface**
```go
// PaymentService interface updated to include timestamp
HandleWebhook(ctx context.Context, payload []byte, signature string, timestamp string) error
```

### **3. Updated Handler**
```go
// Handler now passes timestamp to service
if err := h.paymentService.HandleWebhook(c.Request.Context(), body, signature, timestamp); err != nil {
    // Handle error
}
```

## Revolut Webhook Security Process

### **Step 1: Prepare the Payload to Sign**
Concatenate the following data, separating each item with a full stop (`.`):
```
payload_to_sign = v1.{Revolut-Request-Timestamp}.{raw-payload}
```

**Example:**
```
v1.1683650202360.{"event": "ORDER_COMPLETED","order_id": "9fc01989-3f61-4484-a5d9-ffe768531be9","merchant_order_ext_ref": "Test #3928"}
```

### **Step 2: Compute the Expected Signature**
Use HMAC-SHA256 with:
- **Key**: The webhook signing secret
- **Message**: The prepared payload from Step 1

### **Step 3: Compare Signatures**
The computed signature must match exactly the signature (or one of multiple signatures) sent in the `Revolut-Signature` header.

## Webhook Headers

Revolut sends these headers with each webhook:

| Header | Description | Example |
|--------|-------------|---------|
| `Revolut-Request-Timestamp` | UNIX timestamp of the webhook event | `1683650202360` |
| `Revolut-Signature` | Signature of the request payload | `v1=09a9989dd8d9282c1d34974fc730f5cbfc4f4296941247e90ae5256590a11e8c` |

## Multiple Signatures Support

If multiple signing secrets are active, the `Revolut-Signature` header can contain multiple signatures separated by commas:
```
Revolut-Signature: v1=4fce70bda66b2e713be09fbb7ab1b31b0c8976ea4eeb01b244db7b99aa6482cb,v1=6ffbb59b2300aae63f272406069a9788598b792a944a07aba816edb039989a39
```

## Timestamp Validation

The implementation also includes timestamp validation to prevent replay attacks:
- Webhooks must be within 5 minutes of current time
- Allows for network delays and clock skew

## Testing

### **Unit Tests Added**
```go
func TestRevolutPaymentService_WebhookSignatureValidation(t *testing.T) {
    // Tests various signature validation scenarios
}

func TestRevolutPaymentService_WebhookSignatureValidationWithTimestamp(t *testing.T) {
    // Tests using Revolut documentation example
}
```

### **Test Scenarios**
- ✅ Valid signature with correct timestamp and payload
- ❌ Invalid signature
- ❌ Wrong timestamp
- ❌ Wrong payload
- ✅ Empty webhook secret (skips validation)
- ❌ Malformed signature format

## Debug Logging

Enhanced logging for troubleshooting:
```go
log.Printf("[DEBUG] Payload to sign: %s", payloadToSign)
log.Printf("[DEBUG] Webhook secret length: %d", len(s.webhookSecret))
log.Printf("[DEBUG] Payload length: %d", len(payload))
log.Printf("[DEBUG] Timestamp: %s", timestamp)
```

## Security Benefits

### **1. Payload Integrity**
- Ensures webhook payload hasn't been tampered with
- Any modification to the payload will result in a different signature

### **2. Replay Attack Prevention**
- Timestamp validation prevents old webhooks from being replayed
- 5-minute tolerance accounts for network delays

### **3. Source Verification**
- Confirms webhook originates from Revolut
- Uses HMAC-SHA256 for cryptographic security

## Configuration

### **Environment Variables**
```bash
REVOLUT_WEBHOOK_SECRET=wsk_your_webhook_secret_here
```

### **Webhook Secret Management**
- Webhook secret is generated when webhook is created
- Can be rotated for security
- Old secrets can remain valid during rotation period

## Monitoring

### **Log Messages to Monitor**
- `"Signature validation successful"` - ✅ Webhook verified
- `"Signature validation failed"` - ❌ Potential security issue
- `"Invalid signature format"` - ❌ Malformed webhook
- `"Webhook timestamp is too old"` - ❌ Replay attack attempt

### **Database Monitoring**
Track webhook processing success rates:
```sql
SELECT 
    COUNT(*) as total_webhooks,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as successful,
    ROUND(COUNT(CASE WHEN status = 'completed' THEN 1 END) * 100.0 / COUNT(*), 2) as success_rate
FROM payment_events 
WHERE event = 'webhook_received';
```

## Summary

The webhook signature verification now correctly follows Revolut's security requirements:

1. **Concatenates** version, timestamp, and payload
2. **Computes** HMAC-SHA256 signature using webhook secret
3. **Validates** signature matches exactly
4. **Includes** timestamp validation for replay attack prevention
5. **Provides** comprehensive logging for troubleshooting

This ensures that all webhooks are properly authenticated and secure from tampering or replay attacks. 