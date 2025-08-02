package email

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TemplateEngine interface defines the contract for template engines
type TemplateEngine interface {
	RenderTemplate(templateName string, data map[string]interface{}) (string, string, error)
	GetTemplateList() []string
	ReloadTemplates() error
}

// HTMLTemplateEngine implements TemplateEngine using Go's html/template
type HTMLTemplateEngine struct {
	templates *template.Template
	basePath  string
}

// NewHTMLTemplateEngine creates a new HTML template engine
func NewHTMLTemplateEngine(basePath string) *HTMLTemplateEngine {
	return &HTMLTemplateEngine{
		basePath: basePath,
	}
}

// ReloadTemplates reloads all templates from the base path
func (e *HTMLTemplateEngine) ReloadTemplates() error {
	// Create a new template set
	tmpl := template.New("email_templates")

	// Walk through the templates directory
	err := filepath.Walk(e.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-HTML files
		if info.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}

		// Read template file
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", path, err)
		}

		// Get template name from file path
		relPath, err := filepath.Rel(e.basePath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		templateName := strings.TrimSuffix(relPath, ".html")

		// Parse template
		_, err = tmpl.New(templateName).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", templateName, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to reload templates: %w", err)
	}

	e.templates = tmpl
	return nil
}

// RenderTemplate renders an email template with the given data
func (e *HTMLTemplateEngine) RenderTemplate(templateName string, data map[string]interface{}) (string, string, error) {
	if e.templates == nil {
		if err := e.ReloadTemplates(); err != nil {
			return "", "", fmt.Errorf("failed to load templates: %w", err)
		}
	}

	// Execute HTML template
	var htmlBuffer bytes.Buffer
	if err := e.templates.ExecuteTemplate(&htmlBuffer, templateName, data); err != nil {
		return "", "", fmt.Errorf("failed to render HTML template %s: %w", templateName, err)
	}

	// For now, we'll use the same content for text version
	// In a production system, you might want to create separate text templates
	textContent := htmlBuffer.String()
	// Simple HTML to text conversion (basic implementation)
	textContent = strings.ReplaceAll(textContent, "<br>", "\n")
	textContent = strings.ReplaceAll(textContent, "<br/>", "\n")
	textContent = strings.ReplaceAll(textContent, "<br />", "\n")
	textContent = strings.ReplaceAll(textContent, "<p>", "\n")
	textContent = strings.ReplaceAll(textContent, "</p>", "\n")
	textContent = strings.ReplaceAll(textContent, "<h1>", "\n")
	textContent = strings.ReplaceAll(textContent, "</h1>", "\n")
	textContent = strings.ReplaceAll(textContent, "<h2>", "\n")
	textContent = strings.ReplaceAll(textContent, "</h2>", "\n")
	textContent = strings.ReplaceAll(textContent, "<h3>", "\n")
	textContent = strings.ReplaceAll(textContent, "</h3>", "\n")
	textContent = strings.ReplaceAll(textContent, "<div>", "")
	textContent = strings.ReplaceAll(textContent, "</div>", "\n")
	textContent = strings.ReplaceAll(textContent, "<span>", "")
	textContent = strings.ReplaceAll(textContent, "</span>", "")
	textContent = strings.ReplaceAll(textContent, "<strong>", "")
	textContent = strings.ReplaceAll(textContent, "</strong>", "")
	textContent = strings.ReplaceAll(textContent, "<b>", "")
	textContent = strings.ReplaceAll(textContent, "</b>", "")
	textContent = strings.ReplaceAll(textContent, "<em>", "")
	textContent = strings.ReplaceAll(textContent, "</em>", "")
	textContent = strings.ReplaceAll(textContent, "<i>", "")
	textContent = strings.ReplaceAll(textContent, "</i>", "")
	textContent = strings.ReplaceAll(textContent, "<a href=\"", "")
	textContent = strings.ReplaceAll(textContent, "\">", " ")
	textContent = strings.ReplaceAll(textContent, "</a>", "")
	textContent = strings.ReplaceAll(textContent, "&nbsp;", " ")
	textContent = strings.ReplaceAll(textContent, "&amp;", "&")
	textContent = strings.ReplaceAll(textContent, "&lt;", "<")
	textContent = strings.ReplaceAll(textContent, "&gt;", ">")
	textContent = strings.ReplaceAll(textContent, "&quot;", "\"")
	textContent = strings.ReplaceAll(textContent, "&#39;", "'")

	// Clean up extra whitespace
	textContent = strings.ReplaceAll(textContent, "\n\n\n", "\n\n")
	textContent = strings.TrimSpace(textContent)

	return htmlBuffer.String(), textContent, nil
}

// GetTemplateList returns a list of available templates
func (e *HTMLTemplateEngine) GetTemplateList() []string {
	if e.templates == nil {
		if err := e.ReloadTemplates(); err != nil {
			return []string{}
		}
	}

	var templates []string
	for _, tmpl := range e.templates.Templates() {
		if tmpl.Name() != "email_templates" {
			templates = append(templates, tmpl.Name())
		}
	}

	return templates
}

// Template data structures as specified in the design document

// PasswordResetData represents data for password reset emails
type PasswordResetData struct {
	UserName   string `json:"user_name"`
	ResetLink  string `json:"reset_link"`
	ExpiryTime int    `json:"expiry_time"`
}

// OrderConfirmationData represents data for order confirmation emails
type OrderConfirmationData struct {
	OrderNumber     string      `json:"order_number"`
	OrderDate       time.Time   `json:"order_date"`
	TotalAmount     float64     `json:"total_amount"`
	Items           interface{} `json:"items"`
	ShippingAddress interface{} `json:"shipping_address"`
}

// WelcomeData represents data for welcome emails
type WelcomeData struct {
	UserName       string `json:"user_name"`
	ActivationLink string `json:"activation_link"`
}

// OrderStatusUpdateData represents data for order status update emails
type OrderStatusUpdateData struct {
	OrderNumber    string    `json:"order_number"`
	NewStatus      string    `json:"new_status"`
	TrackingNumber string    `json:"tracking_number"`
	UpdateDate     time.Time `json:"update_date"`
}

// PaymentSuccessData represents data for payment success emails
type PaymentSuccessData struct {
	OrderNumber   string    `json:"order_number"`
	PaymentAmount float64   `json:"payment_amount"`
	PaymentMethod string    `json:"payment_method"`
	PaymentDate   time.Time `json:"payment_date"`
	TransactionID string    `json:"transaction_id"`
}

// PaymentFailedData represents data for payment failed emails
type PaymentFailedData struct {
	OrderNumber   string  `json:"order_number"`
	PaymentAmount float64 `json:"payment_amount"`
	PaymentMethod string  `json:"payment_method"`
	FailureReason string  `json:"failure_reason"`
	RetryLink     string  `json:"retry_link"`
}

// SecurityAlertData represents data for security alert emails
type SecurityAlertData struct {
	UserName       string    `json:"user_name"`
	AlertType      string    `json:"alert_type"`
	AlertDate      time.Time `json:"alert_date"`
	ActionRequired string    `json:"action_required"`
}

// AdminNotificationData represents data for admin notification emails
type AdminNotificationData struct {
	NotificationType string                 `json:"notification_type"`
	Subject          string                 `json:"subject"`
	Message          string                 `json:"message"`
	Data             map[string]interface{} `json:"data"`
	Priority         string                 `json:"priority"`
}

// PromotionalData represents data for promotional emails
type PromotionalData struct {
	CampaignName    string        `json:"campaign_name"`
	Subject         string        `json:"subject"`
	Content         string        `json:"content"`
	Offers          []interface{} `json:"offers"`
	ExpiryDate      *time.Time    `json:"expiry_date"`
	UnsubscribeLink string        `json:"unsubscribe_link"`
}

// CartRecoveryData represents data for cart recovery emails
type CartRecoveryData struct {
	UserName        string      `json:"user_name"`
	CartItems       interface{} `json:"cart_items"`
	DiscountCode    string      `json:"discount_code"`
	DiscountPercent float64     `json:"discount_percent"`
	ExpiryTime      int         `json:"expiry_time"`
}

// ReEngagementData represents data for re-engagement emails
type ReEngagementData struct {
	UserName           string        `json:"user_name"`
	LastPurchase       *time.Time    `json:"last_purchase"`
	PersonalizedOffers []interface{} `json:"personalized_offers"`
	ReEngagementLink   string        `json:"re_engagement_link"`
}
