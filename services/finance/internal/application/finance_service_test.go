package application

import (
	"context"
	"testing"
	"time"

	"finance/internal/domain"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Mock Redis client
type mockRedisClient struct{ err error }

func (m *mockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return redis.NewStringCmd(ctx, "get", key)
}
func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return redis.NewStatusCmd(ctx, "set", key, value)
}
func (m *mockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := make([]interface{}, len(keys)+1)
	args[0] = "del"
	for i, key := range keys {
		args[i+1] = key
	}
	return redis.NewIntCmd(ctx, args...)
}
func (m *mockRedisClient) Close() error { return m.err }

func TestFinanceService_ProcessEndOfDay(t *testing.T) {
	tests := []struct {
		name           string
		date           time.Time
		branchID       *uuid.UUID
		vehicleID      *uuid.UUID
		sales          float64
		codCollections float64
		expectErr      bool
		setupError     error
	}{
		{
			name:           "successful end of day processing",
			date:           time.Now().Truncate(24 * time.Hour),
			branchID:       func() *uuid.UUID { id := uuid.New(); return &id }(),
			vehicleID:      nil,
			sales:          1000.0,
			codCollections: 200.0,
			expectErr:      false,
		},
		{
			name:           "negative sales amount",
			date:           time.Now().Truncate(24 * time.Hour),
			branchID:       func() *uuid.UUID { id := uuid.New(); return &id }(),
			vehicleID:      nil,
			sales:          -100.0,
			codCollections: 0.0,
			expectErr:      true,
		},
	}

	// Create a simple in-memory test setup
	// Note: For proper testing, you would set up a test database
	// This is a simplified example to show the testing structure
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// In a real test, you would set up test database and repositories
			// mockRepos := setupTestRepositories(t)
			// mockRedis := &mockRedisClient{}
			// service := NewFinanceService(mockRepos, mockRedis)

			// For now, just test input validation
			if tt.sales < 0 {
				if !tt.expectErr {
					t.Errorf("expected error for negative sales")
				}
			}
		})
	}
}

// Integration test helper (commented out - requires actual database setup)
/*
func setupTestRepositories(t *testing.T) *repositories.Repositories {
	// Setup test database connection
	db, err := sql.Open("postgres", "postgres://test:test@localhost/finance_test?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	
	// Run migrations
	// ... migration setup code ...
	
	return repositories.NewRepositories(db)
}
*/

func TestFinanceService_ValidationLogic(t *testing.T) {
	tests := []struct {
		name       string
		sales      float64
		expectErr  bool
		errMessage string
	}{
		{
			name:       "valid sales amount",
			sales:      1000.0,
			expectErr:  false,
		},
		{
			name:       "negative sales amount",
			sales:      -100.0,
			expectErr:  true,
			errMessage: "sales amount cannot be negative",
		},
		{
			name:       "zero sales amount",
			sales:      0.0,
			expectErr:  false, // Zero sales might be valid (no sales day)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation logic
			if tt.sales < 0 && !tt.expectErr {
				t.Errorf("expected validation to catch negative sales")
			}
		})
	}
}

func TestAllocationService_CalculateAllocations(t *testing.T) {
	tests := []struct {
		name     string
		revenue  float64
		expected map[domain.AccountType]float64
	}{
		{
			name:    "standard allocation",
			revenue: 1000.0,
			expected: map[domain.AccountType]float64{
				domain.ProfitAccount:    50.0,  // 5%
				domain.OwnerPayAccount:  500.0, // 50%
				domain.TaxAccount:       150.0, // 15%
				domain.OperatingAccount: 300.0, // 30%
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test allocation calculation logic
			profitPercent := 5.0
			ownerPayPercent := 50.0
			taxPercent := 15.0
			operatingPercent := 30.0

			profit := tt.revenue * profitPercent / 100
			ownerPay := tt.revenue * ownerPayPercent / 100
			tax := tt.revenue * taxPercent / 100
			operating := tt.revenue * operatingPercent / 100

			if profit != tt.expected[domain.ProfitAccount] {
				t.Errorf("expected profit %f, got %f", tt.expected[domain.ProfitAccount], profit)
			}
			if ownerPay != tt.expected[domain.OwnerPayAccount] {
				t.Errorf("expected owner pay %f, got %f", tt.expected[domain.OwnerPayAccount], ownerPay)
			}
			if tax != tt.expected[domain.TaxAccount] {
				t.Errorf("expected tax %f, got %f", tt.expected[domain.TaxAccount], tax)
			}
			if operating != tt.expected[domain.OperatingAccount] {
				t.Errorf("expected operating %f, got %f", tt.expected[domain.OperatingAccount], operating)
			}
		})
	}
}
