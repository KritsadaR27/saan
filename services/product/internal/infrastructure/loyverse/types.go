package loyverse

import (
	"time"
)

// LoyverseProduct represents a product from Loyverse API
type LoyverseProduct struct {
	ID          string    `json:"id"`
	Handle      string    `json:"handle"`
	ItemName    string    `json:"item_name"`
	Description string    `json:"description"`
	CategoryID  string    `json:"category_id"`
	ImageURL    string    `json:"image_url"`
	Option1Name string    `json:"option1_name"`
	Option2Name string    `json:"option2_name"`
	Option3Name string    `json:"option3_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Variants    []LoyverseVariant `json:"variants"`
}

// LoyverseVariant represents a product variant from Loyverse API
type LoyverseVariant struct {
	ID               string  `json:"id"`
	ItemName         string  `json:"item_name"`
	Option1Value     string  `json:"option1_value"`
	Option2Value     string  `json:"option2_value"`
	Option3Value     string  `json:"option3_value"`
	SKU              string  `json:"sku"`
	Barcode          string  `json:"barcode"`
	Cost             float64 `json:"cost"`
	DefaultPrice     float64 `json:"default_price"`
	Stores           []LoyverseStore `json:"stores"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// LoyverseStore represents store-specific data for a variant
type LoyverseStore struct {
	StoreID      string  `json:"store_id"`
	Price        float64 `json:"price"`
	Available    bool    `json:"available"`
	Quantity     int     `json:"quantity"`
	ReorderLevel int     `json:"reorder_level"`
	ReorderAmount int    `json:"reorder_amount"`
}

// LoyverseCategory represents a category from Loyverse API
type LoyverseCategory struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProductsResponse represents the response from Loyverse products API
type ProductsResponse struct {
	Products []LoyverseProduct `json:"items"`
	Cursor   string            `json:"cursor"`
}

// CategoriesResponse represents the response from Loyverse categories API
type CategoriesResponse struct {
	Categories []LoyverseCategory `json:"categories"`
}

// ErrorResponse represents an error response from Loyverse API
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	ProductsProcessed   int    `json:"products_processed"`
	CategoriesProcessed int    `json:"categories_processed"`
	Errors              []string `json:"errors"`
	StartTime           time.Time `json:"start_time"`
	EndTime             time.Time `json:"end_time"`
	Duration            time.Duration `json:"duration"`
}
