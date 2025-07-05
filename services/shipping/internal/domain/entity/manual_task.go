package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ManualCoordinationTask represents a manual coordination task for manual providers
type ManualCoordinationTask struct {
	ID           uuid.UUID `json:"id"`
	DeliveryID   uuid.UUID `json:"delivery_id"`
	ProviderCode string    `json:"provider_code"`
	TaskType     TaskType  `json:"task_type"`
	TaskStatus   TaskStatus `json:"task_status"`
	
	// Task Details
	AssignedToUserID    *uuid.UUID             `json:"assigned_to_user_id"`
	TaskInstructions    string                 `json:"task_instructions"`
	ContactInformation  map[string]interface{} `json:"contact_information"`
	
	// Completion Data
	CompletedAt         *time.Time `json:"completed_at"`
	CompletionNotes     string     `json:"completion_notes"`
	ExternalReference   string     `json:"external_reference"` // Tracking number from provider
	
	// Reminder System
	ReminderCount       int        `json:"reminder_count"`
	LastReminderSent    *time.Time `json:"last_reminder_sent"`
	NextReminderDue     *time.Time `json:"next_reminder_due"`
	
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// TaskType represents the type of manual coordination task
type TaskType string

const (
	TaskTypePhoneCoordination TaskType = "phone_coordination"
	TaskTypeAppBooking        TaskType = "app_booking"
	TaskTypeLineMessage       TaskType = "line_message"
	TaskTypeEmailCoordination TaskType = "email_coordination"
	TaskTypePickupSchedule    TaskType = "pickup_schedule"
)

// TaskStatus represents the status of a manual coordination task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// Domain errors
var (
	ErrTaskInvalidDeliveryID    = errors.New("delivery ID cannot be empty")
	ErrTaskInvalidProviderCode  = errors.New("provider code cannot be empty")
	ErrTaskInvalidType          = errors.New("invalid task type")
	ErrTaskInvalidStatus        = errors.New("invalid task status")
	ErrTaskInvalidInstructions  = errors.New("task instructions cannot be empty")
	ErrTaskAlreadyCompleted     = errors.New("task is already completed")
	ErrTaskNotPending           = errors.New("task is not in pending status")
	ErrTaskInvalidCompletion    = errors.New("completion notes required for completed task")
	ErrTaskInvalidReminder      = errors.New("reminder time must be in the future")
)

// NewManualCoordinationTask creates a new manual coordination task with validation
func NewManualCoordinationTask(
	deliveryID uuid.UUID,
	providerCode string,
	taskType TaskType,
	instructions string,
	contactInfo map[string]interface{},
) (*ManualCoordinationTask, error) {
	if err := validateTaskDeliveryID(deliveryID); err != nil {
		return nil, err
	}
	
	if err := validateTaskProviderCode(providerCode); err != nil {
		return nil, err
	}
	
	if err := validateTaskType(taskType); err != nil {
		return nil, err
	}
	
	if err := validateTaskInstructions(instructions); err != nil {
		return nil, err
	}
	
	now := time.Now()
	reminderTime := now.Add(getDefaultReminderInterval(taskType))
	
	return &ManualCoordinationTask{
		ID:                  uuid.New(),
		DeliveryID:          deliveryID,
		ProviderCode:        providerCode,
		TaskType:            taskType,
		TaskStatus:          TaskStatusPending,
		TaskInstructions:    instructions,
		ContactInformation:  contactInfo,
		ReminderCount:       0,
		NextReminderDue:     &reminderTime,
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
}

// AssignToUser assigns the task to a specific user
func (t *ManualCoordinationTask) AssignToUser(userID uuid.UUID) error {
	if t.TaskStatus != TaskStatusPending {
		return ErrTaskNotPending
	}
	
	t.AssignedToUserID = &userID
	t.TaskStatus = TaskStatusInProgress
	t.UpdatedAt = time.Now()
	
	return nil
}

// CompleteTask marks the task as completed
func (t *ManualCoordinationTask) CompleteTask(completionNotes string, externalReference string) error {
	if t.TaskStatus == TaskStatusCompleted {
		return ErrTaskAlreadyCompleted
	}
	
	if completionNotes == "" {
		return ErrTaskInvalidCompletion
	}
	
	now := time.Now()
	t.TaskStatus = TaskStatusCompleted
	t.CompletedAt = &now
	t.CompletionNotes = completionNotes
	t.ExternalReference = externalReference
	t.NextReminderDue = nil // Clear reminder
	t.UpdatedAt = now
	
	return nil
}

// FailTask marks the task as failed
func (t *ManualCoordinationTask) FailTask(reason string) error {
	if t.TaskStatus == TaskStatusCompleted {
		return ErrTaskAlreadyCompleted
	}
	
	now := time.Now()
	t.TaskStatus = TaskStatusFailed
	t.CompletionNotes = reason
	t.CompletedAt = &now
	t.UpdatedAt = now
	
	return nil
}

// CancelTask cancels the task
func (t *ManualCoordinationTask) CancelTask(reason string) error {
	if t.TaskStatus == TaskStatusCompleted {
		return ErrTaskAlreadyCompleted
	}
	
	now := time.Now()
	t.TaskStatus = TaskStatusCancelled
	t.CompletionNotes = reason
	t.CompletedAt = &now
	t.NextReminderDue = nil // Clear reminder
	t.UpdatedAt = now
	
	return nil
}

// SendReminder records that a reminder was sent
func (t *ManualCoordinationTask) SendReminder() error {
	if t.TaskStatus != TaskStatusPending && t.TaskStatus != TaskStatusInProgress {
		return errors.New("cannot send reminder for non-active task")
	}
	
	now := time.Now()
	t.ReminderCount++
	t.LastReminderSent = &now
	t.NextReminderDue = &time.Time{}
	*t.NextReminderDue = now.Add(getReminderInterval(t.TaskType, t.ReminderCount))
	t.UpdatedAt = now
	
	return nil
}

// SetNextReminder sets the next reminder time
func (t *ManualCoordinationTask) SetNextReminder(reminderTime time.Time) error {
	if reminderTime.Before(time.Now()) {
		return ErrTaskInvalidReminder
	}
	
	t.NextReminderDue = &reminderTime
	t.UpdatedAt = time.Now()
	
	return nil
}

// IsOverdue checks if the task is overdue
func (t *ManualCoordinationTask) IsOverdue() bool {
	if t.TaskStatus == TaskStatusCompleted || t.TaskStatus == TaskStatusCancelled || t.TaskStatus == TaskStatusFailed {
		return false
	}
	
	// Consider task overdue if created more than defined threshold ago
	threshold := getOverdueThreshold(t.TaskType)
	return time.Since(t.CreatedAt) > threshold
}

// NeedsReminder checks if a reminder is due
func (t *ManualCoordinationTask) NeedsReminder() bool {
	if t.TaskStatus != TaskStatusPending && t.TaskStatus != TaskStatusInProgress {
		return false
	}
	
	if t.NextReminderDue == nil {
		return false
	}
	
	return time.Now().After(*t.NextReminderDue)
}

// GetDuration calculates how long the task has been active
func (t *ManualCoordinationTask) GetDuration() time.Duration {
	if t.CompletedAt != nil {
		return t.CompletedAt.Sub(t.CreatedAt)
	}
	return time.Since(t.CreatedAt)
}

// IsActive returns true if the task is still active (pending or in progress)
func (t *ManualCoordinationTask) IsActive() bool {
	return t.TaskStatus == TaskStatusPending || t.TaskStatus == TaskStatusInProgress
}

// GetContactInfo safely gets contact information
func (t *ManualCoordinationTask) GetContactInfo(key string) interface{} {
	if t.ContactInformation == nil {
		return nil
	}
	return t.ContactInformation[key]
}

// GetContactPhone gets phone number from contact information
func (t *ManualCoordinationTask) GetContactPhone() string {
	if phone, ok := t.GetContactInfo("phone").(string); ok {
		return phone
	}
	return ""
}

// GetContactLineID gets LINE ID from contact information
func (t *ManualCoordinationTask) GetContactLineID() string {
	if lineID, ok := t.GetContactInfo("line_id").(string); ok {
		return lineID
	}
	return ""
}

// GetContactEmail gets email from contact information
func (t *ManualCoordinationTask) GetContactEmail() string {
	if email, ok := t.GetContactInfo("email").(string); ok {
		return email
	}
	return ""
}

// UpdateInstructions updates the task instructions
func (t *ManualCoordinationTask) UpdateInstructions(instructions string) error {
	if instructions == "" {
		return ErrTaskInvalidInstructions
	}
	
	t.TaskInstructions = instructions
	t.UpdatedAt = time.Now()
	
	return nil
}

// AddContactInfo adds or updates contact information
func (t *ManualCoordinationTask) AddContactInfo(key string, value interface{}) {
	if t.ContactInformation == nil {
		t.ContactInformation = make(map[string]interface{})
	}
	
	t.ContactInformation[key] = value
	t.UpdatedAt = time.Now()
}

// Helper functions for reminder intervals and thresholds
func getDefaultReminderInterval(taskType TaskType) time.Duration {
	switch taskType {
	case TaskTypePhoneCoordination:
		return 30 * time.Minute
	case TaskTypeAppBooking:
		return 2 * time.Hour
	case TaskTypeLineMessage:
		return 1 * time.Hour
	case TaskTypeEmailCoordination:
		return 4 * time.Hour
	case TaskTypePickupSchedule:
		return 6 * time.Hour
	default:
		return 1 * time.Hour
	}
}

func getReminderInterval(taskType TaskType, reminderCount int) time.Duration {
	baseInterval := getDefaultReminderInterval(taskType)
	
	// Exponential backoff with cap
	multiplier := 1
	for i := 0; i < reminderCount && multiplier < 8; i++ {
		multiplier *= 2
	}
	
	return time.Duration(multiplier) * baseInterval
}

func getOverdueThreshold(taskType TaskType) time.Duration {
	switch taskType {
	case TaskTypePhoneCoordination:
		return 4 * time.Hour
	case TaskTypeAppBooking:
		return 12 * time.Hour
	case TaskTypeLineMessage:
		return 6 * time.Hour
	case TaskTypeEmailCoordination:
		return 24 * time.Hour
	case TaskTypePickupSchedule:
		return 24 * time.Hour
	default:
		return 8 * time.Hour
	}
}

// Validation functions
func validateTaskDeliveryID(id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrTaskInvalidDeliveryID
	}
	return nil
}

func validateTaskProviderCode(code string) error {
	if code == "" {
		return ErrTaskInvalidProviderCode
	}
	return nil
}

func validateTaskType(taskType TaskType) error {
	validTypes := map[TaskType]bool{
		TaskTypePhoneCoordination: true,
		TaskTypeAppBooking:        true,
		TaskTypeLineMessage:       true,
		TaskTypeEmailCoordination: true,
		TaskTypePickupSchedule:    true,
	}
	
	if !validTypes[taskType] {
		return ErrTaskInvalidType
	}
	return nil
}

func validateTaskInstructions(instructions string) error {
	if instructions == "" {
		return ErrTaskInvalidInstructions
	}
	return nil
}
