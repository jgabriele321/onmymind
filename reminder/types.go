package reminder

import (
	"time"
)

// Status represents the current state of a reminder
type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
)

// NotificationType represents the type of notification to send
type NotificationType string

const (
	NotificationTelegramMessage NotificationType = "telegram_message"
	NotificationTelegramCall    NotificationType = "telegram_call"
)

// Reminder represents a single reminder instance
type Reminder struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	Title             string    `json:"title"`
	Description       string    `json:"description,omitempty"`
	DueTime           time.Time `json:"due_time"`
	RecurrencePattern string    `json:"recurrence_pattern,omitempty"`
	Priority          bool      `json:"priority"`
	Status            Status    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// NotificationLog represents a log entry for a notification attempt
type NotificationLog struct {
	ID               string           `json:"id"`
	ReminderID       string           `json:"reminder_id"`
	NotificationType NotificationType `json:"notification_type"`
	Status           string           `json:"status"`
	ErrorMessage     string           `json:"error_message,omitempty"`
	AttemptedAt      time.Time        `json:"attempted_at"`
}

// Service defines the interface for reminder operations
type Service interface {
	// Create creates a new reminder
	Create(reminder *Reminder) error

	// Get retrieves a reminder by ID
	Get(id string) (*Reminder, error)

	// List retrieves reminders based on filters
	List(userID string, filter ListFilter) ([]*Reminder, error)

	// Update updates an existing reminder
	Update(reminder *Reminder) error

	// Delete deletes a reminder
	Delete(id string) error

	// Complete marks a reminder as completed
	Complete(id string) error

	// Cancel marks a reminder as cancelled
	Cancel(id string) error

	// LogNotification logs a notification attempt
	LogNotification(log *NotificationLog) error
}

// ListFilter defines filters for listing reminders
type ListFilter struct {
	Status   *Status    // Filter by status
	Priority *bool      // Filter by priority
	FromTime *time.Time // Filter by due time range start
	ToTime   *time.Time // Filter by due time range end
}

// Store defines the interface for reminder persistence
type Store interface {
	// Reminder operations
	CreateReminder(reminder *Reminder) error
	GetReminder(id string) (*Reminder, error)
	ListReminders(userID string, filter ListFilter) ([]*Reminder, error)
	UpdateReminder(reminder *Reminder) error
	DeleteReminder(id string) error

	// Notification log operations
	CreateNotificationLog(log *NotificationLog) error
	GetNotificationLogs(reminderID string) ([]*NotificationLog, error)
}
