package reminder

import (
	"fmt"
	"log"
	"strings"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Handler manages reminder-related commands
type Handler struct {
	service    Service
	timeParser *TimeParser
	location   *time.Location
}

// NewHandler creates a new reminder handler
func NewHandler(service Service, location *time.Location) *Handler {
	if location == nil {
		location = time.UTC
	}
	return &Handler{
		service:    service,
		timeParser: NewTimeParser(location),
		location:   location,
	}
}

// HandleRemindMe handles the /remindme command
func (h *Handler) HandleRemindMe(msg *tgbot.Message) (string, error) {
	args := msg.CommandArguments()
	if args == "" {
		return "Usage: /remindme <time> to <message> [-call]\nExamples:\n" +
			"• /remindme in 2 hours to check email\n" +
			"• /remindme tomorrow at 3pm to call mom -call\n" +
			"• /remindme every Sunday at 10am to water plants\n" +
			"• /remindme 2024-03-20 15:00 to submit report", nil
	}

	// Check if this is a recurring reminder
	if strings.HasPrefix(strings.ToLower(args), "every") {
		return h.handleRecurringReminder(msg.From.ID, args)
	}

	// Parse the command
	dueTime, title, isPriority, err := h.timeParser.ParseCommand(args)
	if err != nil {
		return fmt.Sprintf("❌ Error: %v", err), nil
	}

	// Create the reminder
	reminder := &Reminder{
		UserID:    fmt.Sprintf("%d", msg.From.ID),
		Title:     title,
		DueTime:   dueTime,
		Priority:  isPriority,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.service.Create(reminder); err != nil {
		log.Printf("Error creating reminder: %v", err)
		return "❌ Failed to create reminder", err
	}

	// Format response
	response := fmt.Sprintf("✅ Reminder set for %s\n%s",
		formatTime(dueTime),
		formatReminder(reminder))

	return response, nil
}

// HandleReminders handles the /reminders command
func (h *Handler) HandleReminders(msg *tgbot.Message) (string, error) {
	args := strings.ToLower(msg.CommandArguments())
	userID := fmt.Sprintf("%d", msg.From.ID)

	filter := ListFilter{}
	switch args {
	case "priority", "call":
		priority := true
		filter.Priority = &priority
	case "regular":
		priority := false
		filter.Priority = &priority
	}

	reminders, err := h.service.List(userID, filter)
	if err != nil {
		log.Printf("Error listing reminders: %v", err)
		return "❌ Failed to list reminders", err
	}

	if len(reminders) == 0 {
		return "No reminders found", nil
	}

	// Group reminders by status
	pending := make([]*Reminder, 0)
	completed := make([]*Reminder, 0)
	for _, r := range reminders {
		switch r.Status {
		case StatusPending:
			pending = append(pending, r)
		case StatusCompleted:
			completed = append(completed, r)
		}
	}

	// Build response
	var sb strings.Builder
	sb.WriteString("📅 Your Reminders\n\n")

	if len(pending) > 0 {
		sb.WriteString("Pending:\n")
		for _, r := range pending {
			sb.WriteString(formatReminder(r))
			sb.WriteString("\n")
		}
	}

	if len(completed) > 0 {
		if len(pending) > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("Completed:\n")
		for _, r := range completed {
			sb.WriteString(formatReminder(r))
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

// HandleDelete handles the /delete command
func (h *Handler) HandleDelete(msg *tgbot.Message) (string, error) {
	id := msg.CommandArguments()
	if id == "" {
		return "Usage: /delete <reminder_id>", nil
	}

	if err := h.service.Delete(id); err != nil {
		log.Printf("Error deleting reminder: %v", err)
		return fmt.Sprintf("❌ Failed to delete reminder: %v", err), nil
	}

	return "✅ Reminder deleted", nil
}

// HandleComplete handles the /complete command
func (h *Handler) HandleComplete(msg *tgbot.Message) (string, error) {
	id := msg.CommandArguments()
	if id == "" {
		return "Usage: /complete <reminder_id>", nil
	}

	if err := h.service.Complete(id); err != nil {
		log.Printf("Error completing reminder: %v", err)
		return fmt.Sprintf("❌ Failed to complete reminder: %v", err), nil
	}

	return "✅ Reminder marked as completed", nil
}

// Helper functions

func (h *Handler) handleRecurringReminder(userID int64, args string) (string, error) {
	// Split into pattern and message
	parts := strings.SplitN(args, " to ", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid format: use 'every <schedule> at <time> to <message>'")
	}

	pattern := parts[0]
	title := parts[1]

	// Check for priority flag
	isPriority := false
	if strings.HasSuffix(title, "-call") {
		isPriority = true
		title = strings.TrimSuffix(strings.TrimSpace(title), "-call")
	}

	// Parse the recurrence pattern
	recurrencePattern, nextTime, err := h.timeParser.ParseRecurrencePattern(pattern)
	if err != nil {
		return "", err
	}

	// Create the reminder
	reminder := &Reminder{
		UserID:            fmt.Sprintf("%d", userID),
		Title:             title,
		DueTime:           nextTime,
		RecurrencePattern: recurrencePattern,
		Priority:          isPriority,
		Status:            StatusPending,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := h.service.Create(reminder); err != nil {
		log.Printf("Error creating recurring reminder: %v", err)
		return "", err
	}

	return fmt.Sprintf("✅ Recurring reminder set\n%s", formatReminder(reminder)), nil
}

func formatTime(t time.Time) string {
	return fmt.Sprintf("%s (%s)",
		t.Format("Mon, Jan 2 at 3:04 PM"),
		t.Format("2006-01-02 15:04"))
}

func formatReminder(r *Reminder) string {
	var sb strings.Builder

	// Format the basic reminder info
	sb.WriteString(fmt.Sprintf("🔔 [%s] %s\n", r.ID[:8], r.Title))
	sb.WriteString(fmt.Sprintf("   📅 %s\n", formatTime(r.DueTime)))

	// Add recurrence info if present
	if r.RecurrencePattern != "" {
		sb.WriteString(fmt.Sprintf("   🔄 %s\n", r.RecurrencePattern))
	}

	// Add priority indicator
	if r.Priority {
		sb.WriteString("   ⭐ Priority\n")
	}

	return sb.String()
}
