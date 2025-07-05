package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"shipping/internal/domain/entity"
	"shipping/internal/domain/repository"
)

type manualTaskRepository struct {
	db *sqlx.DB
}

// NewManualTaskRepository creates a new manual task repository implementation
func NewManualTaskRepository(db *sqlx.DB) repository.ManualTaskRepository {
	return &manualTaskRepository{db: db}
}

// Create creates a new manual coordination task
func (r *manualTaskRepository) Create(ctx context.Context, task *entity.ManualCoordinationTask) error {
	query := `
		INSERT INTO manual_coordination_tasks (
			id, delivery_id, provider_code, task_type, task_status,
			assigned_to_user_id, task_instructions, contact_information,
			completed_at, completion_notes, external_reference,
			reminder_count, last_reminder_sent, next_reminder_due,
			created_at, updated_at
		) VALUES (
			:id, :delivery_id, :provider_code, :task_type, :task_status,
			:assigned_to_user_id, :task_instructions, :contact_information,
			:completed_at, :completion_notes, :external_reference,
			:reminder_count, :last_reminder_sent, :next_reminder_due,
			:created_at, :updated_at
		)`

	contactInfoJSON, _ := json.Marshal(task.ContactInformation)

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":                   task.ID,
		"delivery_id":          task.DeliveryID,
		"provider_code":        task.ProviderCode,
		"task_type":            task.TaskType,
		"task_status":          task.TaskStatus,
		"assigned_to_user_id":  task.AssignedToUserID,
		"task_instructions":    task.TaskInstructions,
		"contact_information":  contactInfoJSON,
		"completed_at":         task.CompletedAt,
		"completion_notes":     task.CompletionNotes,
		"external_reference":   task.ExternalReference,
		"reminder_count":       task.ReminderCount,
		"last_reminder_sent":   task.LastReminderSent,
		"next_reminder_due":    task.NextReminderDue,
		"created_at":           task.CreatedAt,
		"updated_at":           task.UpdatedAt,
	})

	return err
}

// GetByID retrieves a manual coordination task by ID
func (r *manualTaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ManualCoordinationTask, error) {
	query := `
		SELECT id, delivery_id, provider_code, task_type, task_status,
			   assigned_to_user_id, task_instructions, contact_information,
			   completed_at, completion_notes, external_reference,
			   reminder_count, last_reminder_sent, next_reminder_due,
			   created_at, updated_at
		FROM manual_coordination_tasks 
		WHERE id = $1`

	var task entity.ManualCoordinationTask
	var contactInfoJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.DeliveryID, &task.ProviderCode, &task.TaskType, &task.TaskStatus,
		&task.AssignedToUserID, &task.TaskInstructions, &contactInfoJSON,
		&task.CompletedAt, &task.CompletionNotes, &task.ExternalReference,
		&task.ReminderCount, &task.LastReminderSent, &task.NextReminderDue,
		&task.CreatedAt, &task.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrManualTaskNotFound
		}
		return nil, fmt.Errorf("failed to get manual coordination task: %w", err)
	}

	// Unmarshal JSON fields
	if len(contactInfoJSON) > 0 {
		json.Unmarshal(contactInfoJSON, &task.ContactInformation)
	}

	return &task, nil
}

// Update updates an existing manual coordination task
func (r *manualTaskRepository) Update(ctx context.Context, task *entity.ManualCoordinationTask) error {
	contactInfoJSON, _ := json.Marshal(task.ContactInformation)

	query := `
		UPDATE manual_coordination_tasks SET
			provider_code = :provider_code,
			task_type = :task_type,
			task_status = :task_status,
			assigned_to_user_id = :assigned_to_user_id,
			task_instructions = :task_instructions,
			contact_information = :contact_information,
			completed_at = :completed_at,
			completion_notes = :completion_notes,
			external_reference = :external_reference,
			reminder_count = :reminder_count,
			last_reminder_sent = :last_reminder_sent,
			next_reminder_due = :next_reminder_due,
			updated_at = :updated_at
		WHERE id = :id`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":                   task.ID,
		"provider_code":        task.ProviderCode,
		"task_type":            task.TaskType,
		"task_status":          task.TaskStatus,
		"assigned_to_user_id":  task.AssignedToUserID,
		"task_instructions":    task.TaskInstructions,
		"contact_information":  contactInfoJSON,
		"completed_at":         task.CompletedAt,
		"completion_notes":     task.CompletionNotes,
		"external_reference":   task.ExternalReference,
		"reminder_count":       task.ReminderCount,
		"last_reminder_sent":   task.LastReminderSent,
		"next_reminder_due":    task.NextReminderDue,
		"updated_at":           task.UpdatedAt,
	})

	return err
}

// Delete deletes a manual coordination task
func (r *manualTaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM manual_coordination_tasks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete manual coordination task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrManualTaskNotFound
	}

	return nil
}

// GetByDeliveryID retrieves all manual coordination tasks for a delivery
func (r *manualTaskRepository) GetByDeliveryID(ctx context.Context, deliveryID uuid.UUID) ([]*entity.ManualCoordinationTask, error) {
	query := `
		SELECT id, delivery_id, provider_code, task_type, task_status,
			   assigned_to_user_id, task_instructions, contact_information,
			   completed_at, completion_notes, external_reference,
			   reminder_count, last_reminder_sent, next_reminder_due,
			   created_at, updated_at
		FROM manual_coordination_tasks 
		WHERE delivery_id = $1
		ORDER BY created_at ASC`

	return r.queryTasks(ctx, query, deliveryID)
}

// GetByProviderCode retrieves manual coordination tasks by provider
func (r *manualTaskRepository) GetByProviderCode(ctx context.Context, providerCode string, limit, offset int) ([]*entity.ManualCoordinationTask, error) {
	query := `
		SELECT id, delivery_id, provider_code, task_type, task_status,
			   assigned_to_user_id, task_instructions, contact_information,
			   completed_at, completion_notes, external_reference,
			   reminder_count, last_reminder_sent, next_reminder_due,
			   created_at, updated_at
		FROM manual_coordination_tasks 
		WHERE provider_code = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	return r.queryTasks(ctx, query, providerCode, limit, offset)
}

// GetByTaskType retrieves manual coordination tasks by task type
func (r *manualTaskRepository) GetByTaskType(ctx context.Context, taskType entity.TaskType, limit, offset int) ([]*entity.ManualCoordinationTask, error) {
	query := `
		SELECT id, delivery_id, provider_code, task_type, task_status,
			   assigned_to_user_id, task_instructions, contact_information,
			   completed_at, completion_notes, external_reference,
			   reminder_count, last_reminder_sent, next_reminder_due,
			   created_at, updated_at
		FROM manual_coordination_tasks 
		WHERE task_type = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	return r.queryTasks(ctx, query, taskType, limit, offset)
}

// GetByStatus retrieves manual coordination tasks by status
func (r *manualTaskRepository) GetByStatus(ctx context.Context, status entity.TaskStatus, limit, offset int) ([]*entity.ManualCoordinationTask, error) {
	query := `
		SELECT id, delivery_id, provider_code, task_type, task_status,
			   assigned_to_user_id, task_instructions, contact_information,
			   completed_at, completion_notes, external_reference,
			   reminder_count, last_reminder_sent, next_reminder_due,
			   created_at, updated_at
		FROM manual_coordination_tasks 
		WHERE task_status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	return r.queryTasks(ctx, query, status, limit, offset)
}

// GetByAssignedUser retrieves manual coordination tasks assigned to a user
func (r *manualTaskRepository) GetByAssignedUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.ManualCoordinationTask, error) {
	query := `
		SELECT id, delivery_id, provider_code, task_type, task_status,
			   assigned_to_user_id, task_instructions, contact_information,
			   completed_at, completion_notes, external_reference,
			   reminder_count, last_reminder_sent, next_reminder_due,
			   created_at, updated_at
		FROM manual_coordination_tasks 
		WHERE assigned_to_user_id = $1
		ORDER BY next_reminder_due ASC, created_at DESC
		LIMIT $2 OFFSET $3`

	return r.queryTasks(ctx, query, userID, limit, offset)
}

// GetPendingTasks retrieves pending manual coordination tasks
func (r *manualTaskRepository) GetPendingTasks(ctx context.Context, limit, offset int) ([]*entity.ManualCoordinationTask, error) {
	query := `
		SELECT id, delivery_id, provider_code, task_type, task_status,
			   assigned_to_user_id, task_instructions, contact_information,
			   completed_at, completion_notes, external_reference,
			   reminder_count, last_reminder_sent, next_reminder_due,
			   created_at, updated_at
		FROM manual_coordination_tasks 
		WHERE task_status = 'pending'
		ORDER BY next_reminder_due ASC, created_at ASC
		LIMIT $1 OFFSET $2`

	return r.queryTasks(ctx, query, limit, offset)
}

// GetActiveTasks retrieves active (pending + in_progress) manual coordination tasks
func (r *manualTaskRepository) GetActiveTasks(ctx context.Context, limit, offset int) ([]*entity.ManualCoordinationTask, error) {
	query := `
		SELECT id, delivery_id, provider_code, task_type, task_status,
			   assigned_to_user_id, task_instructions, contact_information,
			   completed_at, completion_notes, external_reference,
			   reminder_count, last_reminder_sent, next_reminder_due,
			   created_at, updated_at
		FROM manual_coordination_tasks 
		WHERE task_status IN ('pending', 'in_progress')
		ORDER BY next_reminder_due ASC, created_at ASC
		LIMIT $1 OFFSET $2`

	return r.queryTasks(ctx, query, limit, offset)
}

// GetOverdueTasks retrieves overdue manual coordination tasks
func (r *manualTaskRepository) GetOverdueTasks(ctx context.Context) ([]*entity.ManualCoordinationTask, error) {
	query := `
		SELECT id, delivery_id, provider_code, task_type, task_status,
			   assigned_to_user_id, task_instructions, contact_information,
			   completed_at, completion_notes, external_reference,
			   reminder_count, last_reminder_sent, next_reminder_due,
			   created_at, updated_at
		FROM manual_coordination_tasks 
		WHERE next_reminder_due < NOW() AND task_status IN ('pending', 'in_progress')
		ORDER BY next_reminder_due ASC`

	return r.queryTasks(ctx, query)
}

// GetTasksDueForReminder retrieves tasks that need reminders
func (r *manualTaskRepository) GetTasksDueForReminder(ctx context.Context) ([]*entity.ManualCoordinationTask, error) {
	query := `
		SELECT id, delivery_id, provider_code, task_type, task_status,
			   assigned_to_user_id, task_instructions, contact_information,
			   completed_at, completion_notes, external_reference,
			   reminder_count, last_reminder_sent, next_reminder_due,
			   created_at, updated_at
		FROM manual_coordination_tasks 
		WHERE next_reminder_due <= NOW() AND task_status IN ('pending', 'in_progress')
		ORDER BY next_reminder_due ASC`

	return r.queryTasks(ctx, query)
}

// AssignTask assigns a task to a user
func (r *manualTaskRepository) AssignTask(ctx context.Context, taskID, userID uuid.UUID) error {
	query := `
		UPDATE manual_coordination_tasks 
		SET assigned_to_user_id = $1, updated_at = NOW()
		WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, userID, taskID)
	if err != nil {
		return fmt.Errorf("failed to assign task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrManualTaskNotFound
	}

	return nil
}

// UnassignTask unassigns a task from a user
func (r *manualTaskRepository) UnassignTask(ctx context.Context, taskID uuid.UUID) error {
	query := `
		UPDATE manual_coordination_tasks 
		SET assigned_to_user_id = NULL, updated_at = NOW()
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, taskID)
	if err != nil {
		return fmt.Errorf("failed to unassign task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrManualTaskNotFound
	}

	return nil
}

// GetUnassignedTasks retrieves unassigned manual coordination tasks
func (r *manualTaskRepository) GetUnassignedTasks(ctx context.Context, limit, offset int) ([]*entity.ManualCoordinationTask, error) {
	query := `
		SELECT id, delivery_id, provider_code, task_type, task_status,
			   assigned_to_user_id, task_instructions, contact_information,
			   completed_at, completion_notes, external_reference,
			   reminder_count, last_reminder_sent, next_reminder_due,
			   created_at, updated_at
		FROM manual_coordination_tasks 
		WHERE assigned_to_user_id IS NULL AND task_status IN ('pending', 'in_progress')
		ORDER BY created_at ASC
		LIMIT $1 OFFSET $2`

	return r.queryTasks(ctx, query, limit, offset)
}

// CompleteTask marks a task as completed
func (r *manualTaskRepository) CompleteTask(ctx context.Context, taskID uuid.UUID, completionNotes, externalReference string) error {
	query := `
		UPDATE manual_coordination_tasks 
		SET task_status = 'completed', 
			completed_at = NOW(),
			completion_notes = $1,
			external_reference = $2,
			updated_at = NOW()
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, completionNotes, externalReference, taskID)
	if err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrManualTaskNotFound
	}

	return nil
}

// FailTask marks a task as failed
func (r *manualTaskRepository) FailTask(ctx context.Context, taskID uuid.UUID, reason string) error {
	query := `
		UPDATE manual_coordination_tasks 
		SET task_status = 'failed',
			completion_notes = $1,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, reason, taskID)
	if err != nil {
		return fmt.Errorf("failed to fail task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrManualTaskNotFound
	}

	return nil
}

// CancelTask marks a task as cancelled
func (r *manualTaskRepository) CancelTask(ctx context.Context, taskID uuid.UUID, reason string) error {
	query := `
		UPDATE manual_coordination_tasks 
		SET task_status = 'cancelled',
			completion_notes = $1,
			updated_at = NOW()
		WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, reason, taskID)
	if err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrManualTaskNotFound
	}

	return nil
}

// MarkReminderSent marks that a reminder was sent for a task
func (r *manualTaskRepository) MarkReminderSent(ctx context.Context, taskID uuid.UUID) error {
	query := `
		UPDATE manual_coordination_tasks 
		SET reminder_count = reminder_count + 1,
			last_reminder_sent = NOW(),
			updated_at = NOW()
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, taskID)
	if err != nil {
		return fmt.Errorf("failed to mark reminder sent: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrManualTaskNotFound
	}

	return nil
}

// SetNextReminder sets the next reminder time for a task
func (r *manualTaskRepository) SetNextReminder(ctx context.Context, taskID uuid.UUID, reminderTime time.Time) error {
	query := `
		UPDATE manual_coordination_tasks 
		SET next_reminder_due = $1,
			updated_at = NOW()
		WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, reminderTime, taskID)
	if err != nil {
		return fmt.Errorf("failed to set next reminder: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrManualTaskNotFound
	}

	return nil
}

// GetTasksForReminder retrieves tasks that need reminders before a specific time
func (r *manualTaskRepository) GetTasksForReminder(ctx context.Context, maxReminderTime time.Time) ([]*entity.ManualCoordinationTask, error) {
	query := `
		SELECT id, delivery_id, provider_code, task_type, task_status,
			   assigned_to_user_id, task_instructions, contact_information,
			   completed_at, completion_notes, external_reference,
			   reminder_count, last_reminder_sent, next_reminder_due,
			   created_at, updated_at
		FROM manual_coordination_tasks 
		WHERE next_reminder_due <= $1 AND task_status IN ('pending', 'in_progress')
		ORDER BY next_reminder_due ASC`

	return r.queryTasks(ctx, query, maxReminderTime)
}

// Helper method to query tasks and handle common scanning logic
func (r *manualTaskRepository) queryTasks(ctx context.Context, query string, args ...interface{}) ([]*entity.ManualCoordinationTask, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query manual coordination tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*entity.ManualCoordinationTask
	for rows.Next() {
		var task entity.ManualCoordinationTask
		var contactInfoJSON []byte

		err := rows.Scan(
			&task.ID, &task.DeliveryID, &task.ProviderCode, &task.TaskType, &task.TaskStatus,
			&task.AssignedToUserID, &task.TaskInstructions, &contactInfoJSON,
			&task.CompletedAt, &task.CompletionNotes, &task.ExternalReference,
			&task.ReminderCount, &task.LastReminderSent, &task.NextReminderDue,
			&task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan manual coordination task: %w", err)
		}

		// Unmarshal JSON fields
		if len(contactInfoJSON) > 0 {
			json.Unmarshal(contactInfoJSON, &task.ContactInformation)
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

// Provider-specific methods (implement stubs for remaining interface methods)
func (r *manualTaskRepository) GetInterExpressTasks(ctx context.Context, status entity.TaskStatus, limit, offset int) ([]*entity.ManualCoordinationTask, error) {
	return r.GetByProviderCode(ctx, "INTER_EXPRESS", limit, offset)
}

func (r *manualTaskRepository) GetNimExpressTasks(ctx context.Context, status entity.TaskStatus, limit, offset int) ([]*entity.ManualCoordinationTask, error) {
	return r.GetByProviderCode(ctx, "NIM_EXPRESS", limit, offset)
}

func (r *manualTaskRepository) GetRotRaoTasks(ctx context.Context, status entity.TaskStatus, limit, offset int) ([]*entity.ManualCoordinationTask, error) {
	return r.GetByProviderCode(ctx, "ROT_RAO", limit, offset)
}

// Stub implementations for remaining methods (implement as needed)
func (r *manualTaskRepository) GetTaskMetrics(ctx context.Context, startDate, endDate time.Time) (*repository.TaskMetrics, error) {
	// TODO: Implement task metrics calculation
	return &repository.TaskMetrics{}, nil
}

func (r *manualTaskRepository) GetTaskMetricsByProvider(ctx context.Context, providerCode string, startDate, endDate time.Time) (*repository.TaskMetrics, error) {
	// TODO: Implement provider-specific task metrics
	return &repository.TaskMetrics{}, nil
}

func (r *manualTaskRepository) GetTaskMetricsByUser(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*repository.TaskMetrics, error) {
	// TODO: Implement user-specific task metrics
	return &repository.TaskMetrics{}, nil
}

func (r *manualTaskRepository) GetAverageTaskCompletionTime(ctx context.Context, taskType entity.TaskType, startDate, endDate time.Time) (float64, error) {
	// TODO: Implement average completion time calculation
	return 0.0, nil
}

func (r *manualTaskRepository) SearchTasks(ctx context.Context, filters *repository.TaskQueryFilters) ([]*entity.ManualCoordinationTask, error) {
	// TODO: Implement complex search with filters
	return []*entity.ManualCoordinationTask{}, nil
}

func (r *manualTaskRepository) GetTasksByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.ManualCoordinationTask, error) {
	// TODO: Implement date range query
	return []*entity.ManualCoordinationTask{}, nil
}

func (r *manualTaskRepository) GetCompletedTasksByDate(ctx context.Context, date time.Time) ([]*entity.ManualCoordinationTask, error) {
	// TODO: Implement completed tasks by date
	return []*entity.ManualCoordinationTask{}, nil
}

func (r *manualTaskRepository) UpdateMultipleTaskStatuses(ctx context.Context, taskIDs []uuid.UUID, status entity.TaskStatus) error {
	// TODO: Implement bulk status update
	return nil
}

func (r *manualTaskRepository) GetTasksByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.ManualCoordinationTask, error) {
	// TODO: Implement bulk get by IDs
	return []*entity.ManualCoordinationTask{}, nil
}

func (r *manualTaskRepository) CreateBulkTasks(ctx context.Context, tasks []*entity.ManualCoordinationTask) error {
	// TODO: Implement bulk create
	return nil
}

func (r *manualTaskRepository) GetUserProductivity(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*repository.UserProductivityMetrics, error) {
	// TODO: Implement user productivity metrics
	return &repository.UserProductivityMetrics{}, nil
}

func (r *manualTaskRepository) GetProviderTaskEfficiency(ctx context.Context, providerCode string, startDate, endDate time.Time) (*repository.ProviderTaskEfficiency, error) {
	// TODO: Implement provider efficiency metrics
	return &repository.ProviderTaskEfficiency{}, nil
}

func (r *manualTaskRepository) GetTaskBacklog(ctx context.Context) (*repository.TaskBacklog, error) {
	// TODO: Implement task backlog calculation
	return &repository.TaskBacklog{}, nil
}
