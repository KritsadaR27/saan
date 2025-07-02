package domain

import (
	"time"
)

// Analytics represents various analytics data
type Analytics struct {
	// Dashboard Summary
	TotalProducts     int     `json:"total_products"`
	TotalStores       int     `json:"total_stores"`
	LowStockItems     int     `json:"low_stock_items"`
	TotalStockValue   float64 `json:"total_stock_value"`
	
	// Stock Status Distribution
	InStock           int     `json:"in_stock"`
	LowStock          int     `json:"low_stock"`
	OutOfStock        int     `json:"out_of_stock"`
	
	// Top Performers
	TopSellingProducts []ProductPerformance `json:"top_selling_products"`
	TopCategories      []CategoryPerformance `json:"top_categories"`
	
	// Movement Trends
	DailyMovements     []DailyMovement `json:"daily_movements"`
	WeeklyTrends       []WeeklyTrend   `json:"weekly_trends"`
	
	GeneratedAt        time.Time `json:"generated_at"`
}

// ProductPerformance represents product sales/movement performance
type ProductPerformance struct {
	ProductID       string  `json:"product_id"`
	ProductName     string  `json:"product_name"`
	CategoryName    string  `json:"category_name"`
	TotalSold       float64 `json:"total_sold"`
	Revenue         float64 `json:"revenue"`
	CurrentStock    float64 `json:"current_stock"`
	TurnoverRate    float64 `json:"turnover_rate"`
	Rank            int     `json:"rank"`
}

// CategoryPerformance represents category performance metrics
type CategoryPerformance struct {
	CategoryID      string  `json:"category_id"`
	CategoryName    string  `json:"category_name"`
	ProductCount    int     `json:"product_count"`
	TotalSold       float64 `json:"total_sold"`
	Revenue         float64 `json:"revenue"`
	AvgTurnover     float64 `json:"avg_turnover"`
	Rank            int     `json:"rank"`
}

// DailyMovement represents daily stock movement summary
type DailyMovement struct {
	Date         time.Time `json:"date"`
	TotalSales   float64   `json:"total_sales"`
	TotalPurchases float64 `json:"total_purchases"`
	TotalAdjustments float64 `json:"total_adjustments"`
	NetMovement  float64   `json:"net_movement"`
	ValueMoved   float64   `json:"value_moved"`
}

// WeeklyTrend represents weekly trend analysis
type WeeklyTrend struct {
	WeekStart    time.Time `json:"week_start"`
	WeekEnd      time.Time `json:"week_end"`
	TotalSales   float64   `json:"total_sales"`
	Revenue      float64   `json:"revenue"`
	GrowthRate   float64   `json:"growth_rate"`
	TopCategory  string    `json:"top_category"`
}

// StockAlert represents inventory alerts with context
type StockAlert struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"` // LOW_STOCK, OUT_OF_STOCK, REORDER_SUGGESTION
	Severity        string    `json:"severity"` // HIGH, MEDIUM, LOW
	ProductID       string    `json:"product_id"`
	ProductName     string    `json:"product_name"`
	StoreID         string    `json:"store_id"`
	StoreName       string    `json:"store_name"`
	CurrentStock    float64   `json:"current_stock"`
	ReorderLevel    float64   `json:"reorder_level"`
	SuggestedOrder  float64   `json:"suggested_order"`
	DaysToStockout  int       `json:"days_to_stockout"`
	LastSaleDate    *time.Time `json:"last_sale_date"`
	Message         string    `json:"message"`
	CreatedAt       time.Time `json:"created_at"`
}

// ReorderSuggestion represents intelligent reorder suggestions
type ReorderSuggestion struct {
	ProductID       string    `json:"product_id"`
	ProductName     string    `json:"product_name"`
	StoreID         string    `json:"store_id"`
	StoreName       string    `json:"store_name"`
	CurrentStock    float64   `json:"current_stock"`
	SuggestedQty    float64   `json:"suggested_qty"`
	ReasonCode      string    `json:"reason_code"` // LOW_STOCK, SEASONAL, TREND_UP
	Confidence      float64   `json:"confidence"` // 0.0 - 1.0
	EstimatedCost   float64   `json:"estimated_cost"`
	EstimatedDemand float64   `json:"estimated_demand"`
	LeadTimeDays    int       `json:"lead_time_days"`
	Priority        int       `json:"priority"` // 1-5, 1 being highest
	CreatedAt       time.Time `json:"created_at"`
}
