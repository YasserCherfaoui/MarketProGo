package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/YasserCherfaoui/MarketProGo/cfg"
	"github.com/YasserCherfaoui/MarketProGo/models"
)

// GraphEmailProvider implements EmailProvider using Microsoft Graph API
type GraphEmailProvider struct {
	config      *cfg.OutlookConfig
	senderEmail string
	senderName  string
	authClient  *confidential.Client
	httpClient  *http.Client
}

// NewGraphEmailProvider creates a new Microsoft Graph API email provider
func NewGraphEmailProvider(config *cfg.OutlookConfig) (*GraphEmailProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("❌ GRAPH PROVIDER: Configuration is nil")
	}

	// Validate required configuration
	if err := validateGraphConfig(config); err != nil {
		return nil, fmt.Errorf("❌ GRAPH PROVIDER: Configuration validation failed: %w", err)
	}

	log.Printf("🔧 GRAPH PROVIDER: Initializing with TenantID: %s, ClientID: %s", config.TenantID, config.ClientID)
	log.Printf("🔧 GRAPH PROVIDER: ClientSecret length: %d characters", len(config.ClientSecret))

	// Create the confidential client for authentication
	log.Printf("🔧 GRAPH PROVIDER: Creating credential from client secret...")
	cred, err := confidential.NewCredFromSecret(config.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("❌ GRAPH PROVIDER: Failed to create credential from secret: %w", err)
	}
	log.Printf("✅ GRAPH PROVIDER: Credential created successfully")

	// Create the authority URL
	authorityURL := fmt.Sprintf("https://login.microsoftonline.com/%s", config.TenantID)
	log.Printf("🔧 GRAPH PROVIDER: Using authority URL: %s", authorityURL)

	// Validate authority URL format
	if !strings.HasPrefix(authorityURL, "https://login.microsoftonline.com/") {
		return nil, fmt.Errorf("❌ GRAPH PROVIDER: Invalid authority URL format: %s", authorityURL)
	}

	// Create the confidential client with proper error handling
	log.Printf("🔧 GRAPH PROVIDER: Creating confidential client...")

	// Try with default options instead of nil
	authClient, err := confidential.New(authorityURL, config.ClientID, cred)
	if err != nil {
		return nil, fmt.Errorf("❌ GRAPH PROVIDER: Failed to create auth client: %w", err)
	}
	log.Printf("✅ GRAPH PROVIDER: Confidential client created successfully")

	provider := &GraphEmailProvider{
		config:      config,
		senderEmail: config.SenderEmail,
		senderName:  config.SenderName,
		authClient:  &authClient,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}

	// Test the connection
	if err := provider.testConnection(); err != nil {
		return nil, fmt.Errorf("❌ GRAPH PROVIDER: Connection test failed: %w", err)
	}

	log.Printf("✅ GRAPH PROVIDER: Successfully initialized for sender: %s (%s)", config.SenderName, config.SenderEmail)
	return provider, nil
}

// validateGraphConfig validates the Graph API configuration
func validateGraphConfig(config *cfg.OutlookConfig) error {
	log.Printf("🔍 GRAPH PROVIDER: Validating configuration...")

	// Check for empty values
	if strings.TrimSpace(config.TenantID) == "" {
		return fmt.Errorf("TenantID is empty or contains only whitespace")
	}
	if strings.TrimSpace(config.ClientID) == "" {
		return fmt.Errorf("ClientID is empty or contains only whitespace")
	}
	if strings.TrimSpace(config.ClientSecret) == "" {
		return fmt.Errorf("ClientSecret is empty or contains only whitespace")
	}

	// Validate client secret format (should not contain spaces and should be reasonably long)
	clientSecret := strings.TrimSpace(config.ClientSecret)
	if len(clientSecret) < 10 {
		return fmt.Errorf("ClientSecret appears to be too short (minimum 10 characters expected)")
	}

	// Check for common issues with client secret
	if strings.Contains(clientSecret, " ") {
		return fmt.Errorf("ClientSecret contains spaces, which may cause issues")
	}

	// Log client secret format for debugging (without exposing the actual secret)
	log.Printf("🔍 GRAPH PROVIDER: ClientSecret format check - length: %d, contains spaces: %t",
		len(clientSecret), strings.Contains(clientSecret, " "))
	if strings.TrimSpace(config.SenderEmail) == "" {
		return fmt.Errorf("SenderEmail is empty or contains only whitespace")
	}
	if strings.TrimSpace(config.SenderName) == "" {
		return fmt.Errorf("SenderName is empty or contains only whitespace")
	}

	// Validate email format
	if !strings.Contains(config.SenderEmail, "@") {
		return fmt.Errorf("SenderEmail '%s' is not a valid email format", config.SenderEmail)
	}

	log.Printf("✅ GRAPH PROVIDER: Configuration validation passed")
	return nil
}

// testConnection tests the Graph API connection
func (p *GraphEmailProvider) testConnection() error {
	log.Printf("🔍 GRAPH PROVIDER: Testing connection to Microsoft Graph API...")

	ctx := context.Background()

	// Get access token
	scopes := []string{"https://graph.microsoft.com/.default"}
	result, err := p.authClient.AcquireTokenSilent(ctx, scopes)
	if err != nil {
		result, err = p.authClient.AcquireTokenByCredential(ctx, scopes)
		if err != nil {
			return fmt.Errorf("failed to acquire token: %w", err)
		}
	}

	// Test the connection by getting user profile
	url := "https://graph.microsoft.com/v1.0/users"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+result.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to test connection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read the response body to get more details about the error
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Graph API test failed with status %d (could not read response body)", resp.StatusCode)
		}
		return fmt.Errorf("Graph API test failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("✅ GRAPH PROVIDER: Successfully authenticated with Graph API")
	return nil
}

// SendEmail sends a single email via Microsoft Graph API
func (p *GraphEmailProvider) SendEmail(email *models.Email) error {
	if len(email.Recipients) == 0 {
		return fmt.Errorf("❌ GRAPH PROVIDER: No recipients specified")
	}

	log.Printf("📧 GRAPH PROVIDER: Sending email to %d recipients", len(email.Recipients))

	// Get access token
	ctx := context.Background()
	scopes := []string{"https://graph.microsoft.com/.default"}
	result, err := p.authClient.AcquireTokenSilent(ctx, scopes)
	if err != nil {
		result, err = p.authClient.AcquireTokenByCredential(ctx, scopes)
		if err != nil {
			return fmt.Errorf("❌ GRAPH PROVIDER: Failed to acquire token: %w", err)
		}
	}

	// Build recipients
	var toRecipients []GraphRecipient
	for _, recipient := range email.Recipients {
		toRecipients = append(toRecipients, GraphRecipient{
			EmailAddress: GraphEmailAddress{
				Address: recipient.Email,
				Name:    recipient.Name,
			},
		})
	}

	// Create message
	message := GraphMessage{
		Subject: email.Subject,
		Body: GraphItemBody{
			ContentType: "HTML",
			Content:     email.HTMLContent,
		},
		ToRecipients: toRecipients,
	}

	// Create send mail request
	requestBody := GraphSendMailRequest{
		Message:         message,
		SaveToSentItems: true,
	}

	// Serialize request body
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("❌ GRAPH PROVIDER: Failed to marshal request body: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/sendMail", p.senderEmail)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("❌ GRAPH PROVIDER: Failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+result.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	log.Printf("📧 GRAPH PROVIDER: Sending email via Graph API for user: %s", p.senderEmail)

	// Send the request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("❌ GRAPH PROVIDER: Failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle different response status codes with detailed error messages
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			if errorObj, ok := errorResponse["error"].(map[string]interface{}); ok {
				code := "unknown"
				message := "unknown error"
				if c, ok := errorObj["code"].(string); ok {
					code = c
				}
				if m, ok := errorObj["message"].(string); ok {
					message = m
				}

				log.Printf("❌ GRAPH PROVIDER: Graph API Error - Code: %s, Message: %s", code, message)

				// Provide specific guidance based on error codes
				switch code {
				case "ErrorAccessDenied":
					return fmt.Errorf("❌ GRAPH PROVIDER: Access denied. Please check:\n"+
						"1. App registration permissions (Mail.Send)\n"+
						"2. Admin consent for the application\n"+
						"3. User has mailbox access\n"+
						"4. Application has proper API permissions\n"+
						"Error: %s", message)
				case "ErrorInvalidRequest":
					return fmt.Errorf("❌ GRAPH PROVIDER: Invalid request. Please check:\n"+
						"1. Email format is valid\n"+
						"2. Subject is not empty\n"+
						"3. Content is properly formatted\n"+
						"4. Recipients are valid\n"+
						"Error: %s", message)
				case "ErrorQuotaExceeded":
					return fmt.Errorf("❌ GRAPH PROVIDER: Quota exceeded. Please check:\n"+
						"1. Mailbox storage limits\n"+
						"2. API rate limits\n"+
						"3. Daily sending limits\n"+
						"Error: %s", message)
				case "ErrorMailboxNotFound":
					return fmt.Errorf("❌ GRAPH PROVIDER: Mailbox not found. Please check:\n"+
						"1. Sender email exists in tenant\n"+
						"2. User has proper licensing (Microsoft 365)\n"+
						"3. Mailbox is enabled\n"+
						"Error: %s", message)
				case "ErrorInsufficientPermissions":
					return fmt.Errorf("❌ GRAPH PROVIDER: Insufficient permissions. Please check:\n"+
						"1. App has Mail.Send permission\n"+
						"2. Admin consent granted\n"+
						"3. User has mailbox access\n"+
						"Error: %s", message)
				default:
					return fmt.Errorf("❌ GRAPH PROVIDER: Graph API error (Code: %s): %s", code, message)
				}
			}
		}

		return fmt.Errorf("❌ GRAPH PROVIDER: Graph API returned status %d", resp.StatusCode)
	}

	log.Printf("✅ GRAPH PROVIDER: Email sent successfully to %d recipients", len(email.Recipients))
	return nil
}

// SendBulkEmail sends multiple emails in batch via Microsoft Graph API
func (p *GraphEmailProvider) SendBulkEmail(emails []*models.Email) error {
	log.Printf("📧 GRAPH PROVIDER: Sending %d emails in bulk", len(emails))

	for i, email := range emails {
		log.Printf("📧 GRAPH PROVIDER: Processing email %d/%d", i+1, len(emails))

		if err := p.SendEmail(email); err != nil {
			return fmt.Errorf("❌ GRAPH PROVIDER: Failed to send bulk email %d/%d: %w", i+1, len(emails), err)
		}

		// Add small delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("✅ GRAPH PROVIDER: Bulk email operation completed successfully")
	return nil
}

// GetDeliveryStatus retrieves the delivery status of an email
func (p *GraphEmailProvider) GetDeliveryStatus(emailID string) (DeliveryStatus, error) {
	log.Printf("🔍 GRAPH PROVIDER: Getting delivery status for email: %s", emailID)

	ctx := context.Background()

	// Get access token
	scopes := []string{"https://graph.microsoft.com/.default"}
	result, err := p.authClient.AcquireTokenSilent(ctx, scopes)
	if err != nil {
		result, err = p.authClient.AcquireTokenByCredential(ctx, scopes)
		if err != nil {
			return DeliveryStatusFailed, fmt.Errorf("failed to acquire token: %w", err)
		}
	}

	// Try to get the message from the sent items folder
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/messages/%s", p.senderEmail, emailID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return DeliveryStatusFailed, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+result.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return DeliveryStatusFailed, fmt.Errorf("failed to get email status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("⚠️ GRAPH PROVIDER: Email %s not found in sent items", emailID)
		return DeliveryStatusPending, nil
	}

	if resp.StatusCode != http.StatusOK {
		return DeliveryStatusFailed, fmt.Errorf("failed to get email status: HTTP %d", resp.StatusCode)
	}

	// For now, assume sent if we can retrieve the message
	return DeliveryStatusSent, nil
}

// GetBounceList retrieves list of bounced email addresses
func (p *GraphEmailProvider) GetBounceList() ([]string, error) {
	log.Printf("🔍 GRAPH PROVIDER: Getting bounce list (not implemented in Graph API)")

	// Microsoft Graph API doesn't provide direct bounce list
	// In a production environment, you might need to:
	// 1. Use webhooks to track bounces
	// 2. Query Exchange Online Protection (EOP) logs
	// 3. Use Microsoft 365 Defender APIs
	// 4. Implement custom tracking

	log.Printf("⚠️ GRAPH PROVIDER: Bounce list functionality not available in Graph API")
	return []string{}, nil
}

// GetComplaintList retrieves list of email addresses that complained
func (p *GraphEmailProvider) GetComplaintList() ([]string, error) {
	log.Printf("🔍 GRAPH PROVIDER: Getting complaint list (not implemented in Graph API)")

	// Microsoft Graph API doesn't provide direct complaint list
	// Similar to bounce list, this would require:
	// 1. Webhook implementation
	// 2. EOP integration
	// 3. Custom tracking system

	log.Printf("⚠️ GRAPH PROVIDER: Complaint list functionality not available in Graph API")
	return []string{}, nil
}

// GetProviderInfo returns information about the Graph provider
func (p *GraphEmailProvider) GetProviderInfo() map[string]interface{} {
	return map[string]interface{}{
		"provider":     "Microsoft Graph API",
		"sender_email": p.senderEmail,
		"sender_name":  p.senderName,
		"tenant_id":    p.config.TenantID,
		"client_id":    p.config.ClientID,
		"capabilities": []string{
			"Send Email",
			"Send Bulk Email",
			"Get Delivery Status",
			"HTML Content Support",
			"Rich Text Support",
		},
		"limitations": []string{
			"No direct bounce tracking",
			"No direct complaint tracking",
			"Requires Microsoft 365 license",
			"Rate limited by Microsoft",
		},
		"required_permissions": []string{
			"Mail.Send",
			"Mail.ReadWrite",
			"User.Read",
		},
		"setup_requirements": []string{
			"Azure AD App Registration",
			"Microsoft 365 License",
			"Admin Consent",
			"Proper API Permissions",
		},
	}
}
