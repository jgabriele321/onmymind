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

	systemPrompt := `You are a time calculation assistant. Your task is to:
1. Parse time-related queries
2. Show step-by-step calculations when relevant
3. Handle time zones, formats (12h/24h), and arithmetic
4. Present results clearly and concisely
5. Use the provided current time for all calculations
6. Always show both 12h and 24h format in responses when relevant
7. Consider daylight saving time (DST) when calculating time zones

The current time will be provided with each query in UTC. You should use this to answer any questions about current time in any timezone.

Important time zone notes:
- Austin, TX uses Central Time (CT): UTC-6 in winter (CST), UTC-5 in summer (CDT)
- London, UK uses British Time (BT): UTC+0 in winter (GMT), UTC+1 in summer (BST)
- Always check if a location is currently observing DST before calculating

Example outputs:
Q: "What time is 14:00?"
A: 14:00 is 2:00 PM

Q: "What time is it in Austin?"
A: Given the current time [10:30 PM UTC]:
Austin is currently in CDT (UTC-5), so it's 5:30 PM (17:30) in Austin, TX

Q: "If my flight is at 9:45 AM and I need 1h drive + 30m security, when to leave?"
A: Let's calculate backwards:
1. Flight time: 9:45 AM
2. Security: -30 minutes
3. Drive: -1 hour
→ You should leave at 8:15 AM (08:15)

Keep responses focused and precise.`

	// Add current time to user's query
	queryWithTime := fmt.Sprintf("Current time: %s UTC\n\nQuery: %s",
		now.Format("15:04 MST"),
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
