package revolut

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// OrderRequest represents a request to create a payment order
type OrderRequest struct {
	Amount          float64           `json:"amount"`
	Currency        string            `json:"currency"`
	MerchantOrderID string            `json:"merchant_order_id"`
	CustomerEmail   string            `json:"customer_email"`
	CustomerName    string            `json:"customer_name"`
	Description     string            `json:"description"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	CaptureMode     string            `json:"capture_mode,omitempty"` // MANUAL or AUTOMATIC
	PaymentMethods  []string          `json:"payment_methods,omitempty"`
	ReturnURL       string            `json:"return_url,omitempty"`
	CancelURL       string            `json:"cancel_url,omitempty"`
}

// OrderResponse represents a response from creating a payment order
type OrderResponse struct {
	ID              string            `json:"id"`
	PublicID        string            `json:"public_id"`
	Amount          float64           `json:"amount"`
	Currency        string            `json:"currency"`
	State           string            `json:"state"`
	MerchantOrderID string            `json:"merchant_order_id"`
	CustomerEmail   string            `json:"customer_email"`
	CustomerName    string            `json:"customer_name"`
	Description     string            `json:"description"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	CreatedAt       string            `json:"created_at"`
	UpdatedAt       string            `json:"updated_at"`
	CheckoutURL     string            `json:"checkout_url,omitempty"`
	PaymentMethods  []string          `json:"payment_methods,omitempty"`
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

// CreateOrder creates a new payment order
func (c *Client) CreateOrder(req *OrderRequest) (*OrderResponse, error) {
	url := fmt.Sprintf("%s/api/1.0/orders", c.baseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order request: %w", err)
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

	var orderResp OrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order response: %w", err)
	}

	return &orderResp, nil
}

// GetOrder retrieves an order by ID
func (c *Client) GetOrder(orderID string) (*OrderResponse, error) {
	url := fmt.Sprintf("%s/api/1.0/orders/%s", c.baseURL, orderID)

	httpReq, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
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
