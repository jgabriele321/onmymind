package time

import (
	"testing"
	"time"
)

func TestGetCurrentTimeWithTools(t *testing.T) {
	tests := []struct {
		name     string
		location string
		wantErr  bool
	}{
		{
			name:     "Valid IANA zone",
			location: "America/New_York",
			wantErr:  false,
		},
		{
			name:     "Valid common city",
			location: "New York",
			wantErr:  false,
		},
		{
			name:     "Invalid location",
			location: "Invalid City",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCurrentTimeWithTools(tt.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentTimeWithTools() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Error("GetCurrentTimeWithTools() returned empty string for valid location")
			}
		})
	}
}

func TestConvertTimeZonesWithTools(t *testing.T) {
	tests := []struct {
		name      string
		timeStr   string
		fromZone  string
		toZone    string
		wantErr   bool
		checkNext bool // whether to check for "next day"
		checkPrev bool // whether to check for "previous day"
	}{
		{
			name:     "Valid conversion 12h",
			timeStr:  "2:00 PM",
			fromZone: "America/New_York",
			toZone:   "Europe/London",
			wantErr:  false,
		},
		{
			name:     "Valid conversion 24h",
			timeStr:  "14:00",
			fromZone: "America/New_York",
			toZone:   "Europe/London",
			wantErr:  false,
		},
		{
			name:      "Next day check",
			timeStr:   "8:00 PM",
			fromZone:  "America/New_York",
			toZone:    "Asia/Tokyo",
			wantErr:   false,
			checkNext: true,
		},
		{
			name:      "Previous day check",
			timeStr:   "8:00 AM",
			fromZone:  "Asia/Tokyo",
			toZone:    "America/New_York",
			wantErr:   false,
			checkPrev: true,
		},
		{
			name:     "Invalid time format",
			timeStr:  "invalid",
			fromZone: "America/New_York",
			toZone:   "Europe/London",
			wantErr:  true,
		},
		{
			name:     "Invalid source zone",
			timeStr:  "2:00 PM",
			fromZone: "Invalid/Zone",
			toZone:   "Europe/London",
			wantErr:  true,
		},
		{
			name:     "Invalid target zone",
			timeStr:  "2:00 PM",
			fromZone: "America/New_York",
			toZone:   "Invalid/Zone",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertTimeZonesWithTools(tt.timeStr, tt.fromZone, tt.toZone)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertTimeZonesWithTools() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == "" {
					t.Error("ConvertTimeZonesWithTools() returned empty string for valid conversion")
				}
				if tt.checkNext && !contains(got, "next day") {
					t.Error("ConvertTimeZonesWithTools() should indicate next day")
				}
				if tt.checkPrev && !contains(got, "previous day") {
					t.Error("ConvertTimeZonesWithTools() should indicate previous day")
				}
			}
		})
	}
}

func TestGetDetailedTimeZoneInfoWithTools(t *testing.T) {
	tests := []struct {
		name     string
		location string
		wantErr  bool
	}{
		{
			name:     "Valid IANA zone",
			location: "America/New_York",
			wantErr:  false,
		},
		{
			name:     "Valid common city",
			location: "New York",
			wantErr:  false,
		},
		{
			name:     "Invalid location",
			location: "Invalid City",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDetailedTimeZoneInfoWithTools(tt.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDetailedTimeZoneInfoWithTools() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Error("GetDetailedTimeZoneInfoWithTools() returned empty string for valid location")
			}
		})
	}
}

func TestValidateLocationNameWithTools(t *testing.T) {
	tests := []struct {
		name          string
		location      string
		wantValid     bool
		wantSuggCount int
	}{
		{
			name:          "Valid IANA zone",
			location:      "America/New_York",
			wantValid:     true,
			wantSuggCount: 0,
		},
		{
			name:          "Valid common city",
			location:      "New York",
			wantValid:     true,
			wantSuggCount: 0,
		},
		{
			name:          "Invalid with suggestions",
			location:      "york",
			wantValid:     false,
			wantSuggCount: 1, // Should suggest "New York"
		},
		{
			name:          "Completely invalid",
			location:      "xyzabc",
			wantValid:     false,
			wantSuggCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, suggestions := ValidateLocationNameWithTools(tt.location)
			if valid != tt.wantValid {
				t.Errorf("ValidateLocationNameWithTools() valid = %v, want %v", valid, tt.wantValid)
			}
			if len(suggestions) != tt.wantSuggCount {
				t.Errorf("ValidateLocationNameWithTools() got %d suggestions, want %d", len(suggestions), tt.wantSuggCount)
			}
		})
	}
}

func TestIsDSTForLocation(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		time time.Time
		want bool
	}{
		{
			name: "Summer time",
			time: time.Date(2024, 7, 1, 12, 0, 0, 0, loc),
			want: true,
		},
		{
			name: "Winter time",
			time: time.Date(2024, 1, 1, 12, 0, 0, 0, loc),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDSTForLocation(tt.time, loc); got != tt.want {
				t.Errorf("isDSTForLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNextDSTTransition(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		time     time.Time
		wantNull bool
	}{
		{
			name:     "Before spring transition",
			time:     time.Date(2024, 3, 1, 12, 0, 0, 0, loc),
			wantNull: false,
		},
		{
			name:     "Before fall transition",
			time:     time.Date(2024, 11, 1, 12, 0, 0, 0, loc),
			wantNull: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getNextDSTTransition(tt.time, loc)
			if (got == nil) != tt.wantNull {
				t.Errorf("getNextDSTTransition() returned nil = %v, want %v", got == nil, tt.wantNull)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return s != "" && s != substr && len(s) >= len(substr) && s[len(s)-len(substr):] == substr
}
