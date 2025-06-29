package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// StockCheckResponse represents the response from inventory service stock check
type StockCheckResponse struct {
	ProductID     uuid.UUID `json:"product_id"`
	Available     int       `json:"available"`
	Reserved      int       `json:"reserved"`
	IsInStock     bool      `json:"is_in_stock"`
	CanFulfill    bool      `json:"can_fulfill"`
	RequestedQty  int       `json:"requested_qty"`
}

// ProductResponse represents the response from inventory service product info
type ProductResponse struct {
	ID          uuid.UUID `json:"id"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	IsActive    bool      `json:"is_active"`
}

// InventoryClient interface for communicating with inventory service
type InventoryClient interface {
	CheckStock(ctx context.Context, productID uuid.UUID, qty int) (*StockCheckResponse, error)
	GetProduct(ctx context.Context, productID uuid.UUID) (*ProductResponse, error)
}

// HTTPInventoryClient implements InventoryClient using HTTP requests
type HTTPInventoryClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPInventoryClient creates a new HTTP inventory client
func NewHTTPInventoryClient(baseURL string) *HTTPInventoryClient {
	return &HTTPInventoryClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewInventoryClient creates a new inventory client using service name from PROJECT_RULES.md
func NewInventoryClient() *HTTPInventoryClient {
	// Use service name as per PROJECT_RULES.md - inventory-service:8082
	return NewHTTPInventoryClient("http://inventory-service:8082")
}

// CheckStock checks if the product has enough stock for the requested quantity
func (c *HTTPInventoryClient) CheckStock(ctx context.Context, productID uuid.UUID, qty int) (*StockCheckResponse, error) {
	// Construct URL: GET /api/inventory/check?product_id={id}&qty={qty}
	url := fmt.Sprintf("%s/api/inventory/check?product_id=%s&qty=%d", c.baseURL, productID.String(), qty)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "order-service/1.0")
	
	// Execute request with retry logic
	resp, err := c.executeWithRetry(req, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to execute stock check request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("inventory service returned status %d", resp.StatusCode)
	}
	
	var stockResponse StockCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&stockResponse); err != nil {
		return nil, fmt.Errorf("failed to decode stock check response: %w", err)
	}
	
	return &stockResponse, nil
}

// GetProduct retrieves product information from inventory service
func (c *HTTPInventoryClient) GetProduct(ctx context.Context, productID uuid.UUID) (*ProductResponse, error) {
	// Construct URL: GET /api/products/{id}
	url := fmt.Sprintf("%s/api/products/%s", c.baseURL, productID.String())
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "order-service/1.0")
	
	// Execute request with retry logic
	resp, err := c.executeWithRetry(req, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get product request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("product not found")
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("inventory service returned status %d", resp.StatusCode)
	}
	
	var productResponse ProductResponse
	if err := json.NewDecoder(resp.Body).Decode(&productResponse); err != nil {
		return nil, fmt.Errorf("failed to decode product response: %w", err)
	}
	
	return &productResponse, nil
}

// executeWithRetry executes HTTP request with retry logic
func (c *HTTPInventoryClient) executeWithRetry(req *http.Request, maxRetries int) (*http.Response, error) {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := c.client.Do(req)
		if err == nil {
			// Check if we should retry based on status code
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return resp, nil
			}
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				// Client errors - don't retry
				return resp, nil
			}
			// Server errors - retry
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: status %d", resp.StatusCode)
		} else {
			lastErr = err
		}
		
		if attempt < maxRetries {
			// Exponential backoff: 100ms, 200ms, 400ms
			backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
			time.Sleep(backoff)
		}
	}
	
	return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}
