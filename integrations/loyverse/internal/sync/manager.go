// integrations/loyverse/internal/sync/manager.go
package sync

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/robfig/cron/v3"
)

// Manager manages all synchronization tasks
type Manager struct {
	productSync   *ProductSync
	inventorySync *InventorySync
	receiptSync   *ReceiptSync
	customerSync  *CustomerSync
	cron          *cron.Cron
	redis         *redis.Client
	mu            sync.Mutex
}

// Config holds sync configuration
type Config struct {
	ProductSyncInterval   string
	InventorySyncInterval string
	ReceiptSyncInterval   string
	CustomerSyncInterval  string
	TimeZone              string
}

// NewManager creates a new sync manager
func NewManager(
	productSync *ProductSync,
	inventorySync *InventorySync,
	receiptSync *ReceiptSync,
	customerSync *CustomerSync,
	redis *redis.Client,
	config Config,
) (*Manager, error) {
	loc, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		return nil, fmt.Errorf("loading timezone: %w", err)
	}

	return &Manager{
		productSync:   productSync,
		inventorySync: inventorySync,
		receiptSync:   receiptSync,
		customerSync:  customerSync,
		cron:          cron.New(cron.WithLocation(loc)),
		redis:         redis,
	}, nil
}

// Start starts all scheduled sync jobs
func (m *Manager) Start(ctx context.Context) error {
	// Add sync jobs to cron
	if _, err := m.cron.AddFunc("*/30 * * * *", func() {
		m.runSync(ctx, "products", m.productSync.Sync)
	}); err != nil {
		return fmt.Errorf("adding product sync job: %w", err)
	}

	if _, err := m.cron.AddFunc("*/15 * * * *", func() {
		m.runSync(ctx, "inventory", m.inventorySync.Sync)
	}); err != nil {
		return fmt.Errorf("adding inventory sync job: %w", err)
	}

	if _, err := m.cron.AddFunc("*/5 * * * *", func() {
		m.runSync(ctx, "receipts", m.receiptSync.Sync)
	}); err != nil {
		return fmt.Errorf("adding receipt sync job: %w", err)
	}

	if _, err := m.cron.AddFunc("0 * * * *", func() {
		m.runSync(ctx, "customers", m.customerSync.Sync)
	}); err != nil {
		return fmt.Errorf("adding customer sync job: %w", err)
	}

	m.cron.Start()
	log.Println("Sync manager started")
	return nil
}

// Stop stops all sync jobs
func (m *Manager) Stop() {
	m.cron.Stop()
	log.Println("Sync manager stopped")
}

// TriggerSync manually triggers a sync for a specific type
func (m *Manager) TriggerSync(ctx context.Context, syncType string) error {
	switch syncType {
	case "products":
		return m.runSync(ctx, syncType, m.productSync.Sync)
	case "inventory":
		return m.runSync(ctx, syncType, m.inventorySync.Sync)
	case "receipts":
		return m.runSync(ctx, syncType, m.receiptSync.Sync)
	case "customers":
		return m.runSync(ctx, syncType, m.customerSync.Sync)
	default:
		return fmt.Errorf("unknown sync type: %s", syncType)
	}
}

// runSync runs a sync job with distributed locking
func (m *Manager) runSync(ctx context.Context, syncType string, syncFunc func(context.Context) error) error {
	lockKey := fmt.Sprintf("loyverse:sync:lock:%s", syncType)

	// Try to acquire lock
	locked, err := m.redis.SetNX(ctx, lockKey, "1", 5*time.Minute).Result()
	if err != nil {
		return fmt.Errorf("acquiring lock: %w", err)
	}

	if !locked {
		log.Printf("Sync %s already running on another instance", syncType)
		return nil
	}

	// Ensure lock is released
	defer m.redis.Del(ctx, lockKey)

	// Record sync start
	startKey := fmt.Sprintf("loyverse:sync:start:%s", syncType)
	m.redis.Set(ctx, startKey, time.Now().Format(time.RFC3339), 24*time.Hour)

	log.Printf("Starting %s sync", syncType)
	start := time.Now()

	// Run sync
	if err := syncFunc(ctx); err != nil {
		// Record error
		errorKey := fmt.Sprintf("loyverse:sync:error:%s", syncType)
		m.redis.Set(ctx, errorKey, err.Error(), 24*time.Hour)
		return fmt.Errorf("sync %s failed: %w", syncType, err)
	}

	// Record success
	duration := time.Since(start)
	successKey := fmt.Sprintf("loyverse:sync:success:%s", syncType)
	m.redis.Set(ctx, successKey, time.Now().Format(time.RFC3339), 24*time.Hour)

	log.Printf("Completed %s sync in %v", syncType, duration)
	return nil
}

// GetSyncStatus returns the status of all sync jobs
func (m *Manager) GetSyncStatus(ctx context.Context) (map[string]interface{}, error) {
	status := make(map[string]interface{})
	syncTypes := []string{"products", "inventory", "receipts", "customers"}

	for _, syncType := range syncTypes {
		syncStatus := make(map[string]interface{})

		// Get last start time
		startKey := fmt.Sprintf("loyverse:sync:start:%s", syncType)
		if start, err := m.redis.Get(ctx, startKey).Result(); err == nil {
			syncStatus["last_start"] = start
		}

		// Get last success time
		successKey := fmt.Sprintf("loyverse:sync:success:%s", syncType)
		if success, err := m.redis.Get(ctx, successKey).Result(); err == nil {
			syncStatus["last_success"] = success
		}

		// Get last error
		errorKey := fmt.Sprintf("loyverse:sync:error:%s", syncType)
		if errMsg, err := m.redis.Get(ctx, errorKey).Result(); err == nil {
			syncStatus["last_error"] = errMsg
		}

		// Check if currently running
		lockKey := fmt.Sprintf("loyverse:sync:lock:%s", syncType)
		if locked, err := m.redis.Exists(ctx, lockKey).Result(); err == nil && locked > 0 {
			syncStatus["is_running"] = true
		} else {
			syncStatus["is_running"] = false
		}

		status[syncType] = syncStatus
	}

	return status, nil
}
