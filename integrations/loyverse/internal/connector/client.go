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

// GetReceipts retrieves all receipts with proper pagination
func (c *Client) GetReceipts(ctx context.Context) ([]json.RawMessage, error) {
	var allReceipts []json.RawMessage
	cursor := ""
	limit := 250

	for {
		url := fmt.Sprintf("/receipts?limit=%d", limit)
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		body, err := c.Request(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		var response struct {
			Receipts []json.RawMessage `json:"receipts"`
			Cursor   string            `json:"cursor"`
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("parsing receipts response: %w", err)
		}

		allReceipts = append(allReceipts, response.Receipts...)

		if response.Cursor == "" {
			break
		}
		cursor = response.Cursor
	}

	return allReceipts, nil
}

// GetInventoryLevels retrieves all inventory levels with proper pagination
func (c *Client) GetInventoryLevels(ctx context.Context) ([]json.RawMessage, error) {
	var allInventory []json.RawMessage
	cursor := ""
	limit := 250

	for {
		url := fmt.Sprintf("/inventory?limit=%d", limit)
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		body, err := c.Request(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		var response struct {
			InventoryLevels []json.RawMessage `json:"inventory_levels"`
			Cursor          string            `json:"cursor"`
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("parsing inventory response: %w", err)
		}

		allInventory = append(allInventory, response.InventoryLevels...)

		if response.Cursor == "" {
			break
		}
		cursor = response.Cursor
	}

	return allInventory, nil
}

// GetProducts retrieves all products with proper pagination
func (c *Client) GetProducts(ctx context.Context) ([]json.RawMessage, error) {
	var allItems []json.RawMessage
	cursor := ""
	limit := 250

	for {
		url := fmt.Sprintf("/items?limit=%d", limit)
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		body, err := c.Request(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		var response struct {
			Items  []json.RawMessage `json:"items"`
			Cursor string            `json:"cursor"`
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("parsing items response: %w", err)
		}

		allItems = append(allItems, response.Items...)

		if response.Cursor == "" {
			break
		}
		cursor = response.Cursor
	}

	return allItems, nil
}

// GetCustomers retrieves all customers with proper pagination
func (c *Client) GetCustomers(ctx context.Context) ([]json.RawMessage, error) {
	var allCustomers []json.RawMessage
	cursor := ""
	limit := 250

	for {
		url := fmt.Sprintf("/customers?limit=%d", limit)
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		body, err := c.Request(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		var response struct {
			Customers []json.RawMessage `json:"customers"`
			Cursor    string            `json:"cursor"`
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("parsing customers response: %w", err)
		}

		allCustomers = append(allCustomers, response.Customers...)

		if response.Cursor == "" {
			break
		}
		cursor = response.Cursor
	}

	return allCustomers, nil
}

// GetStores retrieves all stores
func (c *Client) GetStores(ctx context.Context) ([]json.RawMessage, error) {
	var allStores []json.RawMessage
	cursor := ""
	limit := 250

	for {
		url := fmt.Sprintf("/stores?limit=%d", limit)
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		body, err := c.Request(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		var response struct {
			Stores []json.RawMessage `json:"stores"`
			Cursor string            `json:"cursor"`
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("parsing stores response: %w", err)
		}

		allStores = append(allStores, response.Stores...)

		if response.Cursor == "" {
			break
		}
		cursor = response.Cursor
	}

	return allStores, nil
}

// GetEmployees retrieves all employees
func (c *Client) GetEmployees(ctx context.Context) ([]json.RawMessage, error) {
	var allEmployees []json.RawMessage
	cursor := ""
	limit := 250

	for {
		url := fmt.Sprintf("/employees?limit=%d", limit)
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		body, err := c.Request(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		var response struct {
			Employees []json.RawMessage `json:"employees"`
			Cursor    string            `json:"cursor"`
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("parsing employees response: %w", err)
		}

		allEmployees = append(allEmployees, response.Employees...)

		if response.Cursor == "" {
			break
		}
		cursor = response.Cursor
	}

	return allEmployees, nil
}

// CreateReceipt creates a new receipt
func (c *Client) CreateReceipt(ctx context.Context, receipt interface{}) error {
	data, err := json.Marshal(receipt)
	if err != nil {
		return fmt.Errorf("marshaling receipt: %w", err)
	}

	_, err = c.Request(ctx, http.MethodPost, "/receipts", bytes.NewReader(data))
	return err
}

// GetAccount retrieves account information (single object, no pagination)
func (c *Client) GetAccount(ctx context.Context) ([]byte, error) {
	return c.Request(ctx, http.MethodGet, "/account", nil)
}

// GetSettings retrieves settings (single object, no pagination)
func (c *Client) GetSettings(ctx context.Context) ([]byte, error) {
	return c.Request(ctx, http.MethodGet, "/settings", nil)
}

// Generic pagination helper for endpoints with standard structure
func (c *Client) getWithStandardPagination(ctx context.Context, endpoint, dataField string) ([]json.RawMessage, error) {
	var allData []json.RawMessage
	cursor := ""
	limit := 250

	for {
		url := fmt.Sprintf("%s?limit=%d", endpoint, limit)
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		body, err := c.Request(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		// Parse response dynamically based on dataField
		var rawResponse map[string]json.RawMessage
		if err := json.Unmarshal(body, &rawResponse); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		// Extract data array
		if dataRaw, exists := rawResponse[dataField]; exists {
			var data []json.RawMessage
			if err := json.Unmarshal(dataRaw, &data); err != nil {
				return nil, fmt.Errorf("parsing %s array: %w", dataField, err)
			}
			allData = append(allData, data...)
		}

		// Extract cursor
		cursor = ""
		if cursorRaw, exists := rawResponse["cursor"]; exists {
			json.Unmarshal(cursorRaw, &cursor)
		}

		if cursor == "" {
			break
		}
	}

	return allData, nil
}

// Additional endpoints using the generic helper
func (c *Client) GetCategories(ctx context.Context) ([]json.RawMessage, error) {
	return c.getWithStandardPagination(ctx, "/categories", "categories")
}

func (c *Client) GetDiscounts(ctx context.Context) ([]json.RawMessage, error) {
	return c.getWithStandardPagination(ctx, "/discounts", "discounts")
}

func (c *Client) GetModifiers(ctx context.Context) ([]json.RawMessage, error) {
	return c.getWithStandardPagination(ctx, "/modifiers", "modifiers")
}

func (c *Client) GetTaxes(ctx context.Context) ([]json.RawMessage, error) {
	return c.getWithStandardPagination(ctx, "/taxes", "taxes")
}

func (c *Client) GetPaymentTypes(ctx context.Context) ([]json.RawMessage, error) {
	return c.getWithStandardPagination(ctx, "/payment_types", "payment_types")
}

func (c *Client) GetVariants(ctx context.Context) ([]json.RawMessage, error) {
	return c.getWithStandardPagination(ctx, "/variants", "variants")
}

func (c *Client) GetSuppliers(ctx context.Context) ([]json.RawMessage, error) {
	return c.getWithStandardPagination(ctx, "/suppliers", "suppliers")
}

func (c *Client) GetPurchaseOrders(ctx context.Context) ([]json.RawMessage, error) {
	return c.getWithStandardPagination(ctx, "/purchase_orders", "purchase_orders")
}

func (c *Client) GetPOSDevices(ctx context.Context) ([]json.RawMessage, error) {
	return c.getWithStandardPagination(ctx, "/pos_devices", "pos_devices")
}

func (c *Client) GetCashRegisters(ctx context.Context) ([]json.RawMessage, error) {
	return c.getWithStandardPagination(ctx, "/cash_registers", "cash_registers")
}

func (c *Client) GetWebhooks(ctx context.Context) ([]json.RawMessage, error) {
	return c.getWithStandardPagination(ctx, "/webhooks", "webhooks")
}

// TestAllEndpoints tests all available endpoints
func (c *Client) TestAllEndpoints(ctx context.Context) map[string]interface{} {
	results := make(map[string]interface{})
	
	// Define endpoints with their expected data field names
	endpoints := map[string]struct {
		path      string
		dataField string
		method    func(context.Context) ([]json.RawMessage, error)
	}{
		"stores":          {"/stores", "stores", c.GetStores},
		"customers":       {"/customers", "customers", c.GetCustomers},
		"employees":       {"/employees", "employees", c.GetEmployees},
		"categories":      {"/categories", "categories", c.GetCategories},
		"items":           {"/items", "items", c.GetProducts},
		"inventory":       {"/inventory", "inventory_levels", c.GetInventoryLevels},
		"receipts":        {"/receipts", "receipts", c.GetReceipts},
		"discounts":       {"/discounts", "discounts", c.GetDiscounts},
		"modifiers":       {"/modifiers", "modifiers", c.GetModifiers},
		"taxes":           {"/taxes", "taxes", c.GetTaxes},
		"payment_types":   {"/payment_types", "payment_types", c.GetPaymentTypes},
		"variants":        {"/variants", "variants", c.GetVariants},
		"suppliers":       {"/suppliers", "suppliers", c.GetSuppliers},
		"purchase_orders": {"/purchase_orders", "purchase_orders", c.GetPurchaseOrders},
		"pos_devices":     {"/pos_devices", "pos_devices", c.GetPOSDevices},
		"cash_registers":  {"/cash_registers", "cash_registers", c.GetCashRegisters},
		"webhooks":        {"/webhooks", "webhooks", c.GetWebhooks},
	}
	
	// Test paginated endpoints
	for name, endpoint := range endpoints {
		log.Printf("Testing endpoint: %s (%s)", name, endpoint.path)
		
		data, err := endpoint.method(ctx)
		if err != nil {
			results[name] = map[string]interface{}{
				"success": false,
				"error":   err.Error(),
				"count":   0,
			}
		} else {
			results[name] = map[string]interface{}{
				"success": true,
				"count":   len(data),
			}
		}
	}
	
	// Test single object endpoints
	singleEndpoints := map[string]string{
		"account":  "/account",
		"settings": "/settings",
	}
	
	for name, path := range singleEndpoints {
		log.Printf("Testing endpoint: %s (%s)", name, path)
		
		var err error
		switch name {
		case "account":
			_, err = c.GetAccount(ctx)
		case "settings":
			_, err = c.GetSettings(ctx)
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
				"count":   1,
			}
		}
	}
	
	return results
}

// Additional helper methods for compatibility with main.go

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
		"store_stocks":    "/store_stocks",
		"categories":      "/categories",
		"items":           "/items",
		"inventory":       "/inventory",
		"receipts":        "/receipts",
		"account":         "/account",
		"settings":        "/settings",
	}
}

// GetEndpoint is a generic method for any endpoint
func (c *Client) GetEndpoint(ctx context.Context, endpoint string) ([]byte, error) {
	return c.Request(ctx, http.MethodGet, endpoint, nil)
}

// GetReceiptsWithParams retrieves receipts with parameters (compatibility method)
func (c *Client) GetReceiptsWithParams(ctx context.Context, params map[string]string) ([]json.RawMessage, error) {
	return c.GetReceipts(ctx)
}

// GetRecentReceipts retrieves recent receipts (compatibility method)
func (c *Client) GetRecentReceipts(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetReceipts(ctx)
}

// GetStoreStocks retrieves store stock levels (using generic helper)
func (c *Client) GetStoreStocks(ctx context.Context) ([]json.RawMessage, error) {
	return c.getWithStandardPagination(ctx, "/store_stocks", "store_stocks")
}

// GetInventory for compatibility (alias for GetInventoryLevels)
func (c *Client) GetInventory(ctx context.Context) ([]json.RawMessage, error) {
	return c.GetInventoryLevels(ctx)
}