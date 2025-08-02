package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/YasserCherfaoui/MarketProGo/cfg"
	"github.com/YasserCherfaoui/MarketProGo/models"
)

// EmailProvider interface defines the contract for email providers
type EmailProvider interface {
	SendEmail(email *models.Email) error
	SendBulkEmail(emails []*models.Email) error
	GetDeliveryStatus(emailID string) (DeliveryStatus, error)
	GetBounceList() ([]string, error)
	GetComplaintList() ([]string, error)
}

// DeliveryStatus represents the delivery status of an email
type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusSent      DeliveryStatus = "sent"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
	DeliveryStatusBounced   DeliveryStatus = "bounced"
)

// Microsoft Graph API structures
type GraphMessage struct {
	Subject      string           `json:"subject"`
	Body         GraphItemBody    `json:"body"`
	ToRecipients []GraphRecipient `json:"toRecipients"`
}

type GraphItemBody struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type GraphRecipient struct {
	EmailAddress GraphEmailAddress `json:"emailAddress"`
}

type GraphEmailAddress struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type GraphSendMailRequest struct {
	Message         GraphMessage `json:"message"`
	SaveToSentItems bool         `json:"saveToSentItems"`
}

// OutlookProvider implements EmailProvider using Microsoft Graph API
type OutlookProvider struct {
	config      *cfg.OutlookConfig
	senderEmail string
	senderName  string
	authClient  *confidential.Client
	httpClient  *http.Client
}

// NewOutlookProvider creates a new Microsoft Graph API email provider
func NewOutlookProvider(config *cfg.OutlookConfig) (*OutlookProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is nil")
	}

	if config.TenantID == "" || config.ClientID == "" || config.ClientSecret == "" {
		return nil, fmt.Errorf("missing required Outlook configuration: TenantID, ClientID, and ClientSecret are required")
	}

	// Validate that the credentials are not empty strings
	if strings.TrimSpace(config.TenantID) == "" ||
		strings.TrimSpace(config.ClientID) == "" ||
		strings.TrimSpace(config.ClientSecret) == "" {
		return nil, fmt.Errorf("Outlook configuration contains empty values")
	}

	// Create the confidential client for authentication
	cred, err := confidential.NewCredFromSecret(config.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential from secret: %w", err)
	}

	// Create the authority URL
	authorityURL := fmt.Sprintf("https://login.microsoftonline.com/%s", config.TenantID)

	// Create the confidential client with proper error handling
	authClient, err := confidential.New(authorityURL, config.ClientID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth client: %w", err)
	}

	return &OutlookProvider{
		config:      config,
		senderEmail: config.SenderEmail,
		senderName:  config.SenderName,
		authClient:  &authClient,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// SendEmail sends a single email via Microsoft Graph API
func (p *OutlookProvider) SendEmail(email *models.Email) error {
	if len(email.Recipients) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	// Get access token
	ctx := context.Background()
	scopes := []string{"https://graph.microsoft.com/.default"}
	result, err := p.authClient.AcquireTokenSilent(ctx, scopes)
	if err != nil {
		result, err = p.authClient.AcquireTokenByCredential(ctx, scopes)
		if err != nil {
			return fmt.Errorf("failed to acquire token: %w", err)
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
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/sendMail", p.senderEmail)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+result.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Graph API returned status %d", resp.StatusCode)
	}

	return nil
}

// SendBulkEmail sends multiple emails in batch via Microsoft Graph API
func (p *OutlookProvider) SendBulkEmail(emails []*models.Email) error {
	for _, email := range emails {
		if err := p.SendEmail(email); err != nil {
			return fmt.Errorf("failed to send bulk email: %w", err)
		}
		// Add small delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// GetDeliveryStatus retrieves the delivery status of an email
func (p *OutlookProvider) GetDeliveryStatus(emailID string) (DeliveryStatus, error) {
	// Microsoft Graph API doesn't provide direct delivery status
	// In a real implementation, you might need to use webhooks or polling
	// For now, we'll return a default status
	return DeliveryStatusDelivered, nil
}

// GetBounceList retrieves list of bounced email addresses
func (p *OutlookProvider) GetBounceList() ([]string, error) {
	// Microsoft Graph API doesn't provide direct bounce list
	// In a real implementation, you might need to use webhooks or other methods
	return []string{}, nil
}

// GetComplaintList retrieves list of email addresses that complained
func (p *OutlookProvider) GetComplaintList() ([]string, error) {
	// Microsoft Graph API doesn't provide direct complaint list
	// In a real implementation, you might need to use webhooks or other methods
	return []string{}, nil
}

// MockEmailProvider implements EmailProvider for testing and development
type MockEmailProvider struct {
	senderEmail string
	senderName  string
}

// NewMockEmailProvider creates a new mock email provider
func NewMockEmailProvider(senderEmail, senderName string) *MockEmailProvider {
	return &MockEmailProvider{
		senderEmail: senderEmail,
		senderName:  senderName,
	}
}

// SendEmail sends a single email (mock implementation)
func (p *MockEmailProvider) SendEmail(email *models.Email) error {
	// Mock implementation - in production this would send via Microsoft Graph API
	fmt.Printf("MOCK: Sending email to %s with subject: %s\n", email.Recipients[0].Email, email.Subject)
	return nil
}

// SendBulkEmail sends multiple emails in batch (mock implementation)
func (p *MockEmailProvider) SendBulkEmail(emails []*models.Email) error {
	for _, email := range emails {
		if err := p.SendEmail(email); err != nil {
			return fmt.Errorf("failed to send bulk email: %w", err)
		}
		// Add small delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// GetDeliveryStatus retrieves the delivery status of an email (mock implementation)
func (p *MockEmailProvider) GetDeliveryStatus(emailID string) (DeliveryStatus, error) {
	// Mock implementation - always return delivered
	return DeliveryStatusDelivered, nil
}

// GetBounceList retrieves list of bounced email addresses (mock implementation)
func (p *MockEmailProvider) GetBounceList() ([]string, error) {
	// Mock implementation - return empty list
	return []string{}, nil
}

// GetComplaintList retrieves list of email addresses that complained (mock implementation)
func (p *MockEmailProvider) GetComplaintList() ([]string, error) {
	// Mock implementation - return empty list
	return []string{}, nil
}
