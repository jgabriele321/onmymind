package reminder

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SQLiteStore implements the Store interface using SQLite
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite store instance
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// CreateReminder implements Store.CreateReminder
func (s *SQLiteStore) CreateReminder(r *Reminder) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now()
	}
	r.UpdatedAt = time.Now()

	query := `
		INSERT INTO reminders (
			id, user_id, title, description, due_time, 
			recurrence_pattern, priority, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query,
		r.ID, r.UserID, r.Title, r.Description, r.DueTime,
		r.RecurrencePattern, r.Priority, r.Status, r.CreatedAt, r.UpdatedAt)

	return err
}

// GetReminder implements Store.GetReminder
func (s *SQLiteStore) GetReminder(id string) (*Reminder, error) {
	r := &Reminder{}
	query := `
		SELECT id, user_id, title, description, due_time,
			   recurrence_pattern, priority, status, created_at, updated_at
		FROM reminders WHERE id = ?`

	err := s.db.QueryRow(query, id).Scan(
		&r.ID, &r.UserID, &r.Title, &r.Description, &r.DueTime,
		&r.RecurrencePattern, &r.Priority, &r.Status, &r.CreatedAt, &r.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("reminder not found: %s", id)
	}
	return r, err
}

// ListReminders implements Store.ListReminders
func (s *SQLiteStore) ListReminders(userID string, filter ListFilter) ([]*Reminder, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "user_id = ?")
	args = append(args, userID)

	if filter.Status != nil {
		conditions = append(conditions, "status = ?")
		args = append(args, *filter.Status)
	}

	if filter.Priority != nil {
		conditions = append(conditions, "priority = ?")
		args = append(args, *filter.Priority)
	}

	if filter.FromTime != nil {
		conditions = append(conditions, "due_time >= ?")
		args = append(args, *filter.FromTime)
	}

	if filter.ToTime != nil {
		conditions = append(conditions, "due_time <= ?")
		args = append(args, *filter.ToTime)
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, title, description, due_time,
			   recurrence_pattern, priority, status, created_at, updated_at
		FROM reminders
		WHERE %s
		ORDER BY due_time ASC`, strings.Join(conditions, " AND "))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []*Reminder
	for rows.Next() {
		r := &Reminder{}
		err := rows.Scan(
			&r.ID, &r.UserID, &r.Title, &r.Description, &r.DueTime,
			&r.RecurrencePattern, &r.Priority, &r.Status, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, err
		}
		reminders = append(reminders, r)
	}

	return reminders, rows.Err()
}

// UpdateReminder implements Store.UpdateReminder
func (s *SQLiteStore) UpdateReminder(r *Reminder) error {
	r.UpdatedAt = time.Now()

	query := `
		UPDATE reminders
		SET user_id = ?, title = ?, description = ?, due_time = ?,
			recurrence_pattern = ?, priority = ?, status = ?, updated_at = ?
		WHERE id = ?`

	result, err := s.db.Exec(query,
		r.UserID, r.Title, r.Description, r.DueTime,
		r.RecurrencePattern, r.Priority, r.Status, r.UpdatedAt,
		r.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("reminder not found: %s", r.ID)
	}

	return nil
}

// DeleteReminder implements Store.DeleteReminder
func (s *SQLiteStore) DeleteReminder(id string) error {
	result, err := s.db.Exec("DELETE FROM reminders WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("reminder not found: %s", id)
	}

	return nil
}

// CreateNotificationLog implements Store.CreateNotificationLog
func (s *SQLiteStore) CreateNotificationLog(log *NotificationLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	if log.AttemptedAt.IsZero() {
		log.AttemptedAt = time.Now()
	}

	query := `
		INSERT INTO reminder_logs (
			id, reminder_id, notification_type, status, error_message, attempted_at
		) VALUES (?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query,
		log.ID, log.ReminderID, log.NotificationType,
		log.Status, log.ErrorMessage, log.AttemptedAt)

	return err
}

// GetNotificationLogs implements Store.GetNotificationLogs
func (s *SQLiteStore) GetNotificationLogs(reminderID string) ([]*NotificationLog, error) {
	query := `
		SELECT id, reminder_id, notification_type, status, error_message, attempted_at
		FROM reminder_logs
		WHERE reminder_id = ?
		ORDER BY attempted_at DESC`

	rows, err := s.db.Query(query, reminderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*NotificationLog
	for rows.Next() {
		log := &NotificationLog{}
		err := rows.Scan(
			&log.ID, &log.ReminderID, &log.NotificationType,
			&log.Status, &log.ErrorMessage, &log.AttemptedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}
