package connector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	loyverseAPIBaseURL = "https://api.loyverse.com/v1.0"
)

// LoyverseAPI client for interacting with Loyverse API
type LoyverseAPI struct {
	apiToken   string
	httpClient *http.Client
}

// NewLoyverseAPI creates a new Loyverse API client
func NewLoyverseAPI(apiToken string) *LoyverseAPI {
	return &LoyverseAPI{
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Category represents a Loyverse category
type Category struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CategoriesResponse represents the response from Loyverse categories API
type CategoriesResponse struct {
	Categories []Category `json:"categories"`
}

// GetCategories fetches all categories from Loyverse API
func (api *LoyverseAPI) GetCategories() ([]Category, error) {
	url := fmt.Sprintf("%s/categories", loyverseAPIBaseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+api.apiToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	
	var categoriesResp CategoriesResponse
	if err := json.Unmarshal(body, &categoriesResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	
	return categoriesResp.Categories, nil
}
