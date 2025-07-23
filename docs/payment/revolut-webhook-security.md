# Revolut Webhook Security Implementation

## Overview

This document outlines the security measures implemented for handling Revolut webhook notifications, ensuring that only legitimate webhook events from Revolut are processed.

## Security Headers

Revolut sends webhook notifications with the following security headers:

| Header | Description | Format | Example |
|--------|-------------|--------|---------|
| `Revolut-Signature` | Signature of the request payload | `v1=hex_signature` | `v1=09a9989dd8d9282c1d34974fc730f5cbfc4f4296941247e90ae5256590a11e8c` |
| `Revolut-Request-Timestamp` | UNIX timestamp of the webhook event | `milliseconds_since_epoch` | `1683650202360` |

## Implementation Details

### 1. Signature Validation

The webhook signature is validated using HMAC-SHA256 with the webhook secret:

```go
// validateWebhookSignature validates the webhook signature according to Revolut's security requirements
func (s *RevolutPaymentService) validateWebhookSignature(payload []byte, signature string) bool {
    // Parse the signature format: v1=signature
    if len(signature) < 3 || signature[:2] != "v1" || signature[2] != '=' {
        return false
    }

    // Extract the actual signature (remove "v1=" prefix)
    actualSignature := signature[3:]

    // Create HMAC-SHA256 signature using the webhook secret
    h := hmac.New(sha256.New, []byte(s.webhookSecret))
    h.Write(payload)
    expectedSignature := hex.EncodeToString(h.Sum(nil))

    // Compare signatures using constant-time comparison
    return hmac.Equal([]byte(actualSignature), []byte(expectedSignature))
}
```

### 2. Timestamp Validation

The webhook timestamp is validated to prevent replay attacks:

```go
// validateWebhookTimestamp validates the webhook timestamp to prevent replay attacks
func (h *PaymentHandler) validateWebhookTimestamp(timestamp string) error {
    // Parse timestamp (milliseconds since epoch)
    ts, err := strconv.ParseInt(timestamp, 10, 64)
    if err != nil {
        return fmt.Errorf("invalid timestamp format: %w", err)
    }

    // Convert to seconds
    webhookTime := time.Unix(ts/1000, (ts%1000)*1000000)
    now := time.Now()

    // Allow webhook to be up to 5 minutes old (to account for network delays)
    maxAge := 5 * time.Minute
    if now.Sub(webhookTime) > maxAge {
        return fmt.Errorf("webhook timestamp is too old")
    }

    // Allow webhook to be up to 1 minute in the future (to account for clock skew)
    maxFuture := 1 * time.Minute
    if webhookTime.Sub(now) > maxFuture {
        return fmt.Errorf("webhook timestamp is too far in the future")
    }

    return nil
}
```

### 3. Webhook Handler

The webhook handler extracts and validates both security headers:

```go
// HandleWebhook handles POST /api/v1/payments/webhook
func (h *PaymentHandler) HandleWebhook(c *gin.Context) {
    // Read request body
    body, err := io.ReadAll(c.Request.Body)
    if err != nil {
        response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
        return
    }

    // Get webhook signature from header
    signature := c.GetHeader("Revolut-Signature")
    if signature == "" {
        response.GenerateErrorResponse(c, http.StatusBadRequest, "MISSING_SIGNATURE", "Webhook signature is required")
        return
    }

    // Get webhook timestamp from header for additional security
    timestamp := c.GetHeader("Revolut-Request-Timestamp")
    if timestamp == "" {
        response.GenerateErrorResponse(c, http.StatusBadRequest, "MISSING_TIMESTAMP", "Webhook timestamp is required")
        return
    }

    // Validate timestamp
    if err := h.validateWebhookTimestamp(timestamp); err != nil {
        response.GenerateErrorResponse(c, http.StatusBadRequest, "INVALID_TIMESTAMP", err.Error())
        return
    }

    // Process webhook
    if err := h.paymentService.HandleWebhook(c.Request.Context(), body, signature); err != nil {
        response.GenerateErrorResponse(c, http.StatusBadRequest, "WEBHOOK_PROCESSING_FAILED", err.Error())
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Webhook processed successfully",
    })
}
```

## Security Features

### 1. **Signature Verification**
- Validates HMAC-SHA256 signature using webhook secret
- Supports Revolut's `v1=signature` format
- Uses constant-time comparison to prevent timing attacks

### 2. **Timestamp Validation**
- Prevents replay attacks by validating webhook age
- Allows 5-minute tolerance for network delays
- Allows 1-minute tolerance for clock skew

### 3. **Header Validation**
- Ensures both required headers are present
- Validates header format and content

### 4. **Error Handling**
- Returns appropriate HTTP status codes
- Logs security violations for monitoring
- Provides detailed error messages for debugging

## Configuration

### Environment Variables

Set the webhook secret in your environment:

```bash
REVOLUT_WEBHOOK_SECRET=your_webhook_secret_here
```

### Webhook URL

Configure the webhook URL in your Revolut dashboard:
```
https://your-domain.com/api/v1/payments/webhook
```

## Testing

### Unit Tests

The implementation includes comprehensive unit tests:

```bash
go test ./payment -v -run TestRevolutPaymentService_WebhookSignatureValidation
```

### Test Cases

1. **Valid signature format with wrong signature** - Should fail
2. **Invalid signature format** - Should fail
3. **Empty signature** - Should fail
4. **Missing v1= prefix** - Should fail
5. **Empty webhook secret** - Should pass (for development)

## Security Best Practices

### 1. **Webhook Secret Management**
- Use a strong, unique webhook secret
- Store the secret securely (environment variables)
- Rotate the secret periodically
- Never commit secrets to version control

### 2. **Network Security**
- Use HTTPS for webhook endpoints
- Implement rate limiting
- Monitor for suspicious activity
- Log all webhook events

### 3. **Error Handling**
- Don't expose sensitive information in error messages
- Log security violations for monitoring
- Implement proper HTTP status codes

### 4. **Monitoring**
- Monitor webhook delivery success rates
- Alert on signature validation failures
- Track timestamp validation failures
- Monitor for unusual webhook patterns

## Troubleshooting

### Common Issues

1. **"Missing Revolut-Signature header"**
   - Ensure Revolut is sending the signature header
   - Check webhook URL configuration in Revolut dashboard

2. **"Invalid webhook signature"**
   - Verify webhook secret is correct
   - Check signature format (should start with "v1=")
   - Ensure payload hasn't been modified

3. **"Webhook timestamp is too old"**
   - Check server clock synchronization
   - Verify network connectivity
   - Consider increasing tolerance if needed

4. **"Missing Revolut-Request-Timestamp header"**
   - Ensure Revolut is sending the timestamp header
   - Check webhook configuration

### Debug Logging

Enable debug logging to troubleshoot webhook issues:

```go
log.Printf("[DEBUG] Revolut-Signature header received: %s", signature)
log.Printf("[DEBUG] Revolut-Request-Timestamp header received: %s", timestamp)
log.Printf("[DEBUG] Webhook request body read successfully: %d bytes", len(body))
```

## References

- [Revolut Webhook Documentation](https://developer.revolut.com/docs/guides/accept-payments/tutorials/work-with-webhooks/using-webhooks)
- [Revolut Security Headers](https://developer.revolut.com/docs/guides/accept-payments/tutorials/work-with-webhooks/using-webhooks#security) 