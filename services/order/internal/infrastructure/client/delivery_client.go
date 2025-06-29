package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// DeliveryQuoteRequest represents request for delivery quote
type DeliveryQuoteRequest struct {
	Origin      AddressInfo `json:"origin"`
	Destination AddressInfo `json:"destination"`
	Weight      float64     `json:"weight"`      // in kg
	Dimensions  Dimensions  `json:"dimensions"`  // in cm
	ServiceType string      `json:"service_type"` // standard, express, same_day
}

// DeliveryQuoteResponse represents delivery quote response
type DeliveryQuoteResponse struct {
	QuoteID          uuid.UUID `json:"quote_id"`
	ServiceType      string    `json:"service_type"`
	EstimatedCost    float64   `json:"estimated_cost"`
	EstimatedDays    int       `json:"estimated_days"`
	Carrier          string    `json:"carrier"`
	IsAvailable      bool      `json:"is_available"`
	ValidUntil       time.Time `json:"valid_until"`
}

// CreateDeliveryRequest represents request to create delivery order
type CreateDeliveryRequest struct {
	OrderID     uuid.UUID   `json:"order_id"`
	QuoteID     uuid.UUID   `json:"quote_id"`
	Origin      AddressInfo `json:"origin"`
	Destination AddressInfo `json:"destination"`
	Weight      float64     `json:"weight"`
	Dimensions  Dimensions  `json:"dimensions"`
	ServiceType string      `json:"service_type"`
	Instructions string     `json:"instructions,omitempty"`
}

// DeliveryResponse represents delivery order response
type DeliveryResponse struct {
	ID              uuid.UUID `json:"id"`
	OrderID         uuid.UUID `json:"order_id"`
	TrackingNumber  string    `json:"tracking_number"`
	Status          string    `json:"status"`
	Carrier         string    `json:"carrier"`
	ServiceType     string    `json:"service_type"`
	EstimatedCost   float64   `json:"estimated_cost"`
	ActualCost      float64   `json:"actual_cost,omitempty"`
	EstimatedDelivery time.Time `json:"estimated_delivery"`
	CreatedAt       time.Time `json:"created_at"`
}

// AddressInfo represents address information for delivery
type AddressInfo struct {
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2,omitempty"`
	City         string `json:"city"`
	Province     string `json:"province"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
}

// Dimensions represents package dimensions
type Dimensions struct {
	Length float64 `json:"length"` // cm
	Width  float64 `json:"width"`  // cm
	Height float64 `json:"height"` // cm
}

// DeliveryClient interface for communicating with shipping service
type DeliveryClient interface {
	GetQuote(ctx context.Context, req *DeliveryQuoteRequest) (*DeliveryQuoteResponse, error)
	CreateDelivery(ctx context.Context, req *CreateDeliveryRequest) (*DeliveryResponse, error)
	GetDeliveryStatus(ctx context.Context, deliveryID uuid.UUID) (*DeliveryResponse, error)
	TrackDelivery(ctx context.Context, trackingNumber string) (*DeliveryResponse, error)
}

// HTTPDeliveryClient implements DeliveryClient using HTTP requests
type HTTPDeliveryClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPDeliveryClient creates a new HTTP delivery client
func NewHTTPDeliveryClient(baseURL string) *HTTPDeliveryClient {
	return &HTTPDeliveryClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewDeliveryClient creates a new delivery client using service name from PROJECT_RULES.md
func NewDeliveryClient() *HTTPDeliveryClient {
	// Use service name as per PROJECT_RULES.md - shipping-service:8086
	return NewHTTPDeliveryClient("http://shipping-service:8086")
}

// GetQuote gets delivery quote from shipping service
func (c *HTTPDeliveryClient) GetQuote(ctx context.Context, req *DeliveryQuoteRequest) (*DeliveryQuoteResponse, error) {
	url := fmt.Sprintf("%s/api/delivery/quote", c.baseURL)
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "order-service/1.0")
	
	resp, err := c.executeWithRetry(httpReq, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to execute quote request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shipping service returned status %d", resp.StatusCode)
	}
	
	var quoteResponse DeliveryQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quoteResponse); err != nil {
		return nil, fmt.Errorf("failed to decode quote response: %w", err)
	}
	
	return &quoteResponse, nil
}

// CreateDelivery creates a delivery order
func (c *HTTPDeliveryClient) CreateDelivery(ctx context.Context, req *CreateDeliveryRequest) (*DeliveryResponse, error) {
	url := fmt.Sprintf("%s/api/delivery", c.baseURL)
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "order-service/1.0")
	
	resp, err := c.executeWithRetry(httpReq, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to execute create delivery request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("shipping service returned status %d", resp.StatusCode)
	}
	
	var deliveryResponse DeliveryResponse
	if err := json.NewDecoder(resp.Body).Decode(&deliveryResponse); err != nil {
		return nil, fmt.Errorf("failed to decode delivery response: %w", err)
	}
	
	return &deliveryResponse, nil
}

// GetDeliveryStatus gets delivery status by delivery ID
func (c *HTTPDeliveryClient) GetDeliveryStatus(ctx context.Context, deliveryID uuid.UUID) (*DeliveryResponse, error) {
	url := fmt.Sprintf("%s/api/delivery/%s", c.baseURL, deliveryID.String())
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "order-service/1.0")
	
	resp, err := c.executeWithRetry(req, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get delivery request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("delivery not found")
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shipping service returned status %d", resp.StatusCode)
	}
	
	var deliveryResponse DeliveryResponse
	if err := json.NewDecoder(resp.Body).Decode(&deliveryResponse); err != nil {
		return nil, fmt.Errorf("failed to decode delivery response: %w", err)
	}
	
	return &deliveryResponse, nil
}

// TrackDelivery tracks delivery by tracking number
func (c *HTTPDeliveryClient) TrackDelivery(ctx context.Context, trackingNumber string) (*DeliveryResponse, error) {
	url := fmt.Sprintf("%s/api/delivery/track/%s", c.baseURL, trackingNumber)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "order-service/1.0")
	
	resp, err := c.executeWithRetry(req, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to execute track delivery request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("tracking number not found")
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shipping service returned status %d", resp.StatusCode)
	}
	
	var deliveryResponse DeliveryResponse
	if err := json.NewDecoder(resp.Body).Decode(&deliveryResponse); err != nil {
		return nil, fmt.Errorf("failed to decode delivery response: %w", err)
	}
	
	return &deliveryResponse, nil
}

// executeWithRetry executes HTTP request with retry logic
func (c *HTTPDeliveryClient) executeWithRetry(req *http.Request, maxRetries int) (*http.Response, error) {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := c.client.Do(req)
		if err == nil {
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return resp, nil
			}
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return resp, nil
			}
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: status %d", resp.StatusCode)
		} else {
			lastErr = err
		}
		
		if attempt < maxRetries {
			backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
			time.Sleep(backoff)
		}
	}
	
	return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}
