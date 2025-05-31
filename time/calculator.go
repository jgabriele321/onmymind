package time

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// TimeZoneInfo holds information about a location's time zone
type TimeZoneInfo struct {
	Location    string
	CurrentTime time.Time
	ZoneName    string
	Offset      int // offset in hours from UTC
}

// GetTimeZoneInfo returns the current time zone information for a location
func GetTimeZoneInfo(location string) (*TimeZoneInfo, error) {
	// Map common city names to IANA time zone names
	timeZoneMap := map[string]string{
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
	}

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

	return &TimeZoneInfo{
		Location:    location,
		CurrentTime: now,
		ZoneName:    name,
		Offset:      offset / 3600, // Convert seconds to hours
	}, nil
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

	// Get current time in UTC
	now := time.Now()

	// Try to get time zone info if this is a time zone query
	var tzInfo *TimeZoneInfo
	if strings.Contains(strings.ToLower(query), "time") && strings.Contains(strings.ToLower(query), "in") {
		// Extract location from query (simple extraction, can be improved)
		parts := strings.Split(strings.ToLower(query), "in")
		if len(parts) > 1 {
			location := strings.TrimSpace(parts[len(parts)-1])
			// Remove common words
			location = strings.TrimSuffix(location, "?")
			location = strings.TrimPrefix(location, "the")
			location = strings.TrimSpace(location)

			info, err := GetTimeZoneInfo(location)
			if err == nil {
				tzInfo = info
			}
		}
	}

	systemPrompt := `You are a time calculation assistant. Your task is to:
1. Parse time-related queries
2. Show step-by-step calculations when relevant
3. Handle time zones, formats (12h/24h), and arithmetic
4. Present results clearly and concisely
5. Use ONLY the provided time zone information for calculations
6. Always show both 12h and 24h format in responses when relevant

IMPORTANT: When time zone information is provided, DO NOT perform manual calculations. 
Instead, use the exact local time that was queried from the IANA Time Zone database.

The current time will be provided with each query in UTC.`

	// Add time zone info to the prompt if available
	if tzInfo != nil {
		systemPrompt += fmt.Sprintf(`

LOCATION TIME ZONE DATA (Use this exact information, do not calculate manually):
• Location: %s
• Current local time: %s
• Time zone: %s (UTC%+d)
• DST status: %s

For time zone queries, use this information directly instead of doing manual calculations.
`,
			tzInfo.Location,
			tzInfo.CurrentTime.Format("3:04 PM (15:04)"),
			tzInfo.ZoneName,
			tzInfo.Offset,
			map[bool]string{true: "in effect", false: "not in effect"}[tzInfo.CurrentTime.IsDST()])
	}

	systemPrompt += `

Example outputs:
Q: "What time is 14:00?"
A: 14:00 is 2:00 PM

Q: "If my flight is at 9:45 AM and I need 1h drive + 30m security, when to leave?"
A: Let's calculate backwards:
1. Flight time: 9:45 AM
2. Security: -30 minutes
3. Drive: -1 hour
→ You should leave at 8:15 AM (08:15)

Keep responses focused and precise. For time zone queries, always use the provided IANA Time Zone data instead of doing manual calculations.`

	// Add current time to user's query
	queryWithTime := fmt.Sprintf("Current time: %s UTC\n\nQuery: %s",
		now.Format("15:04"),
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
	req.Header.Set("HTTP-Referer", "https://github.com/jgabriele321/onmymind") // Updated to your repo
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

	// Clean up the response
	response := strings.TrimSpace(openRouterResp.Choices[0].Message.Content)
	return response, nil
}
