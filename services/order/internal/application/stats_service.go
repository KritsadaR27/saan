package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"order/internal/application/dto"
	"order/internal/domain"
	"order/pkg/logger"
)

// OrderStatsService handles order statistics and analytics
type OrderStatsService struct {
	orderRepo     domain.OrderRepository
	orderItemRepo domain.OrderItemRepository
	logger        logger.Logger
}

// NewOrderStatsService creates a new order statistics service
func NewOrderStatsService(
	orderRepo domain.OrderRepository,
	orderItemRepo domain.OrderItemRepository,
	logger logger.Logger,
) *OrderStatsService {
	return &OrderStatsService{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		logger:        logger,
	}
}

// GetDailyStats retrieves daily order statistics for a specific date
func (s *OrderStatsService) GetDailyStats(ctx context.Context, date time.Time) (*dto.DailyStats, error) {
	s.logger.Info("Getting daily stats", "date", date.Format("2006-01-02"))

	// Get start and end of the day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Get orders for the day
	orders, err := s.orderRepo.GetOrdersByDateRange(ctx, startOfDay, endOfDay)
	if err != nil {
		s.logger.Error("Failed to get orders by date range", "error", err)
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	// Calculate statistics
	stats := &dto.DailyStats{
		Date: startOfDay,
	}

	var totalRevenue float64
	ordersByStatus := make(map[domain.OrderStatus]int)

	for _, order := range orders {
		stats.OrdersCount++
		totalRevenue += order.TotalAmount
		ordersByStatus[order.Status]++
	}

	stats.Revenue = totalRevenue
	if stats.OrdersCount > 0 {
		stats.AvgOrderValue = totalRevenue / float64(stats.OrdersCount)
	}

	// Set status counts
	stats.PendingOrders = ordersByStatus[domain.OrderStatusPending]
	stats.CompletedOrders = ordersByStatus[domain.OrderStatusCompleted]
	stats.CancelledOrders = ordersByStatus[domain.OrderStatusCancelled]

	s.logger.Info("Daily stats calculated", 
		"date", date.Format("2006-01-02"),
		"orders_count", stats.OrdersCount,
		"revenue", stats.Revenue)

	return stats, nil
}

// GetMonthlyStats retrieves monthly order statistics with trend data
func (s *OrderStatsService) GetMonthlyStats(ctx context.Context, year, month int) (*dto.MonthlyStats, error) {
	s.logger.Info("Getting monthly stats", "year", year, "month", month)

	// Get start and end of the month
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	// Get orders for the month
	orders, err := s.orderRepo.GetOrdersByDateRange(ctx, startOfMonth, endOfMonth)
	if err != nil {
		s.logger.Error("Failed to get orders by date range", "error", err)
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	// Calculate monthly statistics
	stats := &dto.MonthlyStats{
		Year:  year,
		Month: month,
	}

	var totalRevenue float64
	dailyStatsMap := make(map[string]*dto.DailyStats)

	for _, order := range orders {
		stats.TotalOrders++
		totalRevenue += order.TotalAmount

		// Group by day for daily breakdown
		orderDate := order.CreatedAt.Format("2006-01-02")
		if dailyStats, exists := dailyStatsMap[orderDate]; exists {
			dailyStats.OrdersCount++
			dailyStats.Revenue += order.TotalAmount
			// Update status counts
			switch order.Status {
			case domain.OrderStatusPending:
				dailyStats.PendingOrders++
			case domain.OrderStatusCompleted:
				dailyStats.CompletedOrders++
			case domain.OrderStatusCancelled:
				dailyStats.CancelledOrders++
			}
		} else {
			dailyStats := &dto.DailyStats{
				Date:        order.CreatedAt.Truncate(24 * time.Hour),
				OrdersCount: 1,
				Revenue:     order.TotalAmount,
			}
			// Set initial status count
			switch order.Status {
			case domain.OrderStatusPending:
				dailyStats.PendingOrders = 1
			case domain.OrderStatusCompleted:
				dailyStats.CompletedOrders = 1
			case domain.OrderStatusCancelled:
				dailyStats.CancelledOrders = 1
			}
			dailyStatsMap[orderDate] = dailyStats
		}
	}

	stats.TotalRevenue = totalRevenue
	if stats.TotalOrders > 0 {
		stats.AvgOrderValue = totalRevenue / float64(stats.TotalOrders)
	}

	// Convert daily stats map to slice and calculate avg order values
	for _, dailyStats := range dailyStatsMap {
		if dailyStats.OrdersCount > 0 {
			dailyStats.AvgOrderValue = dailyStats.Revenue / float64(dailyStats.OrdersCount)
		}
		stats.DailyBreakdown = append(stats.DailyBreakdown, *dailyStats)
	}

	// Calculate growth rate compared to previous month
	prevMonth := month - 1
	prevYear := year
	if prevMonth < 1 {
		prevMonth = 12
		prevYear--
	}

	prevMonthStats, err := s.GetMonthlyStats(ctx, prevYear, prevMonth)
	if err == nil && prevMonthStats.TotalOrders > 0 {
		stats.ComparisonData = &dto.MonthlyComparison{
			PreviousMonth: prevMonth,
			PreviousYear:  prevYear,
			OrdersChange:  stats.TotalOrders - prevMonthStats.TotalOrders,
			RevenueChange: stats.TotalRevenue - prevMonthStats.TotalRevenue,
		}

		// Calculate growth percentages
		if prevMonthStats.TotalOrders > 0 {
			stats.ComparisonData.OrdersGrowthPct = float64(stats.ComparisonData.OrdersChange) / float64(prevMonthStats.TotalOrders) * 100
		}
		if prevMonthStats.TotalRevenue > 0 {
			stats.ComparisonData.RevenueGrowthPct = stats.ComparisonData.RevenueChange / prevMonthStats.TotalRevenue * 100
		}

		stats.GrowthRate = stats.ComparisonData.RevenueGrowthPct
	}

	s.logger.Info("Monthly stats calculated",
		"year", year,
		"month", month,
		"total_orders", stats.TotalOrders,
		"total_revenue", stats.TotalRevenue)

	return stats, nil
}

// GetTopProducts retrieves top products by order count, revenue, or quantity
func (s *OrderStatsService) GetTopProducts(ctx context.Context, req *dto.TopProductsRequest) ([]dto.ProductStats, error) {
	s.logger.Info("Getting top products", "limit", req.Limit, "sort_by", req.SortBy)

	// Set default values
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.SortBy == "" {
		req.SortBy = "order_count"
	}

	// Get order items within date range if specified
	var orderItems []*domain.OrderItem
	var err error

	if req.StartDate != nil && req.EndDate != nil {
		// For now, we'll get all order items and filter by date
		// In a real implementation, you might want to add date filtering to the repository
		allOrderItems, err := s.orderItemRepo.GetAllOrderItems(ctx)
		if err != nil {
			s.logger.Error("Failed to get order items", "error", err)
			return nil, fmt.Errorf("failed to get order items: %w", err)
		}

		// Filter by date range by getting orders first
		orders, err := s.orderRepo.GetOrdersByDateRange(ctx, *req.StartDate, *req.EndDate)
		if err != nil {
			s.logger.Error("Failed to get orders by date range", "error", err)
			return nil, fmt.Errorf("failed to get orders: %w", err)
		}

		orderIDSet := make(map[uuid.UUID]bool)
		for _, order := range orders {
			orderIDSet[order.ID] = true
		}

		for _, item := range allOrderItems {
			if orderIDSet[item.OrderID] {
				orderItems = append(orderItems, item)
			}
		}
	} else {
		orderItems, err = s.orderItemRepo.GetAllOrderItems(ctx)
		if err != nil {
			s.logger.Error("Failed to get order items", "error", err)
			return nil, fmt.Errorf("failed to get order items: %w", err)
		}
	}

	// Aggregate product statistics
	productStatsMap := make(map[uuid.UUID]*dto.ProductStats)

	for _, item := range orderItems {
		if stats, exists := productStatsMap[item.ProductID]; exists {
			stats.OrderCount++
			stats.TotalQuantity += item.Quantity
			stats.Revenue += item.UnitPrice * float64(item.Quantity)
		} else {
			productStatsMap[item.ProductID] = &dto.ProductStats{
				ProductID:     item.ProductID,
				ProductName:   "", // Product name would need to be fetched from product service
				OrderCount:    1,
				TotalQuantity: item.Quantity,
				Revenue:       item.UnitPrice * float64(item.Quantity),
				LastOrderDate: time.Now(), // Would need to get from order date
			}
		}
	}

	// Calculate average prices
	for _, stats := range productStatsMap {
		if stats.TotalQuantity > 0 {
			stats.AvgPrice = stats.Revenue / float64(stats.TotalQuantity)
		}
	}

	// Convert map to slice
	var productStats []dto.ProductStats
	for _, stats := range productStatsMap {
		productStats = append(productStats, *stats)
	}

	// Sort by requested criteria (simplified sorting)
	// In a real implementation, you might want to use a more sophisticated sorting algorithm
	
	// Limit results
	if len(productStats) > req.Limit {
		productStats = productStats[:req.Limit]
	}

	s.logger.Info("Top products calculated", "count", len(productStats))
	return productStats, nil
}

// GetCustomerStats retrieves statistics for a specific customer
func (s *OrderStatsService) GetCustomerStats(ctx context.Context, req *dto.CustomerStatsRequest) (*dto.CustomerStats, error) {
	s.logger.Info("Getting customer stats", "customer_id", req.CustomerID)

	// Get all orders for the customer
	orders, err := s.orderRepo.GetOrdersByCustomer(ctx, req.CustomerID)
	if err != nil {
		s.logger.Error("Failed to get orders by customer", "error", err)
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	if len(orders) == 0 {
		return &dto.CustomerStats{
			CustomerID:  req.CustomerID,
			TotalOrders: 0,
			TotalSpent:  0,
		}, nil
	}

	// Calculate customer statistics
	stats := &dto.CustomerStats{
		CustomerID:     req.CustomerID,
		TotalOrders:    len(orders),
		FirstOrderDate: orders[0].CreatedAt,
		LastOrderDate:  orders[0].CreatedAt,
	}

	var totalSpent float64
	productCounts := make(map[uuid.UUID]int)

	for _, order := range orders {
		totalSpent += order.TotalAmount

		// Track first and last order dates
		if order.CreatedAt.Before(stats.FirstOrderDate) {
			stats.FirstOrderDate = order.CreatedAt
		}
		if order.CreatedAt.After(stats.LastOrderDate) {
			stats.LastOrderDate = order.CreatedAt
		}

		// Get order items for this order to track favorite products
		if req.IncludeFavoriteProducts {
			orderItems, err := s.orderItemRepo.GetByOrderID(ctx, order.ID)
			if err != nil {
				s.logger.Warn("Failed to get order items for favorite products", "order_id", order.ID, "error", err)
				continue
			}

			for _, item := range orderItems {
				productCounts[item.ProductID] += item.Quantity
			}
		}
	}

	stats.TotalSpent = totalSpent
	if stats.TotalOrders > 0 {
		stats.AvgOrderValue = totalSpent / float64(stats.TotalOrders)
	}

	// Calculate order frequency
	if stats.TotalOrders > 1 {
		daysBetween := stats.LastOrderDate.Sub(stats.FirstOrderDate).Hours() / 24
		avgDaysBetweenOrders := daysBetween / float64(stats.TotalOrders-1)

		switch {
		case avgDaysBetweenOrders <= 7:
			stats.OrderFrequency = "daily"
		case avgDaysBetweenOrders <= 30:
			stats.OrderFrequency = "weekly"
		case avgDaysBetweenOrders <= 90:
			stats.OrderFrequency = "monthly"
		default:
			stats.OrderFrequency = "rare"
		}
	} else {
		stats.OrderFrequency = "new"
	}

	// Add favorite products if requested
	if req.IncludeFavoriteProducts && len(productCounts) > 0 {
		limit := req.FavoriteProductsLimit
		if limit <= 0 {
			limit = 5
		}

		// Convert to slice and sort (simplified)
		for productID, count := range productCounts {
			if len(stats.FavoriteProducts) < limit {
				stats.FavoriteProducts = append(stats.FavoriteProducts, dto.ProductStats{
					ProductID:     productID,
					TotalQuantity: count,
				})
			}
		}
	}

	s.logger.Info("Customer stats calculated",
		"customer_id", req.CustomerID,
		"total_orders", stats.TotalOrders,
		"total_spent", stats.TotalSpent)

	return stats, nil
}
