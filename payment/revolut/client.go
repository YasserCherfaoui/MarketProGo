package revolut

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/cfg"
)

// Client represents a Revolut API client
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	merchantID string
}

// NewClient creates a new Revolut API client
func NewClient(config *cfg.RevolutConfig) *Client {
	// Create HTTP client with proper timeouts
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		},
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    config.BaseURL,
		apiKey:     config.APIKey,
		merchantID: config.MerchantID,
	}
}

// Customer represents customer information for an order
type Customer struct {
	ID       string `json:"id,omitempty"`
	FullName string `json:"full_name,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`
}

// LineItem represents a line item in an order
type LineItem struct {
	Name            string   `json:"name"`
	Type            string   `json:"type"` // "physical" or "service"
	Quantity        Quantity `json:"quantity"`
	UnitPriceAmount int64    `json:"unit_price_amount"`
	TotalAmount     int64    `json:"total_amount"`
	ExternalID      string   `json:"external_id,omitempty"`
	Description     string   `json:"description,omitempty"`
	URL             string   `json:"url,omitempty"`
	ImageURLs       []string `json:"image_urls,omitempty"`
}

// Quantity represents the quantity of a line item
type Quantity struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit,omitempty"`
}

// OrderRequest represents a request to create an order according to Revolut Merchant API
type OrderRequest struct {
	Amount                    int64             `json:"amount"`
	Currency                  string            `json:"currency"`
	SettlementCurrency        string            `json:"settlement_currency,omitempty"`
	Description               string            `json:"description,omitempty"`
	Customer                  *Customer         `json:"customer,omitempty"`
	EnforceChallenge          string            `json:"enforce_challenge,omitempty"` // "automatic" or "forced"
	LineItems                 []LineItem        `json:"line_items,omitempty"`
	Shipping                  interface{}       `json:"shipping,omitempty"`
	CaptureMode               string            `json:"capture_mode,omitempty"` // "automatic" or "manual"
	CancelAuthorisedAfter     string            `json:"cancel_authorised_after,omitempty"`
	LocationID                string            `json:"location_id,omitempty"`
	Metadata                  map[string]string `json:"metadata,omitempty"`
	IndustryData              interface{}       `json:"industry_data,omitempty"`
	MerchantOrderData         interface{}       `json:"merchant_order_data,omitempty"`
	UpcomingPaymentData       interface{}       `json:"upcoming_payment_data,omitempty"`
	RedirectURL               string            `json:"redirect_url,omitempty"`
	StatementDescriptorSuffix string            `json:"statement_descriptor_suffix,omitempty"`
}

// OrderResponse represents a response from creating an order
type OrderResponse struct {
	ID                        string            `json:"id"`
	Token                     string            `json:"token"`
	Type                      string            `json:"type"`
	State                     string            `json:"state"`
	CreatedAt                 string            `json:"created_at"`
	UpdatedAt                 string            `json:"updated_at"`
	Amount                    int64             `json:"amount"`
	Currency                  string            `json:"currency"`
	OutstandingAmount         int64             `json:"outstanding_amount"`
	CaptureMode               string            `json:"capture_mode"`
	CheckoutURL               string            `json:"checkout_url"`
	EnforceChallenge          string            `json:"enforce_challenge"`
	Description               string            `json:"description,omitempty"`
	Customer                  *Customer         `json:"customer,omitempty"`
	LineItems                 []LineItem        `json:"line_items,omitempty"`
	Shipping                  interface{}       `json:"shipping,omitempty"`
	Metadata                  map[string]string `json:"metadata,omitempty"`
	IndustryData              interface{}       `json:"industry_data,omitempty"`
	MerchantOrderData         interface{}       `json:"merchant_order_data,omitempty"`
	UpcomingPaymentData       interface{}       `json:"upcoming_payment_data,omitempty"`
	RedirectURL               string            `json:"redirect_url,omitempty"`
	StatementDescriptorSuffix string            `json:"statement_descriptor_suffix,omitempty"`
}

// RefundRequest represents a request to refund a payment
type RefundRequest struct {
	Amount   float64           `json:"amount"`
	Currency string            `json:"currency"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Reason   string            `json:"reason,omitempty"`
}

// RefundResponse represents a response from refunding a payment
type RefundResponse struct {
	ID        string            `json:"id"`
	Amount    float64           `json:"amount"`
	Currency  string            `json:"currency"`
	State     string            `json:"state"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

// CaptureResponse represents a response from capturing a payment
type CaptureResponse struct {
	ID        string `json:"id"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ErrorResponse represents an error response from the Revolut API
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CreateOrder creates a new order using the Revolut Merchant API
func (c *Client) CreateOrder(req *OrderRequest) (*OrderResponse, error) {
	url := fmt.Sprintf("%s/api/orders", c.baseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers according to Revolut API documentation
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Revolut-Api-Version", "2024-09-01") // Use stable API version

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			// Log the raw response for debugging
			log.Printf("Failed to unmarshal error response. Status: %d, Body: %s", resp.StatusCode, string(body))
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
		// Log detailed error information
		log.Printf("Revolut API error - Status: %d, Code: %s, Message: %s", resp.StatusCode, errorResp.Code, errorResp.Message)
		return nil, fmt.Errorf("API request failed: %s - %s", errorResp.Code, errorResp.Message)
	}

	var orderResp OrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order response: %w", err)
	}

	return &orderResp, nil
}

// GetOrder retrieves an order by ID
func (c *Client) GetOrder(orderID string) (*OrderResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderID)

	httpReq, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Revolut-Api-Version", "2023-09-01") // Use stable API version

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API request failed: %s - %s", errorResp.Code, errorResp.Message)
	}

	var orderResp OrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order response: %w", err)
	}

	return &orderResp, nil
}

// RefundPayment refunds a payment
func (c *Client) RefundPayment(paymentID string, req *RefundRequest) (*RefundResponse, error) {
	url := fmt.Sprintf("%s/api/1.0/payments/%s/refund", c.baseURL, paymentID)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal refund request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API request failed: %s - %s", errorResp.Code, errorResp.Message)
	}

	var refundResp RefundResponse
	if err := json.Unmarshal(body, &refundResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal refund response: %w", err)
	}

	return &refundResp, nil
}

// CaptureOrder captures an order
func (c *Client) CaptureOrder(paymentID string) (*CaptureResponse, error) {
	url := fmt.Sprintf("%s/api/1.0/orders/%s/capture", c.baseURL, paymentID)

	httpReq, err := http.NewRequestWithContext(context.Background(), "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API request failed: %s - %s", errorResp.Code, errorResp.Message)
	}

	var captureResp CaptureResponse
	if err := json.Unmarshal(body, &captureResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capture response: %w", err)
	}

	return &captureResp, nil
}
