package reminder

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TimeParser handles parsing of natural language time expressions
type TimeParser struct {
	location *time.Location
}

// NewTimeParser creates a new TimeParser instance
func NewTimeParser(location *time.Location) *TimeParser {
	if location == nil {
		location = time.UTC
	}
	return &TimeParser{location: location}
}

// ParseCommand parses a reminder command into its components
func (p *TimeParser) ParseCommand(input string) (time.Time, string, bool, error) {
	// Check for priority flag
	isPriority := false
	if strings.HasSuffix(input, "-call") {
		isPriority = true
		input = strings.TrimSuffix(strings.TrimSpace(input), "-call")
	}

	// Find the first occurrence of "to" or "that" which separates time from message
	parts := strings.SplitN(input, " to ", 2)
	if len(parts) != 2 {
		parts = strings.SplitN(input, " that ", 2)
		if len(parts) != 2 {
			return time.Time{}, "", false, fmt.Errorf("invalid format: use '/remindme <time> to <message>'")
		}
	}

	timeStr := strings.TrimSpace(parts[0])
	message := strings.TrimSpace(parts[1])

	// Parse the time expression
	dueTime, err := p.ParseTimeExpression(timeStr)
	if err != nil {
		return time.Time{}, "", false, err
	}

	return dueTime, message, isPriority, nil
}

// ParseTimeExpression parses various time formats into a time.Time
func (p *TimeParser) ParseTimeExpression(input string) (time.Time, error) {
	input = strings.ToLower(strings.TrimSpace(input))

	// Try parsing as a recurring pattern first
	if strings.HasPrefix(input, "every") {
		return time.Time{}, fmt.Errorf("recurring reminders should be handled separately")
	}

	// Handle relative time expressions
	if strings.HasPrefix(input, "in") {
		return p.parseRelativeTime(strings.TrimPrefix(input, "in"))
	}

	// Handle "tomorrow at X" format
	if strings.HasPrefix(input, "tomorrow") {
		timeStr := strings.TrimPrefix(input, "tomorrow")
		timeStr = strings.TrimPrefix(timeStr, " at")
		return p.parseTomorrowTime(timeStr)
	}

	// Try parsing as absolute time
	return p.parseAbsoluteTime(input)
}

// ParseRecurrencePattern parses recurring time patterns
func (p *TimeParser) ParseRecurrencePattern(input string) (string, time.Time, error) {
	input = strings.ToLower(strings.TrimSpace(input))
	if !strings.HasPrefix(input, "every") {
		return "", time.Time{}, fmt.Errorf("recurrence pattern must start with 'every'")
	}

	pattern := strings.TrimPrefix(input, "every ")
	parts := strings.Split(pattern, " at ")
	if len(parts) != 2 {
		return "", time.Time{}, fmt.Errorf("invalid format: must include time with 'at'")
	}

	schedule := parts[0]
	timeStr := parts[1]

	// Parse the time portion
	t, err := p.parseTimeOfDay(timeStr)
	if err != nil {
		return "", time.Time{}, err
	}

	// Validate and format the schedule
	pattern, err = p.validateSchedule(schedule)
	if err != nil {
		return "", time.Time{}, err
	}

	return pattern, t, nil
}

func (p *TimeParser) parseRelativeTime(input string) (time.Time, error) {
	input = strings.TrimSpace(input)
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid relative time format")
	}

	amount, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid number in duration: %v", err)
	}

	unit := parts[1]
	now := time.Now().In(p.location)

	switch strings.ToLower(strings.TrimSuffix(unit, "s")) {
	case "minute":
		return now.Add(time.Duration(amount) * time.Minute), nil
	case "hour":
		return now.Add(time.Duration(amount) * time.Hour), nil
	case "day":
		return now.AddDate(0, 0, amount), nil
	case "week":
		return now.AddDate(0, 0, amount*7), nil
	case "month":
		return now.AddDate(0, amount, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported time unit: %s", unit)
	}
}

func (p *TimeParser) parseTomorrowTime(timeStr string) (time.Time, error) {
	now := time.Now().In(p.location)
	tomorrow := now.AddDate(0, 0, 1)

	t, err := p.parseTimeOfDay(strings.TrimSpace(timeStr))
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(
		tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
		t.Hour(), t.Minute(), 0, 0, p.location,
	), nil
}

func (p *TimeParser) parseAbsoluteTime(input string) (time.Time, error) {
	// Try parsing common formats
	formats := []string{
		"2006-01-02 15:04",
		"15:04",
		"3:04pm",
		"3:04 pm",
		"3pm",
		"3 pm",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, input, p.location); err == nil {
			now := time.Now().In(p.location)
			if format == "15:04" || strings.Contains(format, "pm") {
				// For time-only formats, use today's date
				return time.Date(
					now.Year(), now.Month(), now.Day(),
					t.Hour(), t.Minute(), 0, 0, p.location,
				), nil
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported time format: %s", input)
}

func (p *TimeParser) parseTimeOfDay(input string) (time.Time, error) {
	input = strings.ToLower(strings.TrimSpace(input))
	now := time.Now().In(p.location)

	// Try parsing with AM/PM
	if t, err := time.Parse("3:04pm", input); err == nil {
		return time.Date(
			now.Year(), now.Month(), now.Day(),
			t.Hour(), t.Minute(), 0, 0, p.location,
		), nil
	}
	if t, err := time.Parse("3pm", input); err == nil {
		return time.Date(
			now.Year(), now.Month(), now.Day(),
			t.Hour(), 0, 0, 0, p.location,
		), nil
	}

	// Try 24-hour format
	if t, err := time.Parse("15:04", input); err == nil {
		return time.Date(
			now.Year(), now.Month(), now.Day(),
			t.Hour(), t.Minute(), 0, 0, p.location,
		), nil
	}

	return time.Time{}, fmt.Errorf("invalid time format: %s", input)
}

func (p *TimeParser) validateSchedule(schedule string) (string, error) {
	schedule = strings.TrimSpace(schedule)

	// Handle "day" or "daily"
	if schedule == "day" || schedule == "daily" {
		return "daily", nil
	}

	// Handle "weekday"
	if schedule == "weekday" {
		return "weekday", nil
	}

	// Handle "month on the first/last/etc"
	if strings.HasPrefix(schedule, "month") {
		match := regexp.MustCompile(`month on the (first|last|\d+(?:st|nd|rd|th))`).FindStringSubmatch(schedule)
		if match != nil {
			return fmt.Sprintf("monthly:%s", match[1]), nil
		}
		return "", fmt.Errorf("invalid monthly schedule format")
	}

	// Handle specific days
	days := strings.Split(schedule, " and ")
	validDays := map[string]bool{
		"sunday": true, "monday": true, "tuesday": true,
		"wednesday": true, "thursday": true,
		"friday": true, "saturday": true,
	}

	var validatedDays []string
	for _, day := range days {
		day = strings.ToLower(strings.TrimSpace(day))
		if !validDays[day] {
			return "", fmt.Errorf("invalid day: %s", day)
		}
		validatedDays = append(validatedDays, day)
	}

	if len(validatedDays) > 0 {
		return fmt.Sprintf("weekly:%s", strings.Join(validatedDays, ",")), nil
	}

	return "", fmt.Errorf("invalid schedule format")
}
