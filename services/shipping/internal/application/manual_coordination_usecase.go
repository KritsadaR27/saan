package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"shipping/internal/domain/entity"
	"shipping/internal/domain/repository"
)

// ManualCoordinationUseCase handles manual coordination operations for deliveries
type ManualCoordinationUseCase struct {
	deliveryRepo  repository.DeliveryRepository
	manualRepo    repository.ManualTaskRepository
	snapshotRepo  repository.SnapshotRepository
	eventPub      EventPublisher
	cache         Cache
}

// NewManualCoordinationUseCase creates a new manual coordination use case
func NewManualCoordinationUseCase(
	deliveryRepo repository.DeliveryRepository,
	manualRepo repository.ManualTaskRepository,
	snapshotRepo repository.SnapshotRepository,
	eventPub EventPublisher,
	cache Cache,
) *ManualCoordinationUseCase {
	return &ManualCoordinationUseCase{
		deliveryRepo: deliveryRepo,
		manualRepo:   manualRepo,
		snapshotRepo: snapshotRepo,
		eventPub:     eventPub,
		cache:        cache,
	}
}

// CreateManualTaskRequest represents a request to create a manual task
type CreateManualTaskRequest struct {
	DeliveryID       uuid.UUID                  `json:"delivery_id" validate:"required"`
	TaskType         string                     `json:"task_type" validate:"required"`
	Instructions     string                     `json:"instructions" validate:"required"`
	AssignedToUserID *uuid.UUID                 `json:"assigned_to_user_id,omitempty"`
	ContactInfo      map[string]interface{}     `json:"contact_info,omitempty"`
	ReminderTime     *time.Time                 `json:"reminder_time,omitempty"`
	CreatedBy        string                     `json:"created_by" validate:"required"`
}

// UpdateTaskStatusRequest represents a request to update task status
type UpdateTaskStatusRequest struct {
	TaskID            uuid.UUID `json:"task_id" validate:"required"`
	Status            string    `json:"status" validate:"required"`
	CompletionNotes   string    `json:"completion_notes,omitempty"`
	ExternalReference string    `json:"external_reference,omitempty"`
	UpdatedBy         string    `json:"updated_by" validate:"required"`
}

// AssignTaskRequest represents a request to assign a task
type AssignTaskRequest struct {
	TaskID       uuid.UUID `json:"task_id" validate:"required"`
	UserID       uuid.UUID `json:"user_id" validate:"required"`
	AssignedBy   string    `json:"assigned_by" validate:"required"`
}

// CreateManualTask creates a new manual coordination task
func (uc *ManualCoordinationUseCase) CreateManualTask(ctx context.Context, req CreateManualTaskRequest) (*entity.ManualCoordinationTask, error) {
	// Validate request
	if req.DeliveryID == uuid.Nil {
		return nil, fmt.Errorf("delivery ID is required")
	}

	// Get delivery to ensure it exists
	delivery, err := uc.deliveryRepo.GetByID(ctx, req.DeliveryID)
	if err != nil {
		return nil, fmt.Errorf("delivery not found: %w", err)
	}

	// Create contact information map if not provided
	contactInfo := req.ContactInfo
	if contactInfo == nil {
		contactInfo = make(map[string]interface{})
	}
	
	// Add delivery context to contact info
	if delivery.CustomerAddressID != uuid.Nil {
		contactInfo["customer_address_id"] = delivery.CustomerAddressID.String()
	}

	// Create manual coordination task
	task, err := entity.NewManualCoordinationTask(
		req.DeliveryID,
		"manual", // Default provider code for manual tasks 
		entity.TaskType(req.TaskType),
		req.Instructions,
		contactInfo,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create manual task: %w", err)
	}

	// Set optional fields
	if req.AssignedToUserID != nil {
		task.AssignToUser(*req.AssignedToUserID)
	}

	if req.ReminderTime != nil {
		task.SetNextReminder(*req.ReminderTime)
	}

	// Save task
	if err := uc.manualRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to save manual task: %w", err)
	}

	// Create snapshot for task creation
	snapshot := &entity.DeliverySnapshot{
		ID:                uuid.New(),
		DeliveryID:        delivery.ID,
		SnapshotType:      entity.SnapshotTypeAssigned,
		SnapshotData: map[string]interface{}{
			"task_id":      task.ID.String(),
			"task_type":    string(task.TaskType),
			"instructions": task.TaskInstructions,
			"assigned_to":  task.AssignedToUserID,
			"created_by":   req.CreatedBy,
		},
		TriggeredBy:    req.CreatedBy,
		TriggeredEvent: "manual_task_created",
		CreatedAt:      time.Now(),
	}

	if err := uc.snapshotRepo.Create(ctx, snapshot); err != nil {
		return nil, fmt.Errorf("failed to create task snapshot: %w", err)
	}

	// Publish event
	uc.eventPub.Publish(ctx, "manual_task.created", map[string]interface{}{
		"task_id":      task.ID.String(),
		"delivery_id":  req.DeliveryID.String(),
		"task_type":    string(task.TaskType),
		"instructions": task.TaskInstructions,
		"assigned_to":  task.AssignedToUserID,
		"created_by":   req.CreatedBy,
		"created_at":   task.CreatedAt,
	})

	// Clear cache
	uc.cache.Delete(ctx, fmt.Sprintf("manual_tasks_delivery_%s", req.DeliveryID.String()))

	return task, nil
}

// GetManualTasksForDelivery retrieves all manual tasks for a delivery
func (uc *ManualCoordinationUseCase) GetManualTasksForDelivery(ctx context.Context, deliveryID uuid.UUID) ([]*entity.ManualCoordinationTask, error) {
	tasks, err := uc.manualRepo.GetByDeliveryID(ctx, deliveryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get manual tasks for delivery: %w", err)
	}

	return tasks, nil
}

// GetTasksByAssignee retrieves all tasks assigned to a specific user
func (uc *ManualCoordinationUseCase) GetTasksByAssignee(ctx context.Context, userID uuid.UUID) ([]*entity.ManualCoordinationTask, error) {
	tasks, err := uc.manualRepo.GetByAssignedUser(ctx, userID, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by assignee: %w", err)
	}

	return tasks, nil
}

// GetTaskStatistics returns statistics for manual tasks
func (uc *ManualCoordinationUseCase) GetTaskStatistics(ctx context.Context) (map[string]interface{}, error) {
	cacheKey := "manual_task_statistics"
	
	// Try cache first
	if cached, err := uc.cache.Get(ctx, cacheKey); err == nil {
		// Cache returns empty string if not found, skip parsing in that case
		if cached != "" {
			// For simplicity, just skip cache and go to repository
			// TODO: Implement proper JSON marshaling/unmarshaling for cache
		}
	}

	// Get statistics from repository
	pendingTasks, err := uc.manualRepo.GetPendingTasks(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending tasks: %w", err)
	}

	overdueTasks, err := uc.manualRepo.GetOverdueTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue tasks: %w", err)
	}

	activeTasks, err := uc.manualRepo.GetActiveTasks(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get active tasks: %w", err)
	}

	stats := map[string]interface{}{
		"pending_count":  len(pendingTasks),
		"overdue_count":  len(overdueTasks),
		"active_count":   len(activeTasks),
		"last_updated":   time.Now(),
	}

	// TODO: Cache for 5 minutes with proper JSON serialization
	// uc.cache.Set(ctx, cacheKey, marshaledStats, 5*time.Minute)

	return stats, nil
}
