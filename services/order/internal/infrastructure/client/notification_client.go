package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// NotificationRequest represents a notification request
type NotificationRequest struct {
	Type        string                 `json:"type"`         // email, sms, line, facebook
	Recipients  []string               `json:"recipients"`   // email addresses, phone numbers, user IDs
	Template    string                 `json:"template"`     // template name
	Subject     string                 `json:"subject,omitempty"` // for email
	Data        map[string]interface{} `json:"data"`         // template variables
	Priority    string                 `json:"priority"`     // low, normal, high, urgent
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"` // for scheduled notifications
}

// NotificationResponse represents notification response
type NotificationResponse struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`      // queued, sent, failed, scheduled
	Recipients  []string  `json:"recipients"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	FailedAt    *time.Time `json:"failed_at,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// BulkNotificationRequest represents a bulk notification request
type BulkNotificationRequest struct {
	Notifications []NotificationRequest `json:"notifications"`
	BatchID       string                `json:"batch_id,omitempty"`
}

// BulkNotificationResponse represents bulk notification response
type BulkNotificationResponse struct {
	BatchID       string                 `json:"batch_id"`
	TotalCount    int                    `json:"total_count"`
	SuccessCount  int                    `json:"success_count"`
	FailedCount   int                    `json:"failed_count"`
	Notifications []NotificationResponse `json:"notifications"`
}

// NotificationClient interface for communicating with notification service
type NotificationClient interface {
	SendNotification(ctx context.Context, req *NotificationRequest) (*NotificationResponse, error)
	SendBulkNotifications(ctx context.Context, req *BulkNotificationRequest) (*BulkNotificationResponse, error)
	GetNotificationStatus(ctx context.Context, notificationID uuid.UUID) (*NotificationResponse, error)
	
	// Convenience methods for common order notifications
	SendOrderConfirmation(ctx context.Context, customerEmail string, orderData map[string]interface{}) error
	SendOrderStatusUpdate(ctx context.Context, customerEmail string, orderData map[string]interface{}) error
	SendOrderDelivered(ctx context.Context, customerEmail string, orderData map[string]interface{}) error
	
	// Chat-specific methods
	SendChatMessage(ctx context.Context, chatID string, message string) error
}

// HTTPNotificationClient implements NotificationClient using HTTP requests
type HTTPNotificationClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPNotificationClient creates a new HTTP notification client
func NewHTTPNotificationClient(baseURL string) *HTTPNotificationClient {
	return &HTTPNotificationClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewNotificationClient creates a new notification client using service name from PROJECT_RULES.md
func NewNotificationClient() *HTTPNotificationClient {
	// Use service name as per PROJECT_RULES.md - notification-service:8092
	return NewHTTPNotificationClient("http://notification-service:8092")
}

// SendNotification sends a single notification
func (c *HTTPNotificationClient) SendNotification(ctx context.Context, req *NotificationRequest) (*NotificationResponse, error) {
	url := fmt.Sprintf("%s/api/notifications", c.baseURL)
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "order-service/1.0")
	
	resp, err := c.executeWithRetry(httpReq, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to execute send notification request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("notification service returned status %d", resp.StatusCode)
	}
	
	var notificationResponse NotificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&notificationResponse); err != nil {
		return nil, fmt.Errorf("failed to decode notification response: %w", err)
	}
	
	return &notificationResponse, nil
}

// SendBulkNotifications sends multiple notifications in batch
func (c *HTTPNotificationClient) SendBulkNotifications(ctx context.Context, req *BulkNotificationRequest) (*BulkNotificationResponse, error) {
	url := fmt.Sprintf("%s/api/notifications/bulk", c.baseURL)
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "order-service/1.0")
	
	resp, err := c.executeWithRetry(httpReq, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to execute bulk notification request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("notification service returned status %d", resp.StatusCode)
	}
	
	var bulkResponse BulkNotificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&bulkResponse); err != nil {
		return nil, fmt.Errorf("failed to decode bulk notification response: %w", err)
	}
	
	return &bulkResponse, nil
}

// GetNotificationStatus gets notification status by ID
func (c *HTTPNotificationClient) GetNotificationStatus(ctx context.Context, notificationID uuid.UUID) (*NotificationResponse, error) {
	url := fmt.Sprintf("%s/api/notifications/%s", c.baseURL, notificationID.String())
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "order-service/1.0")
	
	resp, err := c.executeWithRetry(req, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get notification request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("notification not found")
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("notification service returned status %d", resp.StatusCode)
	}
	
	var notificationResponse NotificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&notificationResponse); err != nil {
		return nil, fmt.Errorf("failed to decode notification response: %w", err)
	}
	
	return &notificationResponse, nil
}

// SendOrderConfirmation sends order confirmation email
func (c *HTTPNotificationClient) SendOrderConfirmation(ctx context.Context, customerEmail string, orderData map[string]interface{}) error {
	req := &NotificationRequest{
		Type:       "email",
		Recipients: []string{customerEmail},
		Template:   "order_confirmation",
		Subject:    "Order Confirmation",
		Data:       orderData,
		Priority:   "normal",
	}
	
	_, err := c.SendNotification(ctx, req)
	return err
}

// SendOrderStatusUpdate sends order status update notification
func (c *HTTPNotificationClient) SendOrderStatusUpdate(ctx context.Context, customerEmail string, orderData map[string]interface{}) error {
	req := &NotificationRequest{
		Type:       "email",
		Recipients: []string{customerEmail},
		Template:   "order_status_update",
		Subject:    "Order Status Update",
		Data:       orderData,
		Priority:   "normal",
	}
	
	_, err := c.SendNotification(ctx, req)
	return err
}

// SendOrderDelivered sends order delivered notification
func (c *HTTPNotificationClient) SendOrderDelivered(ctx context.Context, customerEmail string, orderData map[string]interface{}) error {
	req := &NotificationRequest{
		Type:       "email",
		Recipients: []string{customerEmail},
		Template:   "order_delivered",
		Subject:    "Your Order Has Been Delivered",
		Data:       orderData,
		Priority:   "high",
	}
	
	_, err := c.SendNotification(ctx, req)
	return err
}

// SendChatMessage sends a chat message through notification service
func (c *HTTPNotificationClient) SendChatMessage(ctx context.Context, chatID string, message string) error {
	req := &NotificationRequest{
		Type:       "line", // หรือ "facebook" ขึ้นอยู่กับ chat platform
		Recipients: []string{chatID},
		Template:   "plain_text",
		Data: map[string]interface{}{
			"message": message,
		},
		Priority: "normal",
	}
	
	_, err := c.SendNotification(ctx, req)
	return err
}

// executeWithRetry executes HTTP request with retry logic
func (c *HTTPNotificationClient) executeWithRetry(req *http.Request, maxRetries int) (*http.Response, error) {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := c.client.Do(req)
		if err == nil {
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return resp, nil
			}
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return resp, nil
			}
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: status %d", resp.StatusCode)
		} else {
			lastErr = err
		}
		
		if attempt < maxRetries {
			backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
			time.Sleep(backoff)
		}
	}
	
	return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}
