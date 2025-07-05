package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"shipping/internal/domain/entity"
)

// ManualTaskRepository defines the contract for manual coordination task data persistence
type ManualTaskRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, task *entity.ManualCoordinationTask) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ManualCoordinationTask, error)
	Update(ctx context.Context, task *entity.ManualCoordinationTask) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Query operations
	GetByDeliveryID(ctx context.Context, deliveryID uuid.UUID) ([]*entity.ManualCoordinationTask, error)
	GetByProviderCode(ctx context.Context, providerCode string, limit, offset int) ([]*entity.ManualCoordinationTask, error)
	GetByTaskType(ctx context.Context, taskType entity.TaskType, limit, offset int) ([]*entity.ManualCoordinationTask, error)
	GetByStatus(ctx context.Context, status entity.TaskStatus, limit, offset int) ([]*entity.ManualCoordinationTask, error)
	GetByAssignedUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.ManualCoordinationTask, error)
	
	// Task management
	GetPendingTasks(ctx context.Context, limit, offset int) ([]*entity.ManualCoordinationTask, error)
	GetActiveTasks(ctx context.Context, limit, offset int) ([]*entity.ManualCoordinationTask, error)
	GetOverdueTasks(ctx context.Context) ([]*entity.ManualCoordinationTask, error)
	GetTasksDueForReminder(ctx context.Context) ([]*entity.ManualCoordinationTask, error)
	
	// Assignment operations
	AssignTask(ctx context.Context, taskID, userID uuid.UUID) error
	UnassignTask(ctx context.Context, taskID uuid.UUID) error
	GetUnassignedTasks(ctx context.Context, limit, offset int) ([]*entity.ManualCoordinationTask, error)
	
	// Completion operations
	CompleteTask(ctx context.Context, taskID uuid.UUID, completionNotes, externalReference string) error
	FailTask(ctx context.Context, taskID uuid.UUID, reason string) error
	CancelTask(ctx context.Context, taskID uuid.UUID, reason string) error
	
	// Reminder operations
	MarkReminderSent(ctx context.Context, taskID uuid.UUID) error
	SetNextReminder(ctx context.Context, taskID uuid.UUID, reminderTime time.Time) error
	GetTasksForReminder(ctx context.Context, maxReminderTime time.Time) ([]*entity.ManualCoordinationTask, error)
	
	// Provider-specific queries
	GetInterExpressTasks(ctx context.Context, status entity.TaskStatus, limit, offset int) ([]*entity.ManualCoordinationTask, error)
	GetNimExpressTasks(ctx context.Context, status entity.TaskStatus, limit, offset int) ([]*entity.ManualCoordinationTask, error)
	GetRotRaoTasks(ctx context.Context, status entity.TaskStatus, limit, offset int) ([]*entity.ManualCoordinationTask, error)
	
	// Analytics and reporting
	GetTaskMetrics(ctx context.Context, startDate, endDate time.Time) (*TaskMetrics, error)
	GetTaskMetricsByProvider(ctx context.Context, providerCode string, startDate, endDate time.Time) (*TaskMetrics, error)
	GetTaskMetricsByUser(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*TaskMetrics, error)
	GetAverageTaskCompletionTime(ctx context.Context, taskType entity.TaskType, startDate, endDate time.Time) (float64, error)
	
	// Search and filtering
	SearchTasks(ctx context.Context, filters *TaskQueryFilters) ([]*entity.ManualCoordinationTask, error)
	GetTasksByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.ManualCoordinationTask, error)
	GetCompletedTasksByDate(ctx context.Context, date time.Time) ([]*entity.ManualCoordinationTask, error)
	
	// Bulk operations
	UpdateMultipleTaskStatuses(ctx context.Context, taskIDs []uuid.UUID, status entity.TaskStatus) error
	GetTasksByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.ManualCoordinationTask, error)
	CreateBulkTasks(ctx context.Context, tasks []*entity.ManualCoordinationTask) error
	
	// Performance tracking
	GetUserProductivity(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*UserProductivityMetrics, error)
	GetProviderTaskEfficiency(ctx context.Context, providerCode string, startDate, endDate time.Time) (*ProviderTaskEfficiency, error)
	GetTaskBacklog(ctx context.Context) (*TaskBacklog, error)
}

// TaskQueryFilters represents filters for task queries
type TaskQueryFilters struct {
	DeliveryID        *uuid.UUID           `json:"delivery_id,omitempty"`
	ProviderCode      *string              `json:"provider_code,omitempty"`
	TaskType          *entity.TaskType     `json:"task_type,omitempty"`
	TaskStatus        *entity.TaskStatus   `json:"task_status,omitempty"`
	AssignedToUserID  *uuid.UUID           `json:"assigned_to_user_id,omitempty"`
	CreatedAfter      *time.Time           `json:"created_after,omitempty"`
	CreatedBefore     *time.Time           `json:"created_before,omitempty"`
	CompletedAfter    *time.Time           `json:"completed_after,omitempty"`
	CompletedBefore   *time.Time           `json:"completed_before,omitempty"`
	IsOverdue         *bool                `json:"is_overdue,omitempty"`
	NeedsReminder     *bool                `json:"needs_reminder,omitempty"`
	MinReminderCount  *int                 `json:"min_reminder_count,omitempty"`
	MaxReminderCount  *int                 `json:"max_reminder_count,omitempty"`
	HasExternalRef    *bool                `json:"has_external_ref,omitempty"`
	Limit             int                  `json:"limit"`
	Offset            int                  `json:"offset"`
}

// TaskMetrics represents task performance metrics
type TaskMetrics struct {
	TotalTasks           int64     `json:"total_tasks"`
	PendingTasks         int64     `json:"pending_tasks"`
	InProgressTasks      int64     `json:"in_progress_tasks"`
	CompletedTasks       int64     `json:"completed_tasks"`
	FailedTasks          int64     `json:"failed_tasks"`
	CancelledTasks       int64     `json:"cancelled_tasks"`
	OverdueTasks         int64     `json:"overdue_tasks"`
	AverageCompletionTime float64  `json:"average_completion_time_hours"`
	CompletionRate       float64   `json:"completion_rate_percentage"`
	OnTimeCompletionRate float64   `json:"on_time_completion_rate_percentage"`
	TotalReminders       int64     `json:"total_reminders"`
	AverageReminders     float64   `json:"average_reminders_per_task"`
	PeriodStart          time.Time `json:"period_start"`
	PeriodEnd            time.Time `json:"period_end"`
}

// UserProductivityMetrics represents user productivity metrics for tasks
type UserProductivityMetrics struct {
	UserID               uuid.UUID `json:"user_id"`
	TasksAssigned        int64     `json:"tasks_assigned"`
	TasksCompleted       int64     `json:"tasks_completed"`
	TasksFailed          int64     `json:"tasks_failed"`
	TasksCancelled       int64     `json:"tasks_cancelled"`
	AverageCompletionTime float64  `json:"average_completion_time_hours"`
	ProductivityScore    float64   `json:"productivity_score"`
	CompletionRate       float64   `json:"completion_rate_percentage"`
	QualityScore         float64   `json:"quality_score"`
	TaskTypeBreakdown    map[entity.TaskType]int64 `json:"task_type_breakdown"`
	ProviderBreakdown    map[string]int64          `json:"provider_breakdown"`
	PeriodStart          time.Time `json:"period_start"`
	PeriodEnd            time.Time `json:"period_end"`
}

// ProviderTaskEfficiency represents task efficiency metrics for a provider
type ProviderTaskEfficiency struct {
	ProviderCode         string               `json:"provider_code"`
	TotalTasks           int64                `json:"total_tasks"`
	CompletedTasks       int64                `json:"completed_tasks"`
	FailedTasks          int64                `json:"failed_tasks"`
	AverageCompletionTime float64             `json:"average_completion_time_hours"`
	SuccessRate          float64              `json:"success_rate_percentage"`
	TaskTypeEfficiency   map[entity.TaskType]float64 `json:"task_type_efficiency"`
	PeakTaskHours        map[string]int64     `json:"peak_task_hours"`
	PeriodStart          time.Time            `json:"period_start"`
	PeriodEnd            time.Time            `json:"period_end"`
}

// TaskBacklog represents current task backlog information
type TaskBacklog struct {
	TotalBacklog         int64                        `json:"total_backlog"`
	BacklogByProvider    map[string]int64             `json:"backlog_by_provider"`
	BacklogByTaskType    map[entity.TaskType]int64    `json:"backlog_by_task_type"`
	BacklogByAge         map[string]int64             `json:"backlog_by_age"`
	OverdueCount         int64                        `json:"overdue_count"`
	UnassignedCount      int64                        `json:"unassigned_count"`
	HighPriorityCount    int64                        `json:"high_priority_count"`
	EstimatedClearTime   float64                      `json:"estimated_clear_time_hours"`
	RecommendedActions   []string                     `json:"recommended_actions"`
	UpdatedAt            time.Time                    `json:"updated_at"`
}

// TaskReminder represents a task reminder
type TaskReminder struct {
	TaskID           uuid.UUID           `json:"task_id"`
	DeliveryID       uuid.UUID           `json:"delivery_id"`
	ProviderCode     string              `json:"provider_code"`
	TaskType         entity.TaskType     `json:"task_type"`
	AssignedToUserID *uuid.UUID          `json:"assigned_to_user_id,omitempty"`
	ReminderCount    int                 `json:"reminder_count"`
	DueTime          time.Time           `json:"due_time"`
	Instructions     string              `json:"instructions"`
	ContactInfo      map[string]interface{} `json:"contact_info"`
	IsOverdue        bool                `json:"is_overdue"`
	UrgencyLevel     string              `json:"urgency_level"`
}

// TaskAssignment represents a task assignment
type TaskAssignment struct {
	TaskID           uuid.UUID       `json:"task_id"`
	UserID           uuid.UUID       `json:"user_id"`
	AssignedAt       time.Time       `json:"assigned_at"`
	AssignedBy       *uuid.UUID      `json:"assigned_by,omitempty"`
	Priority         string          `json:"priority"`
	ExpectedDuration time.Duration   `json:"expected_duration"`
	Notes            string          `json:"notes"`
}
