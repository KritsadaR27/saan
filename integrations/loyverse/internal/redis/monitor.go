package redis

import (
	"context"
	"log"
	"sync"
	"time"
)

// HealthMonitor monitors Redis connection health
type HealthMonitor struct {
	client         *Client
	interval       time.Duration
	timeout        time.Duration
	alertThreshold int
	mu             sync.RWMutex
	isRunning      bool
	stopCh         chan struct{}
	healthHistory  []bool
	consecutiveFails int
	lastHealthCheck time.Time
	onHealthChange func(healthy bool)
}

// NewHealthMonitor creates a new Redis health monitor
func NewHealthMonitor(client *Client, interval time.Duration) *HealthMonitor {
	return &HealthMonitor{
		client:         client,
		interval:       interval,
		timeout:        5 * time.Second,
		alertThreshold: 3, // Alert after 3 consecutive failures
		stopCh:         make(chan struct{}),
		healthHistory:  make([]bool, 0, 10), // Keep last 10 health checks
	}
}

// SetTimeout sets the health check timeout
func (h *HealthMonitor) SetTimeout(timeout time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.timeout = timeout
}

// SetAlertThreshold sets the threshold for consecutive failures before alerting
func (h *HealthMonitor) SetAlertThreshold(threshold int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.alertThreshold = threshold
}

// SetHealthChangeCallback sets a callback function for health status changes
func (h *HealthMonitor) SetHealthChangeCallback(callback func(bool)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onHealthChange = callback
}

// Start begins monitoring Redis health
func (h *HealthMonitor) Start(ctx context.Context) {
	h.mu.Lock()
	if h.isRunning {
		h.mu.Unlock()
		return
	}
	h.isRunning = true
	h.mu.Unlock()

	log.Printf("Starting Redis health monitor with %v interval", h.interval)

	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	// Perform initial health check
	h.performHealthCheck()

	for {
		select {
		case <-ticker.C:
			h.performHealthCheck()
		case <-h.stopCh:
			log.Println("Redis health monitor stopped")
			return
		case <-ctx.Done():
			log.Println("Redis health monitor stopped due to context cancellation")
			return
		}
	}
}

// Stop stops the health monitor
func (h *HealthMonitor) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if !h.isRunning {
		return
	}
	
	h.isRunning = false
	close(h.stopCh)
}

// IsHealthy returns the current health status
func (h *HealthMonitor) IsHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.client.IsHealthy()
}

// GetHealthHistory returns the recent health check history
func (h *HealthMonitor) GetHealthHistory() []bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	// Return a copy of the history
	history := make([]bool, len(h.healthHistory))
	copy(history, h.healthHistory)
	return history
}

// GetConsecutiveFailures returns the number of consecutive failures
func (h *HealthMonitor) GetConsecutiveFailures() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.consecutiveFails
}

// GetLastHealthCheck returns the timestamp of the last health check
func (h *HealthMonitor) GetLastHealthCheck() time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastHealthCheck
}

// performHealthCheck executes a health check
func (h *HealthMonitor) performHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	start := time.Now()
	err := h.client.CheckHealth(ctx)
	duration := time.Since(start)
	
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.lastHealthCheck = time.Now()
	isHealthy := err == nil
	
	// Update health history
	h.healthHistory = append(h.healthHistory, isHealthy)
	if len(h.healthHistory) > 10 {
		h.healthHistory = h.healthHistory[1:]
	}
	
	// Update consecutive failures counter
	if isHealthy {
		if h.consecutiveFails > 0 {
			log.Printf("Redis health recovered after %d failures", h.consecutiveFails)
		}
		h.consecutiveFails = 0
	} else {
		h.consecutiveFails++
		log.Printf("Redis health check failed (attempt %d): %v (took %v)", 
			h.consecutiveFails, err, duration)
		
		// Alert if threshold reached
		if h.consecutiveFails >= h.alertThreshold {
			h.alertHealthIssue(err)
		}
	}
	
	// Call health change callback if set
	if h.onHealthChange != nil {
		go h.onHealthChange(isHealthy)
	}
	
	// Log periodic stats
	if h.consecutiveFails == 0 && len(h.healthHistory) > 0 && len(h.healthHistory)%5 == 0 {
		h.logHealthStats(duration)
	}
}

// alertHealthIssue logs an alert for health issues
func (h *HealthMonitor) alertHealthIssue(err error) {
	log.Printf("ALERT: Redis health check failed %d consecutive times (threshold: %d). Last error: %v", 
		h.consecutiveFails, h.alertThreshold, err)
	
	// Log Redis statistics for debugging
	h.client.LogStats()
}

// logHealthStats logs periodic health statistics
func (h *HealthMonitor) logHealthStats(lastCheckDuration time.Duration) {
	healthyCount := 0
	for _, healthy := range h.healthHistory {
		if healthy {
			healthyCount++
		}
	}
	
	healthPercentage := float64(healthyCount) / float64(len(h.healthHistory)) * 100
	
	log.Printf("Redis Health Stats - Success Rate: %.1f%% (%d/%d), Last Check: %v", 
		healthPercentage, healthyCount, len(h.healthHistory), lastCheckDuration)
}
