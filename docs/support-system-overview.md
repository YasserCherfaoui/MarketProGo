# Support System Overview

## Introduction

The Algeria Market ecommerce platform now includes a comprehensive customer support system with four main features:

1. **Help Ticket System** - For customer support requests
2. **Report Abuse** - For reporting inappropriate content or behavior
3. **Contact Us** - For general inquiries and feedback
4. **Submit Dispute** - For order and payment disputes

## Architecture

### Database Models

The support system uses the following main models:

#### SupportTicket
- `UserID` - User who created the ticket
- `OrderID` - Associated order (optional)
- `Title` - Ticket title
- `Description` - Detailed description
- `Category` - Ticket category (GENERAL, ORDER, PAYMENT, etc.)
- `Priority` - Priority level (LOW, MEDIUM, HIGH, URGENT)
- `Status` - Current status (OPEN, IN_PROGRESS, WAITING, RESOLVED, CLOSED)
- `AssignedTo` - Admin assigned to handle the ticket
- `Resolution` - Resolution details
- `InternalNotes` - Internal notes for admins
- `IsEscalated` - Whether ticket has been escalated
- `Attachments` - File attachments
- `Responses` - Ticket responses

#### AbuseReport
- `ReporterID` - User who reported the abuse
- `ReportedUserID` - User being reported (optional)
- `ProductID` - Product being reported (optional)
- `ReviewID` - Review being reported (optional)
- `OrderID` - Order being reported (optional)
- `Category` - Abuse category (HARASSMENT, SPAM, INAPPROPRIATE, etc.)
- `Description` - Detailed description
- `Status` - Report status (PENDING, REVIEWING, RESOLVED, DISMISSED)
- `Severity` - Severity level (LOW, MEDIUM, HIGH, CRITICAL)
- `Attachments` - Evidence attachments

#### ContactInquiry
- `UserID` - User who submitted inquiry (optional)
- `Name` - Contact name
- `Email` - Contact email
- `Phone` - Contact phone
- `Subject` - Inquiry subject
- `Message` - Inquiry message
- `Category` - Inquiry category (GENERAL, SALES, SUPPORT, etc.)
- `Status` - Inquiry status (NEW, IN_PROGRESS, RESPONDED, CLOSED)
- `Priority` - Priority level (LOW, NORMAL, HIGH, URGENT)
- `Response` - Admin response
- `InternalNotes` - Internal notes

#### Dispute
- `UserID` - User who submitted the dispute
- `OrderID` - Associated order (optional)
- `PaymentID` - Associated payment (optional)
- `Title` - Dispute title
- `Description` - Detailed description
- `Category` - Dispute category (ORDER, PAYMENT, PRODUCT, etc.)
- `Status` - Dispute status (OPEN, IN_PROGRESS, UNDER_REVIEW, RESOLVED, CLOSED, ESCALATED)
- `Priority` - Priority level (LOW, MEDIUM, HIGH, URGENT)
- `Amount` - Dispute amount
- `Currency` - Amount currency
- `Attachments` - Evidence attachments
- `Responses` - Dispute responses

## API Endpoints

### Standard Response Envelope
All endpoints return a consistent envelope:

```json
{
  "status": 200,
  "message": "Human readable summary",
  "data": { /* endpoint-specific payload (optional) */ },
  "error": {
    "code": "machine_readable_code",
    "description": "detailed error" 
  }
}
```

- On success: `status` is 200/201 and `data` is present, `error` is omitted.
- On error: `status` is 4xx/5xx and `error` is present, `data` is omitted.

### Support Tickets

#### Create Ticket
```
POST /api/v1/tickets/
```
**Request Body:**
```json
{
  "title": "Order not received",
  "description": "I placed an order 2 weeks ago but haven't received it yet",
  "category": "ORDER",
  "priority": "HIGH",
  "order_id": 123,
  "attachments": [
    {
      "file_name": "receipt.pdf",
      "file_url": "https://example.com/receipt.pdf",
      "file_size": 1024,
      "file_type": "application/pdf"
    }
  ]
}
```
**Response (200):**
```json
{
  "status": 200,
  "message": "Support ticket created successfully",
  "data": {
    "ID": 101,
    "user_id": 1,
    "order_id": 123,
    "title": "Order not received",
    "description": "I placed an order 2 weeks ago but haven't received it yet",
    "category": "ORDER",
    "priority": "HIGH",
    "status": "OPEN",
    "attachments": [
      {
        "ID": 501,
        "ticket_id": 101,
        "file_name": "receipt.pdf",
        "file_url": "https://example.com/receipt.pdf",
        "file_size": 1024,
        "file_type": "application/pdf"
      }
    ]
  }
}
```
**Error (400):**
```json
{
  "status": 400,
  "message": "Title is required",
  "error": {
    "code": "support/create-ticket",
    "description": "Title is required"
  }
}
```

#### Get User Tickets
```
GET /api/v1/tickets/
```
**Response (200):**
```json
{
  "status": 200,
  "message": "User tickets retrieved successfully",
  "data": [
    { "ID": 101, "user_id": 1, "title": "Order not received", "category": "ORDER", "priority": "HIGH", "status": "OPEN" }
  ]
}
```

#### Get Specific Ticket
```
GET /api/v1/tickets/{id}
```
**Response (200):**
```json
{
  "status": 200,
  "message": "Ticket retrieved successfully",
  "data": {
    "ID": 101,
    "user_id": 1,
    "title": "Order not received",
    "status": "OPEN",
    "responses": [
      { "ID": 801, "ticket_id": 101, "user_id": 999, "message": "We're checking this now", "is_from_admin": true }
    ]
  }
}
```

#### Update Ticket
```
PUT /api/v1/tickets/{id}
```
**Response (200):**
```json
{
  "status": 200,
  "message": "Ticket updated successfully",
  "data": { "ID": 101, "status": "IN_PROGRESS" }
}
```

#### Add Response to Ticket
```
POST /api/v1/tickets/{id}/responses
```
**Request Body:**
```json
{ "message": "Thanks for the update", "is_internal": false }
```
**Response (200):**
```json
{
  "status": 200,
  "message": "Response added successfully",
  "data": { "ID": 802, "ticket_id": 101, "user_id": 1, "message": "Thanks for the update", "is_from_admin": false }
}
```

#### Delete Ticket (Admin only)
```
DELETE /api/v1/tickets/{id}
```
**Response (200):**
```json
{ "status": 200, "message": "Ticket deleted successfully" }
```

#### Get All Tickets (Admin only)
```
GET /api/v1/admin/tickets/
```
**Response (200):**
```json
{ "status": 200, "message": "All tickets retrieved successfully", "data": [ /* tickets */ ] }
```

### Abuse Reports

#### Create Abuse Report
```
POST /api/v1/abuse/reports
```
**Request Body:**
```json
{
  "reported_user_id": 456,
  "category": "HARASSMENT",
  "description": "User is sending inappropriate messages",
  "severity": "HIGH",
  "attachments": [
    {
      "file_name": "screenshot.png",
      "file_url": "https://example.com/screenshot.png",
      "file_size": 2048,
      "file_type": "image/png"
    }
  ]
}
```
**Response (200):**
```json
{
  "status": 200,
  "message": "Abuse report created successfully",
  "data": {
    "ID": 301,
    "reporter_id": 1,
    "reported_user_id": 456,
    "category": "HARASSMENT",
    "status": "PENDING",
    "severity": "HIGH"
  }
}
```

#### Get User Abuse Reports
```
GET /api/v1/abuse/reports
```
**Response (200):**
```json
{ "status": 200, "message": "User abuse reports retrieved successfully", "data": [ /* reports */ ] }
```

#### Get Specific Abuse Report
```
GET /api/v1/abuse/reports/{id}
```
**Response (200):**
```json
{ "status": 200, "message": "Abuse report retrieved successfully", "data": { /* report */ } }
```

#### Update Abuse Report (Admin only)
```
PUT /api/v1/abuse/reports/{id}
```
**Response (200):**
```json
{ "status": 200, "message": "Abuse report updated successfully", "data": { "status": "REVIEWING" } }
```

#### Delete Abuse Report (Admin only)
```
DELETE /api/v1/abuse/reports/{id}
```
**Response (200):**
```json
{ "status": 200, "message": "Abuse report deleted successfully" }
```

#### Get All Abuse Reports (Admin only)
```
GET /api/v1/admin/abuse/reports
```
**Response (200):**
```json
{ "status": 200, "message": "All abuse reports retrieved successfully", "data": [ /* reports */ ] }
```

### Contact Inquiries

#### Create Contact Inquiry
```
POST /api/v1/contact/inquiries
```
**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "+1234567890",
  "subject": "General Inquiry",
  "message": "I have a question about your products",
  "category": "GENERAL",
  "priority": "NORMAL"
}
```
**Response (200):**
```json
{ "status": 200, "message": "Contact inquiry submitted successfully", "data": { "ID": 401, "email": "john@example.com", "status": "NEW" } }
```

#### Get User Contact Inquiries
```
GET /api/v1/contact/inquiries
```
**Response (200):**
```json
{ "status": 200, "message": "User contact inquiries retrieved successfully", "data": [ /* inquiries */ ] }
```

#### Get Specific Contact Inquiry
```
GET /api/v1/contact/inquiries/{id}
```
**Response (200):**
```json
{ "status": 200, "message": "Contact inquiry retrieved successfully", "data": { /* inquiry */ } }
```

#### Update Contact Inquiry (Admin only)
```
PUT /api/v1/contact/inquiries/{id}
```
**Response (200):**
```json
{ "status": 200, "message": "Contact inquiry updated successfully", "data": { "status": "RESPONDED" } }
```

#### Delete Contact Inquiry (Admin only)
```
DELETE /api/v1/contact/inquiries/{id}
```
**Response (200):**
```json
{ "status": 200, "message": "Contact inquiry deleted successfully" }
```

#### Get All Contact Inquiries (Admin only)
```
GET /api/v1/admin/contact/inquiries
```
**Response (200):**
```json
{ "status": 200, "message": "All contact inquiries retrieved successfully", "data": [ /* inquiries */ ] }
```

### Disputes

#### Create Dispute
```
POST /api/v1/disputes/
```
**Request Body:**
```json
{
  "order_id": 123,
  "title": "Wrong item received",
  "description": "I ordered a blue shirt but received a red one",
  "category": "ORDER",
  "priority": "HIGH",
  "amount": 29.99,
  "currency": "USD",
  "attachments": [
    {
      "file_name": "wrong_item.jpg",
      "file_url": "https://example.com/wrong_item.jpg",
      "file_size": 1024,
      "file_type": "image/jpeg"
    }
  ]
}
```
**Response (200):**
```json
{
  "status": 200,
  "message": "Dispute created successfully",
  "data": {
    "ID": 201,
    "user_id": 1,
    "order_id": 123,
    "title": "Wrong item received",
    "category": "ORDER",
    "priority": "HIGH",
    "status": "OPEN"
  }
}
```

#### Get User Disputes
```
GET /api/v1/disputes/
```
**Response (200):**
```json
{ "status": 200, "message": "User disputes retrieved successfully", "data": [ /* disputes */ ] }
```

#### Get Specific Dispute
```
GET /api/v1/disputes/{id}
```
**Response (200):**
```json
{ "status": 200, "message": "Dispute retrieved successfully", "data": { /* dispute incl. responses */ } }
```

#### Update Dispute
```
PUT /api/v1/disputes/{id}
```
**Response (200):**
```json
{ "status": 200, "message": "Dispute updated successfully", "data": { "status": "IN_PROGRESS" } }
```

#### Add Response to Dispute
```
POST /api/v1/disputes/{id}/responses
```
**Request Body:**
```json
{ "message": "We will send a replacement", "is_internal": false }
```
**Response (200):**
```json
{
  "status": 200,
  "message": "Response added successfully",
  "data": { "ID": 905, "dispute_id": 201, "user_id": 999, "message": "We will send a replacement", "is_from_admin": true }
}
```

#### Delete Dispute (Admin only)
```
DELETE /api/v1/disputes/{id}
```
**Response (200):**
```json
{ "status": 200, "message": "Dispute deleted successfully" }
```

#### Get All Disputes (Admin only)
```
GET /api/v1/admin/disputes/
```
**Response (200):**
```json
{ "status": 200, "message": "All disputes retrieved successfully", "data": [ /* disputes */ ] }
```

## Email Notifications

The support system includes automated email notifications for:

1. **Ticket Created** - Sent to user when a support ticket is created
2. **Ticket Updated** - Sent to user when a ticket is updated by admin
3. **Ticket Response** - Sent to user when admin responds to a ticket
4. **Dispute Created** - Sent to user when a dispute is created
5. **Dispute Updated** - Sent to user when a dispute is updated
6. **Contact Inquiry Confirmation** - Sent to user when contact inquiry is submitted

## Admin Features

### Dashboard
- View all support tickets, abuse reports, contact inquiries, and disputes
- Filter by status, category, priority, and date range
- Assign tickets to specific admins
- Update status and add internal notes
- Respond to users
- Escalate issues when needed

### Notifications
- Real-time notifications for new support requests
- Email notifications for high-priority issues
- Dashboard alerts for overdue tickets

### Reporting
- Support metrics and analytics
- Response time tracking
- Resolution rate statistics
- User satisfaction metrics

## Security and Permissions

### User Permissions
- Users can only view and manage their own tickets, reports, inquiries, and disputes
- Users cannot access admin-only endpoints
- File attachments are validated and scanned for security

### Admin Permissions
- Admins can view and manage all support items
- Admins can assign items to other admins
- Admins can escalate issues
- Admins can add internal notes not visible to users

### Data Protection
- All support data is encrypted at rest
- File attachments are stored securely
- Personal information is protected according to privacy regulations
- Audit logs are maintained for all actions

## Integration Points

### Order System
- Support tickets can be linked to specific orders
- Order information is automatically included in ticket details
- Order status updates can trigger support notifications

### Payment System
- Disputes can be linked to specific payments
- Payment information is automatically included in dispute details
- Payment status changes can trigger dispute notifications

### User System
- Support items are linked to user accounts
- User information is automatically included in support details
- User activity can trigger support notifications

### Email System
- Automated email notifications for all support activities
- Email templates for different types of notifications
- Email tracking and analytics

## Future Enhancements

### Planned Features
1. **Live Chat Support** - Real-time chat with support agents
2. **Knowledge Base** - Self-service support articles
3. **FAQ System** - Frequently asked questions
4. **Support Chatbot** - AI-powered support assistant
5. **Video Support** - Video calls with support agents
6. **Mobile App Support** - Native mobile support features

### Technical Improvements
1. **WebSocket Integration** - Real-time updates and notifications
2. **File Upload Optimization** - Better file handling and storage
3. **Search and Filtering** - Advanced search capabilities
4. **Analytics Dashboard** - Comprehensive reporting and analytics
5. **API Rate Limiting** - Better API performance and security

## Deployment

### Database Migration
Run the following command to apply the support system database migrations:

```bash
go run cmd/migrate/main.go
```

### Environment Variables
Add the following environment variables for the support system:

```bash
# Support System Configuration
SUPPORT_EMAIL_ENABLED=true
SUPPORT_NOTIFICATION_EMAIL=support@algeriamarket.co.uk
SUPPORT_MAX_ATTACHMENT_SIZE=10485760  # 10MB
SUPPORT_ALLOWED_FILE_TYPES=pdf,jpg,jpeg,png,doc,docx
```

### Monitoring
Monitor the following metrics for the support system:

1. **Response Time** - Average time to first response
2. **Resolution Time** - Average time to resolution
3. **Ticket Volume** - Number of tickets per day/week/month
4. **User Satisfaction** - User ratings and feedback
5. **System Performance** - API response times and error rates

## Support and Maintenance

### Regular Maintenance
1. **Database Cleanup** - Archive old tickets and reports
2. **File Cleanup** - Remove old attachments
3. **Performance Optimization** - Monitor and optimize queries
4. **Security Updates** - Regular security patches and updates

### Troubleshooting
1. **Email Delivery Issues** - Check email service configuration
2. **File Upload Problems** - Verify file storage configuration
3. **Performance Issues** - Monitor database and API performance
4. **User Access Issues** - Check authentication and authorization

## Conclusion

The support system provides a comprehensive solution for customer support, abuse reporting, contact inquiries, and dispute resolution. It is designed to be scalable, secure, and user-friendly while providing powerful admin tools for managing support operations.

The system integrates seamlessly with the existing ecommerce platform and provides a solid foundation for future enhancements and improvements.

## Resource Schemas (All Fields)

Below are the complete fields returned in the `data` payload for each resource. Timestamps follow ISO-8601. Relationship fields may be omitted when not preloaded.

### SupportTicket
```json
{
  "ID": 0,
  "CreatedAt": "2024-01-01T00:00:00Z",
  "UpdatedAt": "2024-01-01T00:00:00Z",
  "DeletedAt": null,
  "user_id": 0,
  "user": { /* User */ },
  "order_id": 0,
  "order": { /* Order */ },
  "title": "",
  "description": "",
  "category": "GENERAL|ORDER|PAYMENT|PRODUCT|SHIPPING|RETURN|TECHNICAL|ACCOUNT|BILLING|OTHER",
  "priority": "LOW|MEDIUM|HIGH|URGENT",
  "status": "OPEN|IN_PROGRESS|WAITING|RESOLVED|CLOSED",
  "assigned_to": 0,
  "assigned_user": { /* User */ },
  "resolution": "",
  "resolved_at": "2024-01-01T00:00:00Z",
  "resolved_by": 0,
  "resolved_by_user": { /* User */ },
  "internal_notes": "",
  "is_escalated": false,
  "escalated_at": "2024-01-01T00:00:00Z",
  "escalated_by": 0,
  "escalated_by_user": { /* User */ },
  "attachments": [ /* TicketAttachment[] */ ],
  "responses": [ /* TicketResponse[] */ ]
}
```

#### TicketAttachment
```json
{
  "ID": 0,
  "CreatedAt": "2024-01-01T00:00:00Z",
  "UpdatedAt": "2024-01-01T00:00:00Z",
  "DeletedAt": null,
  "ticket_id": 0,
  "file_name": "",
  "file_url": "",
  "file_size": 0,
  "file_type": ""
}
```

#### TicketResponse
```json
{
  "ID": 0,
  "CreatedAt": "2024-01-01T00:00:00Z",
  "UpdatedAt": "2024-01-01T00:00:00Z",
  "DeletedAt": null,
  "ticket_id": 0,
  "user_id": 0,
  "user": { /* User */ },
  "message": "",
  "is_internal": false,
  "is_from_admin": false
}
```

### AbuseReport
```json
{
  "ID": 0,
  "CreatedAt": "2024-01-01T00:00:00Z",
  "UpdatedAt": "2024-01-01T00:00:00Z",
  "DeletedAt": null,
  "reporter_id": 0,
  "reporter": { /* User */ },
  "reported_user_id": 0,
  "reported_user": { /* User */ },
  "product_id": 0,
  "product": { /* Product */ },
  "review_id": 0,
  "review": { /* ProductReview */ },
  "order_id": 0,
  "order": { /* Order */ },
  "category": "HARASSMENT|SPAM|INAPPROPRIATE|FRAUD|COPYRIGHT|VIOLENCE|DISCRIMINATION|OTHER",
  "description": "",
  "status": "PENDING|REVIEWING|RESOLVED|DISMISSED",
  "severity": "LOW|MEDIUM|HIGH|CRITICAL",
  "assigned_to": 0,
  "assigned_user": { /* User */ },
  "resolution": "",
  "resolved_at": "2024-01-01T00:00:00Z",
  "resolved_by": 0,
  "resolved_by_user": { /* User */ },
  "internal_notes": "",
  "attachments": [ /* AbuseReportAttachment[] */ ]
}
```

#### AbuseReportAttachment
```json
{
  "ID": 0,
  "CreatedAt": "2024-01-01T00:00:00Z",
  "UpdatedAt": "2024-01-01T00:00:00Z",
  "DeletedAt": null,
  "abuse_report_id": 0,
  "file_name": "",
  "file_url": "",
  "file_size": 0,
  "file_type": ""
}
```

### ContactInquiry
```json
{
  "ID": 0,
  "CreatedAt": "2024-01-01T00:00:00Z",
  "UpdatedAt": "2024-01-01T00:00:00Z",
  "DeletedAt": null,
  "user_id": 0,
  "user": { /* User */ },
  "name": "",
  "email": "",
  "phone": "",
  "subject": "",
  "message": "",
  "category": "GENERAL|SALES|SUPPORT|FEEDBACK|PARTNERSHIP|PRESS|OTHER",
  "status": "NEW|IN_PROGRESS|RESPONDED|CLOSED",
  "priority": "LOW|NORMAL|HIGH|URGENT",
  "assigned_to": 0,
  "assigned_user": { /* User */ },
  "response": "",
  "responded_at": "2024-01-01T00:00:00Z",
  "responded_by": 0,
  "responded_by_user": { /* User */ },
  "internal_notes": ""
}
```

### Dispute
```json
{
  "ID": 0,
  "CreatedAt": "2024-01-01T00:00:00Z",
  "UpdatedAt": "2024-01-01T00:00:00Z",
  "DeletedAt": null,
  "user_id": 0,
  "user": { /* User */ },
  "order_id": 0,
  "order": { /* Order */ },
  "payment_id": 0,
  "payment": { /* Payment */ },
  "title": "",
  "description": "",
  "category": "ORDER|PAYMENT|PRODUCT|SHIPPING|REFUND|BILLING|SERVICE|OTHER",
  "status": "OPEN|IN_PROGRESS|UNDER_REVIEW|RESOLVED|CLOSED|ESCALATED",
  "priority": "LOW|MEDIUM|HIGH|URGENT",
  "amount": 0,
  "currency": "USD",
  "assigned_to": 0,
  "assigned_user": { /* User */ },
  "resolution": "",
  "resolved_at": "2024-01-01T00:00:00Z",
  "resolved_by": 0,
  "resolved_by_user": { /* User */ },
  "internal_notes": "",
  "is_escalated": false,
  "escalated_at": "2024-01-01T00:00:00Z",
  "escalated_by": 0,
  "escalated_by_user": { /* User */ },
  "attachments": [ /* DisputeAttachment[] */ ],
  "responses": [ /* DisputeResponse[] */ ]
}
```

#### DisputeAttachment
```json
{
  "ID": 0,
  "CreatedAt": "2024-01-01T00:00:00Z",
  "UpdatedAt": "2024-01-01T00:00:00Z",
  "DeletedAt": null,
  "dispute_id": 0,
  "file_name": "",
  "file_url": "",
  "file_size": 0,
  "file_type": ""
}
```

#### DisputeResponse
```json
{
  "ID": 0,
  "CreatedAt": "2024-01-01T00:00:00Z",
  "UpdatedAt": "2024-01-01T00:00:00Z",
  "DeletedAt": null,
  "dispute_id": 0,
  "user_id": 0,
  "user": { /* User */ },
  "message": "",
  "is_internal": false,
  "is_from_admin": false
}
```
