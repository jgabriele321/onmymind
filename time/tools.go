package time

import (
	"fmt"
	"strings"
	"time"
)

// GetCurrentTimeWithTools returns the current time in the specified location
func GetCurrentTimeWithTools(location string) (string, error) {
	loc, err := time.LoadLocation(location)
	if err != nil {
		// Try to map common city names to IANA zones
		if mappedZone, ok := commonCityToZone[strings.ToLower(location)]; ok {
			loc, err = time.LoadLocation(mappedZone)
			if err != nil {
				return "", fmt.Errorf("invalid location after mapping: %v", err)
			}
		} else {
			return "", fmt.Errorf("invalid location: %v", err)
		}
	}

	now := time.Now().In(loc)
	zoneName, offset := now.Zone()
	isDST := isDSTForLocation(now, loc)

	// Format the response with both 12h and 24h time formats
	return fmt.Sprintf("The current time in %s is %s %s (UTC%+d), DST is %s",
		strings.Title(location),
		now.Format("3:04 PM (15:04)"),
		zoneName,
		offset/3600,
		map[bool]string{true: "in effect", false: "not in effect"}[isDST]), nil
}

// ConvertTimeZonesWithTools converts a time from one zone to another
func ConvertTimeZonesWithTools(timeStr, fromZone, toZone string) (string, error) {
	// Clean up input time string and zones
	timeStr = strings.TrimSpace(timeStr)
	timeStr = strings.TrimSuffix(timeStr, " UTC") // Remove UTC suffix if present
	fromZone = strings.ToLower(fromZone)
	toZone = strings.ToLower(toZone)

	// Special handling for UTC
	if fromZone == "utc" {
		fromZone = "UTC"
	}
	if toZone == "utc" {
		toZone = "UTC"
	}

	// Parse the input time
	parsedTime, err := time.Parse("3:04 PM", timeStr)
	if err != nil {
		parsedTime, err = time.Parse("15:04", timeStr)
		if err != nil {
			return "", fmt.Errorf("invalid time format: please use either 12-hour (e.g., 2:00 PM) or 24-hour (e.g., 14:00) format")
		}
	}

	// Load source location
	fromLoc, err := time.LoadLocation(fromZone)
	if err != nil {
		if mappedZone, ok := commonCityToZone[fromZone]; ok {
			fromLoc, err = time.LoadLocation(mappedZone)
			if err != nil {
				return "", fmt.Errorf("invalid source location after mapping: %v", err)
			}
		} else {
			return "", fmt.Errorf("invalid source location: %v (try using a city name like 'New York' or IANA zone like 'America/New_York', or 'UTC')", err)
		}
	}

	// Load target location
	toLoc, err := time.LoadLocation(toZone)
	if err != nil {
		if mappedZone, ok := commonCityToZone[toZone]; ok {
			toLoc, err = time.LoadLocation(mappedZone)
			if err != nil {
				return "", fmt.Errorf("invalid target location after mapping: %v", err)
			}
		} else {
			return "", fmt.Errorf("invalid target location: %v (try using a city name like 'New York' or IANA zone like 'America/New_York', or 'UTC')", err)
		}
	}

	// Set the time in the source location
	now := time.Now()
	sourceTime := time.Date(
		now.Year(), now.Month(), now.Day(),
		parsedTime.Hour(), parsedTime.Minute(), 0, 0,
		fromLoc,
	)

	// Convert to target location
	targetTime := sourceTime.In(toLoc)

	// Get zone information
	_, fromOffset := sourceTime.Zone()
	_, toOffset := targetTime.Zone()

	// Check if either location is in DST
	fromDST := isDSTForLocation(sourceTime, fromLoc)
	toDST := isDSTForLocation(targetTime, toLoc)

	// Format the response
	dayDiff := ""
	if targetTime.Day() != sourceTime.Day() || targetTime.Month() != sourceTime.Month() {
		if targetTime.Day() < sourceTime.Day() || targetTime.Month() < sourceTime.Month() {
			dayDiff = " previous day"
		} else {
			dayDiff = " next day"
		}
	}

	return fmt.Sprintf("%s %s (UTC%+d, DST %s) →\n%s %s (UTC%+d, DST %s)%s",
		sourceTime.Format("3:04 PM (15:04)"),
		fromZone,
		fromOffset/3600,
		map[bool]string{true: "in effect", false: "not in effect"}[fromDST],
		targetTime.Format("3:04 PM (15:04)"),
		toZone,
		toOffset/3600,
		map[bool]string{true: "in effect", false: "not in effect"}[toDST],
		dayDiff), nil
}

// GetDetailedTimeZoneInfoWithTools returns detailed information about a time zone
func GetDetailedTimeZoneInfoWithTools(location string) (string, error) {
	loc, err := time.LoadLocation(location)
	if err != nil {
		if mappedZone, ok := commonCityToZone[strings.ToLower(location)]; ok {
			loc, err = time.LoadLocation(mappedZone)
			if err != nil {
				return "", fmt.Errorf("invalid location after mapping: %v", err)
			}
		} else {
			return "", fmt.Errorf("invalid location: %v", err)
		}
	}

	now := time.Now().In(loc)
	zoneName, offset := now.Zone()
	isDST := isDSTForLocation(now, loc)

	// Get next DST transition if any
	nextTransition := getNextDSTTransition(now, loc)
	transitionInfo := ""
	if nextTransition != nil {
		transitionInfo = fmt.Sprintf("\nNext DST transition: %s", nextTransition.Format("2006-01-02 15:04 MST"))
	}

	return fmt.Sprintf("%s, UTC%+d, DST %s%s",
		zoneName,
		offset/3600,
		map[bool]string{true: "in effect", false: "not in effect"}[isDST],
		transitionInfo), nil
}

// ValidateLocationNameWithTools checks if a location name is valid and returns suggestions if not
func ValidateLocationNameWithTools(location string) (bool, []string) {
	// First check if it's a direct IANA zone
	_, err := time.LoadLocation(location)
	if err == nil {
		return true, nil
	}

	// Check if it's in our common city mappings
	if zone, ok := commonCityToZone[strings.ToLower(location)]; ok {
		_, err := time.LoadLocation(zone)
		if err == nil {
			return true, nil
		}
	}

	// If not found, generate suggestions
	suggestions := []string{}
	searchTerm := strings.ToLower(location)

	// Search through common city mappings
	for city, zone := range commonCityToZone {
		if strings.Contains(city, searchTerm) || strings.Contains(zone, searchTerm) {
			suggestions = append(suggestions, fmt.Sprintf("%s (%s)", strings.Title(city), zone))
		}
	}

	return false, suggestions
}

// Helper function to check if a time is in DST
func isDSTForLocation(t time.Time, loc *time.Location) bool {
	// Get the offset at the given time
	_, offset := t.In(loc).Zone()

	// Get the offset for the same time in January (usually non-DST)
	janTime := time.Date(t.Year(), 1, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
	_, janOffset := janTime.Zone()

	// If the offset is different, we're in DST
	return offset != janOffset
}

// Helper function to get the next DST transition
func getNextDSTTransition(t time.Time, loc *time.Location) *time.Time {
	// Look ahead up to a year
	for i := 0; i < 365; i++ {
		nextDay := t.AddDate(0, 0, i)
		_, offset1 := nextDay.In(loc).Zone()
		nextDayPlus := nextDay.AddDate(0, 0, 1)
		_, offset2 := nextDayPlus.In(loc).Zone()

		if offset1 != offset2 {
			return &nextDay
		}
	}
	return nil
}

// Common city names mapped to IANA time zones
var commonCityToZone = map[string]string{
	"new york":     "America/New_York",
	"nyc":          "America/New_York",
	"london":       "Europe/London",
	"paris":        "Europe/Paris",
	"tokyo":        "Asia/Tokyo",
	"sydney":       "Australia/Sydney",
	"melbourne":    "Australia/Melbourne",
	"singapore":    "Asia/Singapore",
	"hong kong":    "Asia/Hong_Kong",
	"berlin":       "Europe/Berlin",
	"rome":         "Europe/Rome",
	"madrid":       "Europe/Madrid",
	"dubai":        "Asia/Dubai",
	"moscow":       "Europe/Moscow",
	"beijing":      "Asia/Shanghai",
	"shanghai":     "Asia/Shanghai",
	"los angeles":  "America/Los_Angeles",
	"la":           "America/Los_Angeles",
	"chicago":      "America/Chicago",
	"toronto":      "America/Toronto",
	"vancouver":    "America/Vancouver",
	"sao paulo":    "America/Sao_Paulo",
	"mexico city":  "America/Mexico_City",
	"mumbai":       "Asia/Kolkata",
	"delhi":        "Asia/Kolkata",
	"bangkok":      "Asia/Bangkok",
	"cairo":        "Africa/Cairo",
	"johannesburg": "Africa/Johannesburg",
	"auckland":     "Pacific/Auckland",
	"utc":          "UTC", // Add UTC as a valid zone
}
