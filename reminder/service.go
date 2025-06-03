package reminder

import (
	"fmt"
	"time"
)

// service implements the Service interface
type service struct {
	store Store
}

// NewService creates a new reminder service instance
func NewService(store Store) Service {
	return &service{store: store}
}

// Create implements Service.Create
func (s *service) Create(r *Reminder) error {
	if r.Title == "" {
		return fmt.Errorf("reminder title is required")
	}
	if r.DueTime.IsZero() {
		return fmt.Errorf("reminder due time is required")
	}
	if r.DueTime.Before(time.Now()) {
		return fmt.Errorf("reminder due time must be in the future")
	}

	// Set default status if not provided
	if r.Status == "" {
		r.Status = StatusPending
	}

	return s.store.CreateReminder(r)
}

// Get implements Service.Get
func (s *service) Get(id string) (*Reminder, error) {
	return s.store.GetReminder(id)
}

// List implements Service.List
func (s *service) List(userID string, filter ListFilter) ([]*Reminder, error) {
	return s.store.ListReminders(userID, filter)
}

// Update implements Service.Update
func (s *service) Update(r *Reminder) error {
	if r.Title == "" {
		return fmt.Errorf("reminder title is required")
	}
	if r.DueTime.IsZero() {
		return fmt.Errorf("reminder due time is required")
	}
	if r.DueTime.Before(time.Now()) {
		return fmt.Errorf("reminder due time must be in the future")
	}

	return s.store.UpdateReminder(r)
}

// Delete implements Service.Delete
func (s *service) Delete(id string) error {
	return s.store.DeleteReminder(id)
}

// Complete implements Service.Complete
func (s *service) Complete(id string) error {
	reminder, err := s.store.GetReminder(id)
	if err != nil {
		return err
	}

	reminder.Status = StatusCompleted
	return s.store.UpdateReminder(reminder)
}

// Cancel implements Service.Cancel
func (s *service) Cancel(id string) error {
	reminder, err := s.store.GetReminder(id)
	if err != nil {
		return err
	}

	reminder.Status = StatusCancelled
	return s.store.UpdateReminder(reminder)
}

// LogNotification implements Service.LogNotification
func (s *service) LogNotification(log *NotificationLog) error {
	return s.store.CreateNotificationLog(log)
}
