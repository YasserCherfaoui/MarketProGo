# Email System Overview

## Introduction

The Algeria Market email system is a comprehensive, modular email service designed to handle all email communications for the e-commerce platform. The system uses Microsoft Graph API for Outlook Business as the sole email provider and sends all emails from the official business email address `enquirees@algeriamarket.co.uk`.

## Architecture

The email system follows a service-oriented architecture with the following components:

### Core Components

1. **Email Service** (`email/service.go`)
   - Central service that orchestrates all email operations
   - Handles template rendering, queue management, and provider integration
   - Implements the `EmailService` interface

2. **Email Provider** (`email/graph_provider.go`, `email/provider.go`)
   - Microsoft Graph API integration for Outlook Business
   - Mock provider for development and testing
   - Implements the `EmailProvider` interface

3. **Template Engine** (`email/template.go`)
   - HTML template rendering with Go templates
   - Supports hot-reloading for development
   - Converts HTML to plain text automatically

4. **Email Queue** (`email/queue.go`)
   - Redis-based queue for asynchronous email processing
   - Supports both Redis and mock implementations
   - Implements the `EmailQueue` interface

5. **Email Analytics** (`email/analytics.go`)
   - Tracks email delivery, opens, clicks, and bounces
   - Provides metrics and reporting capabilities
   - Implements the `EmailAnalytics` interface

6. **Email Models** (`models/email.go`)
   - Database models for emails and templates
   - Supports all email types and statuses
   - Includes recipient and metadata structures

## Email Types

The system supports the following email types as defined in the design document:

### Customer Emails
- **Password Reset** (`password_reset`)
- **Welcome** (`welcome`)
- **Order Confirmation** (`order_confirmation`)
- **Order Status Update** (`order_status_update`)
- **Payment Success** (`payment_success`)
- **Payment Failed** (`payment_failed`)
- **Security Alert** (`security_alert`)

### Marketing Emails
- **Promotional** (`promotional`)
- **Cart Recovery** (`cart_recovery`)
- **Re-engagement** (`re_engagement`)

### Admin Emails
- **Admin Notification** (`admin_notification`)

## Email Flow

### Transactional Email Flow
1. Business event triggers email sending
2. Email service creates email record with status `pending`
3. Template engine renders email content
4. Email is queued for delivery
5. Background worker processes queue
6. Email provider sends the email
7. Analytics service tracks delivery status
8. Email status is updated to `sent` or `failed`

### Bulk Email Flow
1. Marketing campaign is created
2. Email service identifies target recipients
3. Emails are rendered in batches
4. Bulk emails are queued for delivery
5. Background worker processes bulk delivery
6. Analytics service tracks bulk delivery metrics

## Configuration

### Environment Variables

```bash
# Email Configuration
EMAIL_PROVIDER=outlook
EMAIL_SENDER_EMAIL=enquirees@algeriamarket.co.uk
EMAIL_SENDER_NAME=Algeria Market

# Microsoft Graph API Configuration
OUTLOOK_TENANT_ID=your-tenant-id
OUTLOOK_CLIENT_ID=your-client-id
OUTLOOK_CLIENT_SECRET=your-client-secret
OUTLOOK_SENDER_EMAIL=enquirees@algeriamarket.co.uk
OUTLOOK_SENDER_NAME=Algeria Market

# Redis Configuration (Upstash)
UPSTASH_REDIS_REST_URL=redis://your-instance.upstash.io:port
UPSTASH_REDIS_REST_TOKEN=your-upstash-token
REDIS_POOL_SIZE=10
```

## API Endpoints

### Public Endpoints
- `POST /api/v1/email/send` - Send a single email
- `POST /api/v1/email/bulk` - Send bulk emails
- `POST /api/v1/email/transactional` - Send transactional email
- `GET /api/v1/email/status/:id` - Get email status
- `GET /api/v1/email/queue/status` - Get queue status
- `GET /api/v1/email/templates` - Get available templates

### Admin Endpoints (Require Authentication)
- `GET /api/v1/email/admin/list` - List emails with pagination
- `POST /api/v1/email/admin/retry/:id` - Retry failed email
- `POST /api/v1/email/admin/metrics` - Get email metrics

## Template System

### Template Structure
Templates are stored in `templates/emails/` as HTML files with Go template syntax:

```html
<!-- templates/emails/welcome.html -->
<!DOCTYPE html>
<html>
<head>
    <title>Welcome to Algeria Market</title>
</head>
<body>
    <h1>Welcome, {{.UserName}}!</h1>
    <p>Thank you for joining Algeria Market!</p>
</body>
</html>
```

### Template Data Structures
Each email type has a corresponding data structure defined in `email/template.go`:

```go
type WelcomeData struct {
    UserName       string `json:"user_name"`
    ActivationLink string `json:"activation_link"`
}
```

## Queue System

### Redis Queue
- Uses Upstash Redis for reliable email queuing
- Supports FIFO (First In, First Out) processing
- Implements retry mechanisms for failed emails
- Provides queue monitoring and status tracking

### Background Worker
- Continuously processes emails from the queue
- Handles provider failures gracefully
- Logs all email processing activities
- Updates email status in real-time

## Error Handling

### Email Delivery Errors
- **Temporary Failures**: Implemented exponential backoff retry
- **Permanent Failures**: Marked as failed and logged
- **Provider Failures**: Circuit breaker pattern for Microsoft Graph API

### Template Errors
- **Missing Templates**: Uses fallback template
- **Rendering Errors**: Logs detailed error information
- **Data Validation**: Validates template variables

## Security Considerations

### Email Content Security
- Template injection prevention through data sanitization
- Secure link generation with expiration
- Email content encryption for sensitive data

### Authentication and Authorization
- Secure API keys and credentials management
- Rate limiting for email endpoints
- Abuse monitoring and prevention

## Performance Optimization

### Email Delivery Optimization
- Batch processing for bulk emails
- Connection pooling for Microsoft Graph API
- Asynchronous processing with background workers

### Template Optimization
- Template caching for improved performance
- Template preloading on service startup
- Optimized HTML to text conversion

## Monitoring and Analytics

### Key Metrics
- **Delivery Metrics**: Sent count, delivered count, bounce rate
- **Engagement Metrics**: Open rate, click-through rate, unsubscribe rate
- **Performance Metrics**: Send time, delivery time, queue processing time

### Real-time Monitoring
- Current queue size and processing status
- Emails sent per minute
- Error rates and provider status
- Historical analytics and trends

## Development and Testing

### Mock Provider
For development and testing, the system includes a mock email provider that:
- Simulates email sending without actual delivery
- Logs all email operations for debugging
- Provides realistic testing environment

### Testing Strategy
- Unit tests for all email service methods
- Integration tests for email provider
- End-to-end tests for complete email flows
- Performance testing under load

## Deployment

### Production Setup
1. Configure Microsoft Graph API credentials
2. Set up Upstash Redis for email queue
3. Configure environment variables
4. Deploy with background worker enabled

### Development Setup
1. Use mock provider for local development
2. Configure local Redis or use mock queue
3. Set up email templates
4. Test with sample data

## Troubleshooting

### Common Issues
1. **Graph API Authentication**: Check tenant ID, client ID, and secret
2. **Queue Processing**: Verify Redis connection and worker status
3. **Template Rendering**: Check template syntax and data structure
4. **Email Delivery**: Monitor provider logs and status

### Debugging
- Enable detailed logging for email operations
- Check queue status endpoint for processing issues
- Monitor email status endpoint for delivery problems
- Review analytics for performance insights

## Future Enhancements

### Planned Features
1. **A/B Testing**: Template and content testing
2. **Advanced Analytics**: Detailed engagement tracking
3. **Template Management**: Admin interface for template editing
4. **Campaign Management**: Marketing campaign tools
5. **Webhook Integration**: Real-time delivery notifications

### Scalability Improvements
1. **Horizontal Scaling**: Multiple worker instances
2. **Queue Partitioning**: Separate queues for different email types
3. **Caching**: Redis caching for templates and data
4. **CDN Integration**: Static asset delivery optimization

## Conclusion

The Algeria Market email system provides a robust, scalable, and maintainable solution for all email communications. The modular architecture allows for easy extension and customization while maintaining high deliverability rates and comprehensive analytics capabilities. 