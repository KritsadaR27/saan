// integrations/loyverse/internal/connector/client.go
package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// Client represents Loyverse API client
type Client struct {
	httpClient  *http.Client
	baseURL     string
	token       string
	rateLimiter *rate.Limiter
}

// NewClient creates a new Loyverse API client
func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:     "https://api.loyverse.com/v1.0",
		token:       token,
		rateLimiter: rate.NewLimiter(rate.Every(time.Second), 10), // 10 requests per second
	}
}

// Request makes an HTTP request to Loyverse API with rate limiting
func (c *Client) Request(ctx context.Context, method, endpoint string, body io.Reader) ([]byte, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetWithPagination handles paginated GET requests
func (c *Client) GetWithPagination(ctx context.Context, endpoint string, limit int) ([]json.RawMessage, error) {
	var allResults []json.RawMessage
	cursor := ""

	for {
		url := fmt.Sprintf("%s?limit=%d", endpoint, limit)
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		body, err := c.Request(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		var response struct {
			Data   []json.RawMessage `json:"data"`
			Cursor string            `json:"cursor"`
		}

		log.Printf("DEBUG: Raw response body: %s", string(body))

		// Try to unmarshal with different field names based on endpoint
		if err := json.Unmarshal(body, &response); err != nil {
			// Try alternative structure for specific endpoints
			var altResponse map[string]json.RawMessage
			if err := json.Unmarshal(body, &altResponse); err != nil {
				return nil, fmt.Errorf("parsing response: %w", err)
			}

			log.Printf("DEBUG: Response structure: %+v", altResponse)

			// Extract data based on endpoint type
			dataFound := false
			for key, value := range altResponse {
				if key == "cursor" {
					json.Unmarshal(value, &cursor)
				} else if key != "cursor" && !dataFound {
					// This is likely our data array - try to unmarshal as array
					var items []json.RawMessage
					if err := json.Unmarshal(value, &items); err == nil {
						log.Printf("DEBUG: Found %d items in key '%s'", len(items), key)
						allResults = append(allResults, items...)
						dataFound = true
						// Look for cursor in the response
						if cursorValue, exists := altResponse["cursor"]; exists {
							json.Unmarshal(cursorValue, &cursor)
						}
					}
				}
			}
		} else {
			log.Printf("DEBUG: Standard response unmarshaled, found %d items", len(response.Data))
			
			// If no data found in standard response, try alternative structure
			if len(response.Data) == 0 {
				var altResponse map[string]json.RawMessage
				if err := json.Unmarshal(body, &altResponse); err == nil {
					log.Printf("DEBUG: Alternative response structure: %+v", altResponse)
					
					// Extract data from alternative keys
					dataFound := false
					for key, value := range altResponse {
						if key == "cursor" {
							json.Unmarshal(value, &cursor)
						} else if key != "cursor" && !dataFound {
							// This is likely our data array - try to unmarshal as array
							var items []json.RawMessage
							if err := json.Unmarshal(value, &items); err == nil {
								log.Printf("DEBUG: Found %d items in alternative key '%s'", len(items), key)
								allResults = append(allResults, items...)
								dataFound = true
								// Look for cursor in the response
								if cursorValue, exists := altResponse["cursor"]; exists {
									json.Unmarshal(cursorValue, &cursor)
								}
							}
						}
					}
				}
			} else {
				allResults = append(allResults, response.Data...)
			}
			
			// Use cursor from standard response if not found in alternative
			if cursor == "" {
				cursor = response.Cursor
			}
		}

		if cursor == "" {
			break
		}
	}

	return allResults, nil
}

// GetReceiptsBatch handles receipts with proper cursor pagination
func (c *Client) GetReceiptsBatch(ctx context.Context, cursor string, limit int) ([]json.RawMessage, error) {
	var allResults []json.RawMessage
	currentCursor := cursor

	for {
		url := fmt.Sprintf("/receipts?limit=%d", limit)
		if currentCursor != "" {
			url += "&cursor=" + currentCursor
		}

		body, err := c.Request(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		log.Printf("DEBUG: Receipts raw response: %s", string(body))

		// Parse receipts response
		var receiptsResponse struct {
			Receipts []json.RawMessage `json:"receipts"`
			Cursor   string            `json:"cursor"`
		}

		if err := json.Unmarshal(body, &receiptsResponse); err != nil {
			log.Printf("ERROR: Failed to parse receipts response: %v", err)
			return nil, fmt.Errorf("parsing receipts response: %w", err)
		}

		log.Printf("DEBUG: Found %d receipts, next cursor: %s", len(receiptsResponse.Receipts), receiptsResponse.Cursor)
		allResults = append(allResults, receiptsResponse.Receipts...)

		if receiptsResponse.Cursor == "" {
			break
		}
		currentCursor = receiptsResponse.Cursor
	}

	return allResults, nil
}

// Product-specific methods
func (c *Client) GetProducts(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/items", 250)
}

func (c *Client) GetInventoryLevels(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/inventory", 250)
}

// GetReceipts retrieves receipts with cursor-based pagination
func (c *Client) GetReceipts(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetReceiptsBatch(ctx, "", 250)
}

// GetRecentReceipts retrieves recent receipts
func (c *Client) GetRecentReceipts(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetReceipts(ctx)
}

func (c *Client) CreateReceipt(ctx context.Context, receipt interface{}) error {
	data, err := json.Marshal(receipt)
	if err != nil {
		return fmt.Errorf("marshaling receipt: %w", err)
	}

	_, err = c.Request(ctx, http.MethodPost, "/receipts", bytes.NewReader(data))
	return err
}

// Additional API methods for testing various endpoints

// GetStores retrieves all stores
func (c *Client) GetStores(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/stores", 250)
}

// GetCustomers retrieves all customers
func (c *Client) GetCustomers(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/customers", 250)
}

// GetEmployees retrieves all employees
func (c *Client) GetEmployees(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/employees", 250)
}

// GetDiscounts retrieves all discounts
func (c *Client) GetDiscounts(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/discounts", 250)
}

// GetModifiers retrieves all modifiers
func (c *Client) GetModifiers(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/modifiers", 250)
}

// GetTaxes retrieves taxes
func (c *Client) GetTaxes(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/taxes", 250)
}

// GetPaymentTypes retrieves payment types
func (c *Client) GetPaymentTypes(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/payment_types", 250)
}

// GetVariants retrieves product variants
func (c *Client) GetVariants(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/variants", 250)
}

// GetInventory retrieves inventory levels
func (c *Client) GetInventory(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/inventory", 250)
}

// GetSuppliers retrieves suppliers
func (c *Client) GetSuppliers(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/suppliers", 250)
}

// GetPurchaseOrders retrieves purchase orders
func (c *Client) GetPurchaseOrders(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/purchase_orders", 250)
}

// GetReceiptsWithParams retrieves receipts with parameters
func (c *Client) GetReceiptsWithParams(ctx context.Context, params map[string]string) ([]json.RawMessage, error) {
	return c.GetReceipts(ctx)
}

// GetPOSDevices retrieves POS devices
func (c *Client) GetPOSDevices(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/pos_devices", 250)
}

// GetCashRegisters retrieves cash registers
func (c *Client) GetCashRegisters(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/cash_registers", 250)
}

// GetWebhooks retrieves webhooks
func (c *Client) GetWebhooks(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/webhooks", 250)
}

// GetCategories retrieves categories
func (c *Client) GetCategories(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetWithPagination(ctx, "/categories", 250)
}

// GetAccount retrieves account information
func (c *Client) GetAccount(ctx context.Context) ([]byte, error) {
	return c.Request(ctx, http.MethodGet, "/account", nil)
}

// GetSettings retrieves settings
func (c *Client) GetSettings(ctx context.Context) ([]byte, error) {
	return c.Request(ctx, http.MethodGet, "/settings", nil)
}

// GetEndpoint is a generic method for any endpoint
func (c *Client) GetEndpoint(ctx context.Context, endpoint string) ([]byte, error) {
	return c.Request(ctx, http.MethodGet, endpoint, nil)
}

// GetAvailableEndpoints returns map of available endpoints
func (c *Client) GetAvailableEndpoints() map[string]string {
	return map[string]string{
		"stores":          "/stores",
		"customers":       "/customers", 
		"employees":       "/employees",
		"discounts":       "/discounts",
		"modifiers":       "/modifiers",
		"taxes":           "/taxes",
		"payment_types":   "/payment_types",
		"variants":        "/variants",
		"suppliers":       "/suppliers",
		"purchase_orders": "/purchase_orders",
		"pos_devices":     "/pos_devices",
		"cash_registers":  "/cash_registers",
		"webhooks":        "/webhooks",
		"categories":      "/categories",
		"items":           "/items",
		"inventory":       "/inventory",
		"receipts":        "/receipts",
		"account":         "/account",
		"settings":        "/settings",
	}
}

// TestAllEndpoints tests all available endpoints
func (c *Client) TestAllEndpoints(ctx context.Context) map[string]interface{} {
	results := make(map[string]interface{})
	endpoints := c.GetAvailableEndpoints()
	
	for name, path := range endpoints {
		log.Printf("Testing endpoint: %s (%s)", name, path)
		
		var count int
		var err error
		
		switch name {
		case "receipts":
			data, testErr := c.GetReceipts(ctx)
			count = len(data)
			err = testErr
		case "variants":
			data, testErr := c.GetVariants(ctx)
			count = len(data)
			err = testErr
		case "inventory":
			data, testErr := c.GetInventory(ctx)
			count = len(data)
			err = testErr
		case "account", "settings":
			_, testErr := c.GetEndpoint(ctx, path)
			if testErr == nil {
				count = 1
			}
			err = testErr
		default:
			data, testErr := c.GetWithPagination(ctx, path, 250)
			count = len(data)
			err = testErr
		}
		
		if err != nil {
			results[name] = map[string]interface{}{
				"success": false,
				"error":   err.Error(),
				"count":   0,
			}
		} else {
			results[name] = map[string]interface{}{
				"success": true,
				"count":   count,
			}
		}
	}
	
	return results
}