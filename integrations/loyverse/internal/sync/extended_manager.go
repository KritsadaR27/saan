// integrations/loyverse/internal/sync/extended_manager.go
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

// ExtendedManager manages all synchronization tasks including new entities
type ExtendedManager struct {
	// Core syncs
	productSync   *ProductSync
	inventorySync *InventorySync
	receiptSync   *ReceiptSync
	customerSync  *CustomerSync

	// Extended syncs
	employeeSync    *EmployeeSync
	categorySync    *CategorySync
	supplierSync    *SupplierSync
	discountSync    *DiscountSync
	paymentTypeSync *PaymentTypeSync
	storeSync       *StoreSync

	cron  *cron.Cron
	redis *redis.Client
	mu    sync.Mutex
}

// ExtendedConfig holds sync configuration
type ExtendedConfig struct {
	// Core sync intervals
	ProductSyncInterval   string
	InventorySyncInterval string
	ReceiptSyncInterval   string
	CustomerSyncInterval  string

	// Extended sync intervals
	EmployeeSyncInterval    string
	CategorySyncInterval    string
	SupplierSyncInterval    string
	DiscountSyncInterval    string
	PaymentTypeSyncInterval string
	StoreSyncInterval       string

	// Master data sync (all at once)
	MasterDataSyncTime string // e.g., "03:00" for 3 AM

	TimeZone string
}

// NewExtendedManager creates a new extended sync manager
func NewExtendedManager(
	// Core syncs
	productSync *ProductSync,
	inventorySync *InventorySync,
	receiptSync *ReceiptSync,
	customerSync *CustomerSync,
	// Extended syncs
	employeeSync *EmployeeSync,
	categorySync *CategorySync,
	supplierSync *SupplierSync,
	discountSync *DiscountSync,
	paymentTypeSync *PaymentTypeSync,
	storeSync *StoreSync,
	redis *redis.Client,
	config ExtendedConfig,
) (*ExtendedManager, error) {
	loc, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		return nil, fmt.Errorf("loading timezone: %w", err)
	}

	return &ExtendedManager{
		productSync:     productSync,
		inventorySync:   inventorySync,
		receiptSync:     receiptSync,
		customerSync:    customerSync,
		employeeSync:    employeeSync,
		categorySync:    categorySync,
		supplierSync:    supplierSync,
		discountSync:    discountSync,
		paymentTypeSync: paymentTypeSync,
		storeSync:       storeSync,
		cron:            cron.New(cron.WithLocation(loc)),
		redis:           redis,
	}, nil
}

// Start starts all scheduled sync jobs
func (m *ExtendedManager) Start(ctx context.Context) error {
	// High-frequency syncs
	if _, err := m.cron.AddFunc("*/5 * * * *", func() {
		m.runSync(ctx, "receipts", m.receiptSync.Sync)
	}); err != nil {
		return fmt.Errorf("adding receipt sync job: %w", err)
	}

	if _, err := m.cron.AddFunc("*/15 * * * *", func() {
		m.runSync(ctx, "inventory", m.inventorySync.Sync)
	}); err != nil {
		return fmt.Errorf("adding inventory sync job: %w", err)
	}

	// Medium-frequency syncs
	if _, err := m.cron.AddFunc("*/30 * * * *", func() {
		m.runSync(ctx, "products", m.productSync.Sync)
	}); err != nil {
		return fmt.Errorf("adding product sync job: %w", err)
	}

	if _, err := m.cron.AddFunc("0 * * * *", func() {
		m.runSync(ctx, "customers", m.customerSync.Sync)
	}); err != nil {
		return fmt.Errorf("adding customer sync job: %w", err)
	}

	// Master data sync (once daily)
	masterDataTime := "0 3 * * *" // 3 AM daily
	if _, err := m.cron.AddFunc(masterDataTime, func() {
		m.runMasterDataSync(ctx)
	}); err != nil {
		return fmt.Errorf("adding master data sync job: %w", err)
	}

	// Category sync (twice daily - can change frequently)
	if _, err := m.cron.AddFunc("0 */12 * * *", func() {
		m.runSync(ctx, "categories", m.categorySync.Sync)
	}); err != nil {
		return fmt.Errorf("adding category sync job: %w", err)
	}

	// Discount sync (every 4 hours - promotions can change)
	if _, err := m.cron.AddFunc("0 */4 * * *", func() {
		m.runSync(ctx, "discounts", m.discountSync.Sync)
	}); err != nil {
		return fmt.Errorf("adding discount sync job: %w", err)
	}

	m.cron.Start()
	log.Println("Extended sync manager started")
	return nil
}

// runMasterDataSync runs all master data syncs
func (m *ExtendedManager) runMasterDataSync(ctx context.Context) {
	log.Println("Starting master data sync...")

	syncs := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"stores", m.storeSync.Sync},
		{"employees", m.employeeSync.Sync},
		{"payment_types", m.paymentTypeSync.Sync},
		{"suppliers", m.supplierSync.Sync},
		{"categories", m.categorySync.Sync},
		{"discounts", m.discountSync.Sync},
	}

	for _, sync := range syncs {
		if err := m.runSync(ctx, sync.name, sync.fn); err != nil {
			log.Printf("Master data sync failed for %s: %v", sync.name, err)
		}
		// Small delay between syncs
		time.Sleep(5 * time.Second)
	}

	log.Println("Master data sync completed")
}

// TriggerSync manually triggers a sync for a specific type
func (m *ExtendedManager) TriggerSync(ctx context.Context, syncType string) error {
	switch syncType {
	case "products":
		return m.runSync(ctx, syncType, m.productSync.Sync)
	case "inventory":
		return m.runSync(ctx, syncType, m.inventorySync.Sync)
	case "receipts":
		return m.runSync(ctx, syncType, m.receiptSync.Sync)
	case "customers":
		return m.runSync(ctx, syncType, m.customerSync.Sync)
	case "employees":
		return m.runSync(ctx, syncType, m.employeeSync.Sync)
	case "categories":
		return m.runSync(ctx, syncType, m.categorySync.Sync)
	case "suppliers":
		return m.runSync(ctx, syncType, m.supplierSync.Sync)
	case "discounts":
		return m.runSync(ctx, syncType, m.discountSync.Sync)
	case "payment_types":
		return m.runSync(ctx, syncType, m.paymentTypeSync.Sync)
	case "stores":
		return m.runSync(ctx, syncType, m.storeSync.Sync)
	case "master_data":
		m.runMasterDataSync(ctx)
		return nil
	default:
		return fmt.Errorf("unknown sync type: %s", syncType)
	}
}

// GetSyncStatus returns the status of all sync jobs
func (m *ExtendedManager) GetSyncStatus(ctx context.Context) (map[string]interface{}, error) {
	status := make(map[string]interface{})
	syncTypes := []string{
		"products", "inventory", "receipts", "customers",
		"employees", "categories", "suppliers", "discounts",
		"payment_types", "stores",
	}

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

		// Get record count from cache
		countKey := fmt.Sprintf("loyverse:sync:count:%s", syncType)
		if count, err := m.redis.Get(ctx, countKey).Result(); err == nil {
			syncStatus["record_count"] = count
		}

		status[syncType] = syncStatus
	}

	// Add master data sync status
	masterStatus := make(map[string]interface{})
	if lastRun, err := m.redis.Get(ctx, "loyverse:sync:master_data:last").Result(); err == nil {
		masterStatus["last_run"] = lastRun
	}
	status["master_data"] = masterStatus

	return status, nil
}

// runSync runs a sync job with distributed locking
func (m *ExtendedManager) runSync(ctx context.Context, syncType string, syncFunc func(context.Context) error) error {
	lockKey := fmt.Sprintf("loyverse:sync:lock:%s", syncType)

	// Try to acquire lock with 10 minute timeout
	locked, err := m.redis.SetNX(ctx, lockKey, "1", 10*time.Minute).Result()
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
	m.redis.Set(ctx, startKey, time.Now().Format(time.RFC3339), 7*24*time.Hour)

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
	m.redis.Set(ctx, successKey, time.Now().Format(time.RFC3339), 7*24*time.Hour)

	// Clear error if exists
	errorKey := fmt.Sprintf("loyverse:sync:error:%s", syncType)
	m.redis.Del(ctx, errorKey)

	log.Printf("Completed %s sync in %v", syncType, duration)
	return nil
}
