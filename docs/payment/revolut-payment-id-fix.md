# RevolutPaymentID Field Fix

## Issue Identified

The `revolut_payment_id` field in the database was always empty because it was never being set anywhere in the codebase.

## Root Cause

### 1. **Initial Payment Creation**
In `payment/revolut_service.go` line 151:
```go
payment := &models.Payment{
    OrderID:          req.OrderID,
    RevolutOrderID:   revolutResp.ID,
    RevolutPaymentID: "", // Will be set when payment is completed
    // ... other fields
}
```

The comment said "Will be set when payment is completed" but this was never actually happening.

### 2. **Incorrect Understanding of Revolut API**
The code was looking for `payment_id` in webhook payloads, but Revolut webhooks only contain `order_id`.

### 3. **Missing Payment ID Assignment**
The payment ID should be set during order creation, not during webhook processing.

## Correct Revolut API Understanding

### **Order Creation Response**
When creating an order, Revolut returns:
```json
{
  "id": "6516e61c-d279-a454-a837-bc52ce55ed49",  // This is the payment ID
  "token": "0adc0e3c-ab44-4f33-bcc0-534ded7354ce",
  "type": "payment",
  "state": "pending",
  "created_at": "2023-09-29T14:58:36.079398Z",
  "updated_at": "2023-09-29T14:58:36.079398Z",
  "amount": 500,
  "currency": "GBP",
  "outstanding_amount": 500,
  "capture_mode": "automatic",
  "checkout_url": "https://checkout.revolut.com/payment-link/0adc0e3c-ab44-4f33-bcc0-534ded7354ce",
  "enforce_challenge": "automatic"
}
```

### **Webhook Payload**
Revolut webhooks contain:
```json
{
  "event": "ORDER_COMPLETED",
  "order_id": "9fc01989-3f61-4484-a5d9-ffe768531be9",  // This matches the id from order creation
  "merchant_order_ext_ref": "Test #3928"
}
```

**Key Insight**: The `order_id` in webhooks corresponds to the `id` from the order creation response. This `id` is actually the payment ID.

## Fixes Implemented

### 1. **Correct Payment ID Assignment During Order Creation**
Updated the payment creation to set `RevolutPaymentID` immediately:

```go
payment := &models.Payment{
    OrderID:          req.OrderID,
    RevolutOrderID:   revolutResp.ID,
    RevolutPaymentID: revolutResp.ID, // The order ID from Revolut is actually the payment ID
    // ... other fields
}
```

### 2. **Removed Incorrect Webhook Processing**
Removed the incorrect logic that was looking for `payment_id` in webhook payloads:

```go
// REMOVED: This was incorrect
// if paymentID, ok := webhookData["payment_id"].(string); ok && paymentID != "" {
//     payment.RevolutPaymentID = paymentID
// }
```

### 3. **Updated Webhook Processing**
Updated webhook processing to understand that `order_id` in webhooks corresponds to the payment ID we already have:

```go
// The order_id from webhook is the same as the RevolutPaymentID we stored during order creation
// No need to extract payment_id since it doesn't exist in webhook payload
// The order_id in webhook corresponds to the id from order creation response
```

### 4. **Simplified Status Polling**
Removed unnecessary payment ID setting in status polling since it's already set during creation:

```go
// The RevolutPaymentID is already set during order creation
// No need to set it again here
```

### 5. **Enhanced Logging**
Added comprehensive logging to track payment ID usage:

```go
s.logPaymentEvent(ctx, payment.ID, "payment_completed", "Payment completed successfully", map[string]interface{}{
    "old_status":        oldStatus,
    "new_status":        payment.Status,
    "completed_at":      now,
    "revolut_payment_id": payment.RevolutPaymentID,
})
```

## When RevolutPaymentID Gets Set

The `RevolutPaymentID` field is now set **immediately during order creation**:

### 1. **Order Creation**
- When `CreateOrder` API call succeeds
- The `id` from Revolut response becomes the `RevolutPaymentID`
- This happens before any webhook events

### 2. **Webhook Events**
- Webhooks only update payment status
- The `order_id` in webhooks matches the `RevolutPaymentID` we already have
- No additional payment ID setting needed

### 3. **Status Polling**
- Only updates payment status
- Payment ID is already set from order creation

## Data Flow

### **Step 1: Order Creation**
1. Frontend calls `POST /api/v1/payments`
2. Backend calls Revolut `POST /api/orders`
3. Revolut returns order with `id: "6516e61c-d279-a454-a837-bc52ce55ed49"`
4. Backend stores this `id` as both `RevolutOrderID` and `RevolutPaymentID`
5. Frontend gets `checkout_url` and redirects user

### **Step 2: Payment Processing**
1. User completes payment on Revolut
2. Revolut sends webhook with `order_id: "6516e61c-d279-a454-a837-bc52ce55ed49"`
3. Backend matches this `order_id` to existing `RevolutPaymentID`
4. Backend updates payment status to `completed`

## Testing the Fix

### 1. **Check Order Creation**
Monitor logs when orders are created:
```
Revolut order created successfully: 6516e61c-d279-a454-a837-bc52ce55ed49
```

### 2. **Check Webhook Processing**
Monitor logs when webhooks are received:
```
Webhook event: ORDER_COMPLETED
revolut_order_id: 6516e61c-d279-a454-a837-bc52ce55ed49
revolut_payment_id: 6516e61c-d279-a454-a837-bc52ce55ed49
```

### 3. **Database Verification**
Query to verify payment IDs are set immediately:
```sql
SELECT id, revolut_order_id, revolut_payment_id, status, created_at
FROM payments 
WHERE revolut_payment_id IS NOT NULL AND revolut_payment_id != ''
ORDER BY created_at DESC;
```

## Manual Fix for Existing Records

For existing payments that don't have `RevolutPaymentID` set, you can:

### 1. **Use the API Method**
```go
err := paymentService.UpdateRevolutPaymentID(ctx, "payment_id", "revolut_order_id")
```

### 2. **Database Update**
```sql
UPDATE payments 
SET revolut_payment_id = revolut_order_id 
WHERE revolut_payment_id = '' AND revolut_order_id != '';
```

## Monitoring

### 1. **Log Monitoring**
Monitor these log messages:
- `"Revolut order created successfully"`
- `"Webhook event: ORDER_COMPLETED"`
- `"Payment completed successfully"`

### 2. **Database Monitoring**
Track the percentage of payments with `RevolutPaymentID` set:
```sql
SELECT 
    COUNT(*) as total_payments,
    COUNT(CASE WHEN revolut_payment_id != '' THEN 1 END) as with_payment_id,
    ROUND(COUNT(CASE WHEN revolut_payment_id != '' THEN 1 END) * 100.0 / COUNT(*), 2) as percentage
FROM payments;
```

## Summary

The `RevolutPaymentID` field is now properly set **immediately during order creation** using the `id` from Revolut's order creation response. This `id` serves as both the order ID and payment ID throughout the payment lifecycle.

**Key Changes:**
- Payment ID is set during order creation, not webhook processing
- Webhooks only contain `order_id` (no `payment_id`)
- The `order_id` in webhooks matches the payment ID we already have
- Simplified logic with no redundant payment ID setting

This ensures that the payment tracking is complete and accurate from the moment the order is created. 