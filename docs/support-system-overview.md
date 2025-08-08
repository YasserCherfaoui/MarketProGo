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

#### Get User Tickets
```
GET /api/v1/tickets/
```

#### Get Specific Ticket
```
GET /api/v1/tickets/{id}
```

#### Update Ticket
```
PUT /api/v1/tickets/{id}
```

#### Add Response to Ticket
```
POST /api/v1/tickets/{id}/responses
```

#### Delete Ticket (Admin only)
```
DELETE /api/v1/tickets/{id}
```

#### Get All Tickets (Admin only)
```
GET /api/v1/admin/tickets/
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

#### Get User Abuse Reports
```
GET /api/v1/abuse/reports
```

#### Get Specific Abuse Report
```
GET /api/v1/abuse/reports/{id}
```

#### Update Abuse Report (Admin only)
```
PUT /api/v1/abuse/reports/{id}
```

#### Delete Abuse Report (Admin only)
```
DELETE /api/v1/abuse/reports/{id}
```

#### Get All Abuse Reports (Admin only)
```
GET /api/v1/admin/abuse/reports
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

#### Get User Contact Inquiries
```
GET /api/v1/contact/inquiries
```

#### Get Specific Contact Inquiry
```
GET /api/v1/contact/inquiries/{id}
```

#### Update Contact Inquiry (Admin only)
```
PUT /api/v1/contact/inquiries/{id}
```

#### Delete Contact Inquiry (Admin only)
```
DELETE /api/v1/contact/inquiries/{id}
```

#### Get All Contact Inquiries (Admin only)
```
GET /api/v1/admin/contact/inquiries
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

#### Get User Disputes
```
GET /api/v1/disputes/
```

#### Get Specific Dispute
```
GET /api/v1/disputes/{id}
```

#### Update Dispute
```
PUT /api/v1/disputes/{id}
```

#### Add Response to Dispute
```
POST /api/v1/disputes/{id}/responses
```

#### Delete Dispute (Admin only)
```
DELETE /api/v1/disputes/{id}
```

#### Get All Disputes (Admin only)
```
GET /api/v1/admin/disputes/
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
