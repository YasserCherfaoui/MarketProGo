# Microsoft Graph API Email Provider Setup Guide

This guide will help you set up the Microsoft Graph API email provider for production use in the Algeria Market application.

## Prerequisites

1. **Microsoft 365 Account**: You need a Microsoft 365 account with Exchange Online
2. **Azure AD Subscription**: Access to Azure Active Directory
3. **Application Registration**: An app registered in Azure AD

## Step 1: Azure AD App Registration

### 1.1 Create App Registration

1. Go to [Azure Portal](https://portal.azure.com)
2. Navigate to **Azure Active Directory** > **App registrations**
3. Click **New registration**
4. Fill in the details:
   - **Name**: `Algeria Market Email Service`
   - **Supported account types**: `Accounts in this organizational directory only`
   - **Redirect URI**: `Web` > `https://localhost:8080/auth/callback`

### 1.2 Configure API Permissions

1. In your app registration, go to **API permissions**
2. Click **Add a permission**
3. Select **Microsoft Graph**
4. Choose **Application permissions**
5. Add the following permissions:
   - `Mail.Send` - Send emails
   - `Mail.ReadWrite` - Read and write emails
   - `User.Read` - Read user profiles
6. Click **Add permissions**
7. **Important**: Click **Grant admin consent for [Your Organization]**

### 1.3 Create Client Secret

1. Go to **Certificates & secrets**
2. Click **New client secret**
3. Add a description: `Email Service Secret`
4. Choose expiration (recommend 24 months for production)
5. **Copy the secret value immediately** (you won't see it again)

### 1.4 Get Application Details

Note down these values from the **Overview** page:
- **Application (client) ID**
- **Directory (tenant) ID**

## Step 2: Configuration

### 2.1 Update Configuration File

Add the following to your configuration file:

```json
{
  "outlook": {
    "tenant_id": "your-tenant-id",
    "client_id": "your-client-id", 
    "client_secret": "your-client-secret",
    "sender_email": "noreply@yourdomain.com",
    "sender_name": "Algeria Market"
  }
}
```

### 2.2 Environment Variables (Alternative)

You can also use environment variables:

```bash
export OUTLOOK_TENANT_ID="your-tenant-id"
export OUTLOOK_CLIENT_ID="your-client-id"
export OUTLOOK_CLIENT_SECRET="your-client-secret"
export OUTLOOK_SENDER_EMAIL="noreply@yourdomain.com"
export OUTLOOK_SENDER_NAME="Algeria Market"
```

## Step 3: Testing the Setup

### 3.1 Test Email Sending

Send a test email using the API:

```bash
curl -X POST http://localhost:8080/api/v1/email/send \
  -H "Content-Type: application/json" \
  -d '{
    "template": "welcome",
    "data": {
      "user_name": "Test User",
      "subject": "Welcome to Algeria Market"
    },
    "recipient": {
      "email": "test@example.com",
      "name": "Test User"
    }
  }'
```

### 3.2 Check Logs

Monitor the application logs for detailed error messages:

```bash
tail -f your-app.log
```

## Step 4: Troubleshooting

### Common Error Codes and Solutions

#### ErrorAccessDenied
**Symptoms**: 403 Forbidden error
**Solutions**:
1. Check if admin consent is granted
2. Verify API permissions are correctly assigned
3. Ensure the sender email exists in your tenant
4. Check if the user has a mailbox

#### ErrorInvalidRequest
**Symptoms**: 400 Bad Request error
**Solutions**:
1. Verify email format is valid
2. Check that subject is not empty
3. Ensure HTML content is properly formatted
4. Validate recipient email addresses

#### ErrorMailboxNotFound
**Symptoms**: Mailbox not found error
**Solutions**:
1. Ensure sender email exists in your Microsoft 365 tenant
2. Verify the user has a Microsoft 365 license
3. Check that Exchange Online is enabled
4. Confirm the mailbox is not disabled

#### ErrorInsufficientPermissions
**Symptoms**: Permission denied error
**Solutions**:
1. Grant admin consent for the application
2. Verify Mail.Send permission is assigned
3. Check if the app registration is active
4. Ensure the client secret is valid

#### ErrorQuotaExceeded
**Symptoms**: Rate limit or quota exceeded
**Solutions**:
1. Check mailbox storage limits
2. Monitor API rate limits
3. Implement retry logic with exponential backoff
4. Consider using bulk email for large campaigns

### Debugging Steps

1. **Check Authentication**:
   ```bash
   # Test token acquisition
   curl -X GET "https://graph.microsoft.com/v1.0/me" \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

2. **Verify Permissions**:
   ```bash
   # Check app permissions
   curl -X GET "https://graph.microsoft.com/v1.0/me" \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

3. **Test Mailbox Access**:
   ```bash
   # Test mailbox access
   curl -X GET "https://graph.microsoft.com/v1.0/users/SENDER_EMAIL/mailboxSettings" \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

## Step 5: Production Considerations

### 5.1 Security Best Practices

1. **Store Secrets Securely**:
   - Use environment variables or secure secret management
   - Never commit secrets to version control
   - Rotate client secrets regularly

2. **Network Security**:
   - Use HTTPS for all API calls
   - Implement proper firewall rules
   - Monitor API usage and logs

3. **Error Handling**:
   - Implement retry logic for transient failures
   - Log all errors for monitoring
   - Set up alerts for critical failures

### 5.2 Monitoring and Logging

1. **Application Logs**:
   - Monitor email sending success/failure rates
   - Track API response times
   - Log all Graph API errors

2. **Metrics to Track**:
   - Emails sent per hour/day
   - Error rates by error type
   - API quota usage
   - Delivery success rates

### 5.3 Rate Limiting

Microsoft Graph API has rate limits:
- **Mail.Send**: 10,000 requests per 10 minutes per app
- **User.Read**: 1,000 requests per 10 minutes per app

Implement proper rate limiting in your application.

## Step 6: Advanced Configuration

### 6.1 Webhook Setup (Optional)

For bounce tracking and delivery notifications:

1. Create a webhook endpoint in your application
2. Register the webhook with Microsoft Graph
3. Handle webhook notifications for email events

### 6.2 Custom Templates

Create custom email templates in the `templates/emails/` directory:

```html
<!-- templates/emails/welcome.html -->
<!DOCTYPE html>
<html>
<head>
    <title>{{.subject}}</title>
</head>
<body>
    <h1>Welcome {{.user_name}}!</h1>
    <p>Thank you for joining Algeria Market.</p>
</body>
</html>
```

## Support and Resources

- [Microsoft Graph API Documentation](https://docs.microsoft.com/en-us/graph/)
- [Graph Explorer](https://developer.microsoft.com/en-us/graph/graph-explorer)
- [Azure AD App Registration Guide](https://docs.microsoft.com/en-us/azure/active-directory/develop/quickstart-register-app)
- [Mail API Reference](https://docs.microsoft.com/en-us/graph/api/resources/mail-api-overview)

## Troubleshooting Checklist

- [ ] App registration created in Azure AD
- [ ] API permissions granted (Mail.Send, Mail.ReadWrite, User.Read)
- [ ] Admin consent granted
- [ ] Client secret created and copied
- [ ] Configuration updated with correct values
- [ ] Sender email exists in Microsoft 365 tenant
- [ ] User has Microsoft 365 license
- [ ] Exchange Online enabled
- [ ] Test email sent successfully
- [ ] Logs monitored for errors
- [ ] Rate limiting implemented
- [ ] Error handling in place 