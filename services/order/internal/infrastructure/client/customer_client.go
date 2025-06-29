package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// CustomerResponse represents customer information from user service
type CustomerResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Phone       string    `json:"phone"`
	IsActive    bool      `json:"is_active"`
	MemberLevel string    `json:"member_level"`
	CreatedAt   time.Time `json:"created_at"`
}

// CustomerAddressResponse represents customer address information
type CustomerAddressResponse struct {
	ID           uuid.UUID `json:"id"`
	CustomerID   uuid.UUID `json:"customer_id"`
	Type         string    `json:"type"` // shipping, billing, default
	AddressLine1 string    `json:"address_line1"`
	AddressLine2 string    `json:"address_line2"`
	City         string    `json:"city"`
	Province     string    `json:"province"`
	PostalCode   string    `json:"postal_code"`
	Country      string    `json:"country"`
	IsDefault    bool      `json:"is_default"`
}

// CustomerClient interface for communicating with user service
type CustomerClient interface {
	GetCustomer(ctx context.Context, customerID uuid.UUID) (*CustomerResponse, error)
	GetCustomerAddresses(ctx context.Context, customerID uuid.UUID) ([]CustomerAddressResponse, error)
	ValidateCustomer(ctx context.Context, customerID uuid.UUID) (bool, error)
}

// HTTPCustomerClient implements CustomerClient using HTTP requests
type HTTPCustomerClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPCustomerClient creates a new HTTP customer client
func NewHTTPCustomerClient(baseURL string) *HTTPCustomerClient {
	return &HTTPCustomerClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewCustomerClient creates a new customer client using service name from PROJECT_RULES.md
func NewCustomerClient() *HTTPCustomerClient {
	// Use service name as per PROJECT_RULES.md - user-service:8088
	return NewHTTPCustomerClient("http://user-service:8088")
}

// GetCustomer retrieves customer information by ID
func (c *HTTPCustomerClient) GetCustomer(ctx context.Context, customerID uuid.UUID) (*CustomerResponse, error) {
	url := fmt.Sprintf("%s/api/customers/%s", c.baseURL, customerID.String())
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "order-service/1.0")
	
	resp, err := c.executeWithRetry(req, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get customer request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("customer not found")
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}
	
	var customerResponse CustomerResponse
	if err := json.NewDecoder(resp.Body).Decode(&customerResponse); err != nil {
		return nil, fmt.Errorf("failed to decode customer response: %w", err)
	}
	
	return &customerResponse, nil
}

// GetCustomerAddresses retrieves all addresses for a customer
func (c *HTTPCustomerClient) GetCustomerAddresses(ctx context.Context, customerID uuid.UUID) ([]CustomerAddressResponse, error) {
	url := fmt.Sprintf("%s/api/customers/%s/addresses", c.baseURL, customerID.String())
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "order-service/1.0")
	
	resp, err := c.executeWithRetry(req, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get addresses request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}
	
	var addresses []CustomerAddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&addresses); err != nil {
		return nil, fmt.Errorf("failed to decode addresses response: %w", err)
	}
	
	return addresses, nil
}

// ValidateCustomer checks if customer exists and is active
func (c *HTTPCustomerClient) ValidateCustomer(ctx context.Context, customerID uuid.UUID) (bool, error) {
	customer, err := c.GetCustomer(ctx, customerID)
	if err != nil {
		if err.Error() == "customer not found" {
			return false, nil
		}
		return false, err
	}
	
	return customer.IsActive, nil
}

// executeWithRetry executes HTTP request with retry logic (same as inventory client)
func (c *HTTPCustomerClient) executeWithRetry(req *http.Request, maxRetries int) (*http.Response, error) {
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
