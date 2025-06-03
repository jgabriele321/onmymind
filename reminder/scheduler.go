package reminder

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Scheduler manages reminder scheduling and notifications
type Scheduler struct {
	service  Service
	bot      *tgbot.BotAPI
	location *time.Location
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewScheduler creates a new scheduler instance
func NewScheduler(service Service, bot *tgbot.BotAPI, location *time.Location) *Scheduler {
	if location == nil {
		location = time.UTC
	}
	return &Scheduler{
		service:  service,
		bot:      bot,
		location: location,
		stopChan: make(chan struct{}),
	}
}

// Start begins the scheduler
func (s *Scheduler) Start() {
	s.wg.Add(1)
	go s.run()
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopChan)
	s.wg.Wait()
}

func (s *Scheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkReminders()
		}
	}
}

func (s *Scheduler) checkReminders() {
	// Get all pending reminders
	filter := ListFilter{
		Status: &[]Status{StatusPending}[0],
		ToTime: &[]time.Time{time.Now().Add(time.Minute)}[0],
	}

	// We'll check all users' reminders
	reminders, err := s.service.List("", filter)
	if err != nil {
		log.Printf("Error fetching reminders: %v", err)
		return
	}

	for _, r := range reminders {
		// Skip if the reminder is not due yet
		if time.Now().Before(r.DueTime) {
			continue
		}

		// Send notification
		if err := s.sendNotification(r); err != nil {
			log.Printf("Error sending notification for reminder %s: %v", r.ID, err)
			continue
		}

		// Handle recurring reminders
		if r.RecurrencePattern != "" {
			if err := s.scheduleNextRecurrence(r); err != nil {
				log.Printf("Error scheduling next recurrence for reminder %s: %v", r.ID, err)
			}
		} else {
			// Mark one-time reminder as completed
			if err := s.service.Complete(r.ID); err != nil {
				log.Printf("Error completing reminder %s: %v", r.ID, err)
			}
		}
	}
}

func (s *Scheduler) sendNotification(r *Reminder) error {
	userID, err := parseUserID(r.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	// Send initial message
	msg := tgbot.NewMessage(userID, formatReminderNotification(r))
	if _, err := s.bot.Send(msg); err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	// Log the notification
	notifLog := &NotificationLog{
		ReminderID:       r.ID,
		NotificationType: NotificationTelegramMessage,
		Status:           "success",
		AttemptedAt:      time.Now(),
	}
	if err := s.service.LogNotification(notifLog); err != nil {
		log.Printf("Error logging notification: %v", err)
	}

	// For priority reminders, make a call after a delay if no response
	if r.Priority {
		go func() {
			// Wait for 2 minutes before making the call
			time.Sleep(2 * time.Minute)

			// Check if the reminder is still pending
			reminder, err := s.service.Get(r.ID)
			if err != nil || reminder.Status != StatusPending {
				return
			}

			// Send a call notification
			callMsg := tgbot.NewMessage(userID,
				fmt.Sprintf("⚠️ Priority Reminder: %s", r.Title))
			if _, err := s.bot.Send(callMsg); err != nil {
				log.Printf("Error sending call notification: %v", err)
				return
			}

			// Log the call attempt
			callLog := &NotificationLog{
				ReminderID:       r.ID,
				NotificationType: NotificationTelegramCall,
				Status:           "success",
				AttemptedAt:      time.Now(),
			}
			if err := s.service.LogNotification(callLog); err != nil {
				log.Printf("Error logging call notification: %v", err)
			}
		}()
	}

	return nil
}

func (s *Scheduler) scheduleNextRecurrence(r *Reminder) error {
	nextTime, err := s.calculateNextOccurrence(r)
	if err != nil {
		return err
	}

	// Create a new reminder for the next occurrence
	nextReminder := &Reminder{
		UserID:            r.UserID,
		Title:             r.Title,
		Description:       r.Description,
		DueTime:           nextTime,
		RecurrencePattern: r.RecurrencePattern,
		Priority:          r.Priority,
		Status:            StatusPending,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Mark the current reminder as completed
	if err := s.service.Complete(r.ID); err != nil {
		return fmt.Errorf("failed to complete current reminder: %v", err)
	}

	// Create the next reminder
	return s.service.Create(nextReminder)
}

func (s *Scheduler) calculateNextOccurrence(r *Reminder) (time.Time, error) {
	parts := strings.SplitN(r.RecurrencePattern, ":", 2)
	if len(parts) != 2 && parts[0] != "daily" && parts[0] != "weekday" {
		return time.Time{}, fmt.Errorf("invalid recurrence pattern: %s", r.RecurrencePattern)
	}

	base := r.DueTime
	now := time.Now()

	switch parts[0] {
	case "daily":
		next := base.AddDate(0, 0, 1)
		for next.Before(now) {
			next = next.AddDate(0, 0, 1)
		}
		return next, nil

	case "weekday":
		next := base.AddDate(0, 0, 1)
		for next.Before(now) || next.Weekday() == time.Saturday || next.Weekday() == time.Sunday {
			next = next.AddDate(0, 0, 1)
		}
		return next, nil

	case "weekly":
		days := strings.Split(parts[1], ",")
		weekdays := make(map[time.Weekday]bool)
		for _, day := range days {
			weekdays[parseWeekday(day)] = true
		}

		next := base.AddDate(0, 0, 1)
		for next.Before(now) || !weekdays[next.Weekday()] {
			next = next.AddDate(0, 0, 1)
		}
		return next, nil

	case "monthly":
		daySpec := parts[1]
		next := base.AddDate(0, 1, 0)
		for next.Before(now) {
			next = next.AddDate(0, 1, 0)
		}

		switch daySpec {
		case "first":
			next = time.Date(next.Year(), next.Month(), 1,
				base.Hour(), base.Minute(), 0, 0, s.location)
		case "last":
			next = time.Date(next.Year(), next.Month()+1, 0,
				base.Hour(), base.Minute(), 0, 0, s.location)
		default:
			day := parseMonthDay(daySpec)
			next = time.Date(next.Year(), next.Month(), day,
				base.Hour(), base.Minute(), 0, 0, s.location)
		}
		return next, nil

	default:
		return time.Time{}, fmt.Errorf("unsupported recurrence pattern: %s", r.RecurrencePattern)
	}
}

func parseUserID(userID string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(userID, "%d", &id)
	return id, err
}

func parseWeekday(day string) time.Weekday {
	switch strings.ToLower(day) {
	case "sunday":
		return time.Sunday
	case "monday":
		return time.Monday
	case "tuesday":
		return time.Tuesday
	case "wednesday":
		return time.Wednesday
	case "thursday":
		return time.Thursday
	case "friday":
		return time.Friday
	case "saturday":
		return time.Saturday
	default:
		return time.Sunday
	}
}

func parseMonthDay(daySpec string) int {
	var day int
	fmt.Sscanf(daySpec, "%d", &day)
	if day < 1 {
		day = 1
	} else if day > 28 {
		day = 28
	}
	return day
}

func formatReminderNotification(r *Reminder) string {
	var sb strings.Builder
	sb.WriteString("🔔 Reminder!\n\n")
	sb.WriteString(r.Title)
	if r.Description != "" {
		sb.WriteString("\n\n")
		sb.WriteString(r.Description)
	}
	if r.Priority {
		sb.WriteString("\n\n⭐ This is a priority reminder!")
	}
	if r.RecurrencePattern != "" {
		sb.WriteString("\n\n🔄 This reminder will recur.")
	}
	return sb.String()
}
