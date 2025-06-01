package time

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Common city names to IANA time zone mappings
var timeZoneMap = map[string]string{
	"london":    "Europe/London",
	"austin":    "America/Chicago", // Austin uses Central Time
	"new york":  "America/New_York",
	"tokyo":     "Asia/Tokyo",
	"paris":     "Europe/Paris",
	"sydney":    "Australia/Sydney",
	"singapore": "Asia/Singapore",
	"dubai":     "Asia/Dubai",
	"moscow":    "Europe/Moscow",
	"berlin":    "Europe/Berlin",
	"nyc":       "America/New_York",
	"la":        "America/Los_Angeles",
	"sf":        "America/Los_Angeles",
}

// TimeZoneInfo holds information about a location's time zone
type TimeZoneInfo struct {
	Location       string
	CurrentTime    time.Time
	ZoneName       string
	Offset         int        // offset in hours from UTC
	IsDST          bool       // whether DST is in effect
	NextTransition *time.Time // next DST transition, if any
}

// GetDetailedTimeZoneInfo returns detailed time zone information for a location
func GetDetailedTimeZoneInfo(location string) (*TimeZoneInfo, error) {
	// Clean up input
	location = strings.ToLower(strings.TrimSpace(location))

	// Try to find the time zone name
	zoneName, ok := timeZoneMap[location]
	if !ok {
		// If not found in our map, try using the input directly
		zoneName = location
	}

	// Load the location from the time zone database
	loc, err := time.LoadLocation(zoneName)
	if err != nil {
		return nil, fmt.Errorf("unknown location %q: %v", location, err)
	}

	// Get current time in that location
	now := time.Now().In(loc)

	// Get zone name and offset
	name, offset := now.Zone()

	// Get next transition (if any)
	var nextTransition *time.Time
	if trans, ok := getNextTransition(now, loc); ok {
		nextTransition = &trans
	}

	return &TimeZoneInfo{
		Location:       location,
		CurrentTime:    now,
		ZoneName:       name,
		Offset:         offset / 3600, // Convert seconds to hours
		IsDST:          now.IsDST(),
		NextTransition: nextTransition,
	}, nil
}

// ConvertTimeZones converts a time between two locations
func ConvertTimeZones(timeStr, fromLocation, toLocation string) (string, error) {
	// Get time zone info for both locations
	fromInfo, err := GetDetailedTimeZoneInfo(fromLocation)
	if err != nil {
		return "", fmt.Errorf("invalid source location: %v", err)
	}

	toInfo, err := GetDetailedTimeZoneInfo(toLocation)
	if err != nil {
		return "", fmt.Errorf("invalid destination location: %v", err)
	}

	// Parse the input time
	t, err := time.Parse("3:04 PM", timeStr)
	if err != nil {
		return "", fmt.Errorf("invalid time format: %v", err)
	}

	// Create a time in the source location
	sourceTime := time.Date(
		time.Now().Year(),
		time.Now().Month(),
		time.Now().Day(),
		t.Hour(),
		t.Minute(),
		0, 0,
		time.Local,
	).In(fromInfo.CurrentTime.Location())

	// Convert to destination time zone
	destTime := sourceTime.In(toInfo.CurrentTime.Location())

	// Check if it's the next/previous day
	dayDiff := destTime.Day() - sourceTime.Day()
	nextDay := ""
	if dayDiff == 1 {
		nextDay = " (next day)"
	} else if dayDiff == -1 {
		nextDay = " (previous day)"
	}

	return fmt.Sprintf("%s %s (UTC%+d) → %s %s (UTC%+d)%s",
		sourceTime.Format("3:04 PM (15:04)"),
		fromInfo.ZoneName,
		fromInfo.Offset,
		destTime.Format("3:04 PM (15:04)"),
		toInfo.ZoneName,
		toInfo.Offset,
		nextDay,
	), nil
}

// ValidateLocationName checks if a location is valid and returns suggestions if not
func ValidateLocationName(location string) (bool, []string) {
	location = strings.ToLower(strings.TrimSpace(location))

	// Check common locations first
	if _, ok := timeZoneMap[location]; ok {
		return true, nil
	}

	// Try loading the location directly
	_, err := time.LoadLocation(location)
	if err == nil {
		return true, nil
	}

	// If not found, look for similar locations
	var suggestions []string
	for loc := range timeZoneMap {
		if strings.Contains(loc, location) || strings.Contains(location, loc) {
			suggestions = append(suggestions, loc)
		}
	}

	return false, suggestions
}

// Helper function to get the next DST transition
func getNextTransition(t time.Time, loc *time.Location) (time.Time, bool) {
	// Look ahead up to a year for the next transition
	for i := 0; i < 365; i++ {
		t = t.AddDate(0, 0, 1)
		_, offset1 := t.Zone()
		tomorrow := t.AddDate(0, 0, 1)
		_, offset2 := tomorrow.Zone()
		if offset1 != offset2 {
			return t, true
		}
	}
	return t, false
}

type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// TimeCalculator handles time-related calculations and queries
type TimeCalculator struct {
	openRouterKey string
	client        *http.Client
}

// NewTimeCalculator creates a new TimeCalculator instance
func NewTimeCalculator(openRouterKey string) *TimeCalculator {
	return &TimeCalculator{
		openRouterKey: openRouterKey,
		client:        &http.Client{},
	}
}

// ProcessQuery handles time-related queries using OpenRouter
func (tc *TimeCalculator) ProcessQuery(query string) (string, error) {
	if tc.openRouterKey == "" {
		log.Printf("Error: OpenRouter API key is not set")
		return "", fmt.Errorf("OpenRouter API key is not configured")
	}

	// We'll use Claude-2 for its strong reasoning capabilities
	model := "anthropic/claude-2"

	systemPrompt := `You are a time calculation assistant. To perform time calculations, you MUST use the exact tool call format:

Tool: GetCurrentTime("location")
Tool: ConvertTimeZones("time", "fromZone", "toZone")
Tool: GetDetailedTimeZoneInfo("location")
Tool: ValidateLocationName("location")

The tools available are:

1. GetCurrentTime(location)
   Input: City or location name in quotes
   Returns: Current time, zone name, and DST status
   Example: Tool: GetCurrentTime("New York")

2. ConvertTimeZones(time, fromZone, toZone)
   Input: Time expression and location names in quotes
   Returns: Converted time with zone details
   Example: Tool: ConvertTimeZones("2:30 PM", "New York", "Tokyo")

3. GetDetailedTimeZoneInfo(location)
   Input: City or location name in quotes
   Returns: Zone name, offset, and DST information
   Example: Tool: GetDetailedTimeZoneInfo("London")

4. ValidateLocationName(location)
   Input: City or location name in quotes
   Returns: Whether location is valid and suggestions if not
   Example: Tool: ValidateLocationName("NYC")

IMPORTANT RULES:
1. ALWAYS use the EXACT tool call format shown above
2. NEVER perform manual time calculations
3. NEVER assume time zones or offsets
4. NEVER use hardcoded example times - always use the tools
5. Validate locations before using them
6. Show both 12h and 24h time formats
7. Include DST information when relevant
8. For queries about current time, ALWAYS use GetCurrentTime
9. For time conversions, ALWAYS use ConvertTimeZones

Example Usage:

Q: "What time is it in Tokyo?"
A: Let me check the current time in Tokyo.
First, I'll validate the location:
Tool: ValidateLocationName("Tokyo")
Now I'll get the current time:
Tool: GetCurrentTime("Tokyo")

Q: "If it's 2pm in New York, what time is it in London?"
A: I'll help you with that conversion.
1. Validate both locations:
   Tool: ValidateLocationName("New York")
   Tool: ValidateLocationName("London")
2. Convert the time:
   Tool: ConvertTimeZones("2:00 PM", "New York", "London")

Q: "What's the time difference between Paris and Sydney?"
A: Let me check both time zones.
1. Get information for both cities:
   Tool: GetDetailedTimeZoneInfo("Paris")
   Tool: GetDetailedTimeZoneInfo("Sydney")

For any time-related query:
1. Always validate locations first using Tool: ValidateLocationName("location")
2. For current time, use Tool: GetCurrentTime("location")
3. For conversions, use Tool: ConvertTimeZones("time", "from", "to")
4. For zone info, use Tool: GetDetailedTimeZoneInfo("location")
5. Format responses clearly with both 12h and 24h times
6. Include relevant DST information
7. Show step-by-step calculations when needed`

	// Extract tool calls from the response and execute them
	toolPattern := regexp.MustCompile(`Tool: (\w+)\("([^"]+)"(?:, "([^"]+)")?(?:, "([^"]+)")?\)`)

	// Process any tool calls in the query first
	toolCalls := toolPattern.FindAllStringSubmatch(query, -1)
	for _, call := range toolCalls {
		toolName := call[1]
		args := call[2:]

		// Remove empty args
		var validArgs []string
		for _, arg := range args {
			if arg != "" {
				validArgs = append(validArgs, arg)
			}
		}

		// Execute the tool and replace the call with its result
		result, err := tc.executeTool(toolName, validArgs...)
		if err != nil {
			log.Printf("Error executing tool %s: %v", toolName, err)
			continue
		}

		// Replace the tool call with its result
		query = strings.Replace(query, call[0], result, 1)
	}

	// Add current time to user's query
	queryWithTime := fmt.Sprintf("Current time: %s UTC\n\nQuery: %s",
		time.Now().Format("15:04"),
		query)

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: queryWithTime},
	}

	reqBody := OpenRouterRequest{
		Model:    model,
		Messages: messages,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tc.openRouterKey)
	req.Header.Set("HTTP-Referer", "https://github.com/jgabriele321/onmymind")
	req.Header.Set("X-Title", "OnMuyMind Bot")

	resp, err := tc.client.Do(req)
	if err != nil {
		log.Printf("Error making request to OpenRouter: %v", err)
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return "", fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("OpenRouter API error: Status %d, Body: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("OpenRouter API error: %s", resp.Status)
	}

	var openRouterResp OpenRouterResponse
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		log.Printf("Error decoding response: %v, Body: %s", err, string(body))
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	if len(openRouterResp.Choices) == 0 {
		log.Printf("No choices in response. Full response: %s", string(body))
		return "", fmt.Errorf("no response from OpenRouter")
	}

	response := openRouterResp.Choices[0].Message.Content

	// Process any tool calls in the response
	toolCalls = toolPattern.FindAllStringSubmatch(response, -1)
	for _, call := range toolCalls {
		toolName := call[1]
		args := call[2:]

		// Remove empty args
		var validArgs []string
		for _, arg := range args {
			if arg != "" {
				validArgs = append(validArgs, arg)
			}
		}

		// Execute the tool and replace the call with its result
		result, err := tc.executeTool(toolName, validArgs...)
		if err != nil {
			log.Printf("Error executing tool %s: %v", toolName, err)
			continue
		}

		// Replace the tool call with its result
		response = strings.Replace(response, call[0], result, 1)
	}

	return strings.TrimSpace(response), nil
}

// executeTool executes a tool function with the given arguments
func (tc *TimeCalculator) executeTool(name string, args ...string) (string, error) {
	switch name {
	case "GetCurrentTime":
		if len(args) != 1 {
			return "", fmt.Errorf("GetCurrentTime requires exactly 1 argument")
		}
		result, err := GetCurrentTimeWithTools(args[0])
		if err != nil {
			return "", err
		}
		return result, nil

	case "ConvertTimeZones":
		if len(args) != 3 {
			return "", fmt.Errorf("ConvertTimeZones requires exactly 3 arguments")
		}
		result, err := ConvertTimeZonesWithTools(args[0], args[1], args[2])
		if err != nil {
			return "", err
		}
		return result, nil

	case "GetDetailedTimeZoneInfo":
		if len(args) != 1 {
			return "", fmt.Errorf("GetDetailedTimeZoneInfo requires exactly 1 argument")
		}
		result, err := GetDetailedTimeZoneInfoWithTools(args[0])
		if err != nil {
			return "", err
		}
		return result, nil

	case "ValidateLocationName":
		if len(args) != 1 {
			return "", fmt.Errorf("ValidateLocationName requires exactly 1 argument")
		}
		valid, suggestions := ValidateLocationNameWithTools(args[0])
		if valid {
			return "true", nil
		}
		if len(suggestions) > 0 {
			return fmt.Sprintf("false, suggestions: %s", strings.Join(suggestions, ", ")), nil
		}
		return "false", nil

	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
