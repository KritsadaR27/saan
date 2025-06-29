package dto

import (
	"time"
	"github.com/google/uuid"
)

// DailyStats represents daily order statistics
type DailyStats struct {
	Date           time.Time `json:"date"`
	OrdersCount    int       `json:"orders_count"`
	Revenue        float64   `json:"revenue"`
	AvgOrderValue  float64   `json:"avg_order_value"`
	PendingOrders  int       `json:"pending_orders"`
	CompletedOrders int      `json:"completed_orders"`
	CancelledOrders int      `json:"cancelled_orders"`
}

// MonthlyStats represents monthly order statistics with trend data
type MonthlyStats struct {
	Year            int         `json:"year"`
	Month           int         `json:"month"`
	TotalOrders     int         `json:"total_orders"`
	TotalRevenue    float64     `json:"total_revenue"`
	AvgOrderValue   float64     `json:"avg_order_value"`
	DailyBreakdown  []DailyStats `json:"daily_breakdown"`
	GrowthRate      float64     `json:"growth_rate"`      // Compared to previous month
	ComparisonData  *MonthlyComparison `json:"comparison_data,omitempty"`
}

// MonthlyComparison represents comparison with previous month
type MonthlyComparison struct {
	PreviousMonth   int     `json:"previous_month"`
	PreviousYear    int     `json:"previous_year"`
	OrdersChange    int     `json:"orders_change"`
	RevenueChange   float64 `json:"revenue_change"`
	OrdersGrowthPct float64 `json:"orders_growth_pct"`
	RevenueGrowthPct float64 `json:"revenue_growth_pct"`
}

// ProductStats represents statistics for a specific product
type ProductStats struct {
	ProductID     uuid.UUID `json:"product_id"`
	ProductName   string    `json:"product_name,omitempty"`
	OrderCount    int       `json:"order_count"`
	TotalQuantity int       `json:"total_quantity"`
	Revenue       float64   `json:"revenue"`
	AvgPrice      float64   `json:"avg_price"`
	LastOrderDate time.Time `json:"last_order_date"`
}

// CustomerStats represents statistics for a specific customer
type CustomerStats struct {
	CustomerID       uuid.UUID `json:"customer_id"`
	CustomerEmail    string    `json:"customer_email,omitempty"`
	CustomerName     string    `json:"customer_name,omitempty"`
	TotalOrders      int       `json:"total_orders"`
	TotalSpent       float64   `json:"total_spent"`
	AvgOrderValue    float64   `json:"avg_order_value"`
	FirstOrderDate   time.Time `json:"first_order_date"`
	LastOrderDate    time.Time `json:"last_order_date"`
	FavoriteProducts []ProductStats `json:"favorite_products,omitempty"`
	OrderFrequency   string    `json:"order_frequency"` // daily, weekly, monthly, rare
}

// GetStatsRequest represents request for getting statistics
type GetStatsRequest struct {
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	GroupBy      string     `json:"group_by,omitempty"`      // daily, weekly, monthly
	IncludeComparison bool   `json:"include_comparison"`
}

// TopProductsRequest represents request for top products
type TopProductsRequest struct {
	Limit     int        `json:"limit"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	SortBy    string     `json:"sort_by"` // order_count, revenue, quantity
}

// CustomerStatsRequest represents request for customer statistics
type CustomerStatsRequest struct {
	CustomerID           uuid.UUID `json:"customer_id"`
	IncludeFavoriteProducts bool   `json:"include_favorite_products"`
	FavoriteProductsLimit   int    `json:"favorite_products_limit"`
}
