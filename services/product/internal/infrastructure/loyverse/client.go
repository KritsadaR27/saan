package loyverse

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// Client represents a Loyverse API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewClient creates a new Loyverse API client
func NewClient(apiKey string, logger *logrus.Logger) *Client {
	return &Client{
		baseURL: "https://api.loyverse.com/v1.0",
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GetProducts fetches products from Loyverse API with pagination
func (c *Client) GetProducts(ctx context.Context, cursor string, limit int) (*ProductsResponse, error) {
	endpoint := "/items"
	
	// Build query parameters
	params := url.Values{}
	if cursor != "" {
		params.Set("cursor", cursor)
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	
	url := c.baseURL + endpoint + "?" + params.Encode()
	
	c.logger.WithFields(logrus.Fields{
		"url":    url,
		"cursor": cursor,
		"limit":  limit,
	}).Debug("Fetching products from Loyverse")
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Saan-System/1.0")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}
	
	var result ProductsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	c.logger.WithFields(logrus.Fields{
		"products_count": len(result.Products),
		"next_cursor":    result.Cursor,
	}).Debug("Successfully fetched products from Loyverse")
	
	return &result, nil
}

// GetCategories fetches categories from Loyverse API
func (c *Client) GetCategories(ctx context.Context) (*CategoriesResponse, error) {
	endpoint := "/categories"
	url := c.baseURL + endpoint
	
	c.logger.WithField("url", url).Debug("Fetching categories from Loyverse")
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Saan-System/1.0")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}
	
	var result CategoriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	c.logger.WithField("categories_count", len(result.Categories)).Debug("Successfully fetched categories from Loyverse")
	
	return &result, nil
}

// GetProduct fetches a single product by ID from Loyverse API
func (c *Client) GetProduct(ctx context.Context, productID string) (*LoyverseProduct, error) {
	endpoint := fmt.Sprintf("/items/%s", productID)
	url := c.baseURL + endpoint
	
	c.logger.WithFields(logrus.Fields{
		"url":        url,
		"product_id": productID,
	}).Debug("Fetching product from Loyverse")
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Saan-System/1.0")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}
	
	var product LoyverseProduct
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	c.logger.WithField("product_id", productID).Debug("Successfully fetched product from Loyverse")
	
	return &product, nil
}

// GetCategory fetches a single category by ID from Loyverse API
func (c *Client) GetCategory(ctx context.Context, categoryID string) (*LoyverseCategory, error) {
	endpoint := fmt.Sprintf("/categories/%s", categoryID)
	url := c.baseURL + endpoint
	
	c.logger.WithFields(logrus.Fields{
		"url":         url,
		"category_id": categoryID,
	}).Debug("Fetching category from Loyverse")
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Saan-System/1.0")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}
	
	var category LoyverseCategory
	if err := json.NewDecoder(resp.Body).Decode(&category); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	c.logger.WithField("category_id", categoryID).Debug("Successfully fetched category from Loyverse")
	
	return &category, nil
}

// handleErrorResponse handles error responses from Loyverse API
func (c *Client) handleErrorResponse(resp *http.Response) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HTTP %d: failed to read error response", resp.StatusCode)
	}
	
	var errorResp ErrorResponse
	if err := json.Unmarshal(bodyBytes, &errorResp); err != nil {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	c.logger.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"error_code":  errorResp.Code,
		"message":     errorResp.Message,
		"details":     errorResp.Details,
	}).Error("Loyverse API error")
	
	return fmt.Errorf("Loyverse API error (HTTP %d): %s - %s", resp.StatusCode, errorResp.Message, errorResp.Details)
}

// TestConnection tests the connection to Loyverse API
func (c *Client) TestConnection(ctx context.Context) error {
	// Test by fetching categories with a small limit
	_, err := c.GetCategories(ctx)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	
	c.logger.Info("Loyverse API connection test successful")
	return nil
}
