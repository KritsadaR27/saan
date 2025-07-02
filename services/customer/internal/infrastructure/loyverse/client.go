package loyverse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/saan-system/services/customer/internal/domain"
)

// Client represents a Loyverse API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// LoyverseCustomer represents a customer in Loyverse format
type LoyverseCustomer struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Note        string `json:"note,omitempty"`
}

// NewClient creates a new Loyverse API client
func NewClient() domain.LoyverseClient {
	return &Client{
		baseURL: getEnv("LOYVERSE_API_URL", "https://api.loyverse.com/v1.0"),
		apiKey:  getEnv("LOYVERSE_API_KEY", ""),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateCustomer creates a customer in Loyverse
func (c *Client) CreateCustomer(ctx context.Context, customer *domain.Customer) (*string, error) {
	if c.apiKey == "" {
		return nil, domain.ErrLoyverseAPIError
	}

	loyverseCustomer := LoyverseCustomer{
		Name:        customer.GetFullName(),
		Email:       customer.Email,
		PhoneNumber: customer.Phone,
		Note:        fmt.Sprintf("Customer ID: %s, Tier: %s", customer.ID.String(), string(customer.Tier)),
	}

	data, err := json.Marshal(loyverseCustomer)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal customer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/customers", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create customer in Loyverse: status %d", resp.StatusCode)
	}

	var createdCustomer LoyverseCustomer
	if err := json.NewDecoder(resp.Body).Decode(&createdCustomer); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdCustomer.ID, nil
}

// GetCustomer retrieves a customer from Loyverse
func (c *Client) GetCustomer(ctx context.Context, loyverseID string) (*domain.Customer, error) {
	if c.apiKey == "" {
		return nil, domain.ErrLoyverseAPIError
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/customers/"+loyverseID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, domain.ErrLoyverseCustomerNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get customer from Loyverse: status %d", resp.StatusCode)
	}

	var loyverseCustomer LoyverseCustomer
	if err := json.NewDecoder(resp.Body).Decode(&loyverseCustomer); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert Loyverse customer to domain customer (basic mapping)
	customer := &domain.Customer{
		Email: loyverseCustomer.Email,
		Phone: loyverseCustomer.PhoneNumber,
		LoyverseID: &loyverseCustomer.ID,
	}

	return customer, nil
}

// UpdateCustomer updates a customer in Loyverse
func (c *Client) UpdateCustomer(ctx context.Context, loyverseID string, customer *domain.Customer) error {
	if c.apiKey == "" {
		return domain.ErrLoyverseAPIError
	}

	loyverseCustomer := LoyverseCustomer{
		Name:        customer.GetFullName(),
		Email:       customer.Email,
		PhoneNumber: customer.Phone,
		Note:        fmt.Sprintf("Customer ID: %s, Tier: %s", customer.ID.String(), string(customer.Tier)),
	}

	data, err := json.Marshal(loyverseCustomer)
	if err != nil {
		return fmt.Errorf("failed to marshal customer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", c.baseURL+"/customers/"+loyverseID, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return domain.ErrLoyverseCustomerNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update customer in Loyverse: status %d", resp.StatusCode)
	}

	return nil
}

// SearchCustomerByEmail searches for a customer by email in Loyverse
func (c *Client) SearchCustomerByEmail(ctx context.Context, email string) (*string, error) {
	if c.apiKey == "" {
		return nil, domain.ErrLoyverseAPIError
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/customers?email="+email, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to search customer in Loyverse: status %d", resp.StatusCode)
	}

	var customers []LoyverseCustomer
	if err := json.NewDecoder(resp.Body).Decode(&customers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(customers) == 0 {
		return nil, domain.ErrLoyverseCustomerNotFound
	}

	return &customers[0].ID, nil
}

// SearchCustomerByPhone searches for a customer by phone in Loyverse
func (c *Client) SearchCustomerByPhone(ctx context.Context, phone string) (*string, error) {
	if c.apiKey == "" {
		return nil, domain.ErrLoyverseAPIError
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/customers?phone="+phone, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to search customer in Loyverse: status %d", resp.StatusCode)
	}

	var customers []LoyverseCustomer
	if err := json.NewDecoder(resp.Body).Decode(&customers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(customers) == 0 {
		return nil, domain.ErrLoyverseCustomerNotFound
	}

	return &customers[0].ID, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
