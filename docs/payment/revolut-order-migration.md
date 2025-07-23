# Revolut Payment Integration: Migration to Order-Based API

## Overview

This document outlines the migration from the payment-based Revolut API to the order-based API, which provides a more robust and feature-rich approach to handling payments.

## Why Order-Based API?

The order-based API offers several advantages over the payment-based approach:

1. **Checkout URL Generation**: Orders automatically provide a checkout URL that customers can use to complete payments
2. **Better State Management**: Orders have clearer state transitions (pending → authorized → completed)
3. **Rich Metadata Support**: Orders support line items, customer information, shipping details, and more
4. **Industry-Specific Features**: Support for retail, marketplace, crypto, and other industry-specific data
5. **Automatic Capture**: Orders can be configured for automatic or manual capture
6. **Better Webhook Support**: More comprehensive webhook events for order lifecycle management

## Changes Made

### 1. Updated Revolut Client (`payment/revolut/client.go`)

#### New Data Structures
- **Customer**: Represents customer information for orders
- **LineItem**: Represents individual items in an order with quantity, price, and metadata
- **Quantity**: Represents quantity with value and unit
- **OrderRequest**: Updated to match Revolut Merchant API specification
- **OrderResponse**: Updated to include all order response fields

#### Key Changes
- Amount now uses `int64` (minor units) instead of `float64`
- Customer information is structured as a nested object
- Support for line items, shipping, and industry-specific data
- Updated API endpoints to use `/api/orders` instead of `/api/1.0/orders`

### 2. Updated Payment Service (`payment/revolut_service.go`)

#### CreatePayment Method
- Converts amount from major units (e.g., 50.00) to minor units (5000 cents)
- Creates proper customer object structure
- Uses automatic capture mode by default
- Includes redirect URL for post-payment navigation

#### Status Mapping
- Updated to use lowercase state names from Revolut API
- Maps order states to payment statuses correctly

#### Webhook Processing
- Simplified webhook event processing
- Removed dependency on merchant order external reference
- Better error handling and logging

### 3. Configuration Updates

The configuration remains the same, but now uses the correct API endpoints:
- Sandbox: `https://sandbox-merchant.revolut.com`
- Production: `https://merchant.revolut.com`

## API Request Structure

### Before (Payment-Based)
```json
{
  "amount": 50.00,
  "currency": "GBP",
  "merchant_order_id": "123",
  "customer_email": "user@example.com",
  "customer_name": "John Doe",
  "description": "Payment for order 123"
}
```

### After (Order-Based)
```json
{
  "amount": 5000,
  "currency": "GBP",
  "description": "Payment for order 123",
  "customer": {
    "id": "123",
    "full_name": "John Doe",
    "email": "user@example.com",
    "phone": "+1234567890"
  },
  "capture_mode": "automatic",
  "enforce_challenge": "automatic",
  "redirect_url": "https://example.com/return",
  "metadata": {
    "order_id": "123"
  }
}
```

## API Response Structure

### Before
```json
{
  "id": "order_id",
  "public_id": "public_id",
  "amount": 50.00,
  "currency": "GBP",
  "state": "PENDING",
  "checkout_url": "https://checkout.revolut.com/..."
}
```

### After
```json
{
  "id": "6516e61c-d279-a454-a837-bc52ce55ed49",
  "token": "0adc0e3c-ab44-4f33-bcc0-534ded7354ce",
  "type": "payment",
  "state": "pending",
  "amount": 5000,
  "currency": "GBP",
  "outstanding_amount": 5000,
  "capture_mode": "automatic",
  "checkout_url": "https://checkout.revolut.com/payment-link/0adc0e3c-ab44-4f33-bcc0-534ded7354ce",
  "enforce_challenge": "automatic"
}
```

## Benefits of the Migration

1. **Simplified Integration**: Orders provide a complete payment flow with checkout URLs
2. **Better User Experience**: Customers get a hosted checkout page
3. **Enhanced Security**: Automatic 3DS challenge handling
4. **Rich Data Support**: Line items, customer details, shipping information
5. **Industry Compliance**: Support for various industry-specific requirements
6. **Better Error Handling**: More detailed error responses and webhook events

## Testing

The migration includes comprehensive tests for:
- Status mapping between Revolut order states and payment statuses
- Webhook signature validation
- Order creation flow

Run tests with:
```bash
go test ./payment -v
```

## Migration Checklist

- [x] Updated Revolut client to use order-based API
- [x] Modified payment service to create orders instead of payments
- [x] Updated status mapping to use correct state names
- [x] Enhanced webhook processing for order events
- [x] Added comprehensive tests
- [x] Updated documentation

## Environment Variables

Ensure the following environment variables are set:
```bash
REVOLUT_API_KEY=your_api_key
REVOLUT_MERCHANT_ID=your_merchant_id
REVOLUT_WEBHOOK_SECRET=your_webhook_secret
REVOLUT_SANDBOX=true  # Set to false for production
```

## Next Steps

1. Test the integration in sandbox environment
2. Verify webhook processing works correctly
3. Test with real payment scenarios
4. Monitor order states and payment flows
5. Update frontend to use checkout URLs from order responses

## References

- [Revolut Merchant API Documentation](https://developer.revolut.com/docs/merchant/create-order)
- [Revolut Webhook Documentation](https://developer.revolut.com/docs/guides/accept-payments/tutorials/work-with-webhooks/using-webhooks) 