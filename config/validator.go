package config

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Validate validates the entire configuration
func (c *Config) Validate() error {
	if err := c.ValidateScheduler(); err != nil {
		return err
	}
	if err := c.ValidatePlatform(); err != nil {
		return err
	}
	return nil
}

// ValidateScheduler validates scheduler configuration
func (c *Config) ValidateScheduler() error {
	if c.Scheduler.IntervalHours <= 0 {
		return fmt.Errorf("scheduler.interval_hours must be positive")
	}
	if c.Scheduler.IntervalHours > 168 { // 1 week
		return fmt.Errorf("scheduler.interval_hours too large (max 168)")
	}
	if len(c.Scheduler.TargetTimes) == 0 {
		return fmt.Errorf("scheduler.target_times is required (at least one time)")
	}
	for _, t := range c.Scheduler.TargetTimes {
		if !isValidTimeFormat(t) {
			return fmt.Errorf("scheduler.target_times must be in HH:MM format, got: %s", t)
		}
	}
	if c.Scheduler.SafetyBufferSeconds < 0 {
		return fmt.Errorf("scheduler.safety_buffer_seconds cannot be negative")
	}
	if c.Scheduler.SafetyBufferSeconds > 3600 { // 1 hour
		return fmt.Errorf("scheduler.safety_buffer_seconds too large (max 3600)")
	}
	// Validate no conflicts between target times
	if err := validateTargetTimesConflict(c.Scheduler.TargetTimes, c.Scheduler.IntervalHours); err != nil {
		return err
	}
	return nil
}

// validateTargetTimesConflict checks if trigger times conflict with each other
// Conflict definition: trigger_B falls within [trigger_A, trigger_A + interval_hours)
func validateTargetTimesConflict(targetTimes []string, intervalHours int) error {
	// Parse and sort all target times
	parsedTimes := make([]timeOfDay, len(targetTimes))
	for i, t := range targetTimes {
		hour, min := parseTime(t)
		parsedTimes[i] = timeOfDay{hour, min}
	}
	sort.Slice(parsedTimes, func(i, j int) bool {
		if parsedTimes[i].hour != parsedTimes[j].hour {
			return parsedTimes[i].hour < parsedTimes[j].hour
		}
		return parsedTimes[i].min < parsedTimes[j].min
	})

	// Calculate trigger times for each target
	// trigger_time = target_time - interval_hours + safety_buffer (simplified to just -interval_hours for validation)
	interval := time.Duration(intervalHours) * time.Hour

	for i := 0; i < len(parsedTimes); i++ {
		// Calculate when this trigger occurs (in terms of offset from midnight)
		// Since we're comparing within a day, we use the time directly
		targetTrigger := parsedTimes[i].toTime().Add(-interval)

		// Check against all other triggers
		for j := 0; j < len(parsedTimes); j++ {
			if i == j {
				continue
			}

			otherTrigger := parsedTimes[j].toTime().Add(-interval)
			quotaEnd := otherTrigger.Add(interval)

			// Check if targetTrigger falls within [otherTrigger, quotaEnd)
			// We need to handle day wraparound
			if isTimeInRange(targetTrigger, otherTrigger, quotaEnd) {
				return fmt.Errorf("conflicting target times: %s and %s\n"+
					"  Trigger for %s would be at %s\n"+
					"  Trigger for %s would be at %s\n"+
					"  The second trigger falls within the first quota's validity period [%s, %s)\n"+
					"  Please ensure target times are at least %d hours apart",
					targetTimes[i], targetTimes[j],
					targetTimes[i], formatTime(targetTrigger),
					targetTimes[j], formatTime(otherTrigger),
					formatTime(otherTrigger), formatTime(quotaEnd),
					intervalHours)
			}
		}
	}

	return nil
}

// isTimeInRange checks if t is in [start, end)
// Handles day wraparound if end is before start
func isTimeInRange(t, start, end time.Time) bool {
	// Extract time components for comparison
	tSec := t.Hour()*3600 + t.Minute()*60 + t.Second()
	startSec := start.Hour()*3600 + start.Minute()*60 + start.Second()
	endSec := end.Hour()*3600 + end.Minute()*60 + end.Second()

	if startSec < endSec {
		// Normal case: [09:00, 14:00)
		return tSec >= startSec && tSec < endSec
	} else {
		// Wraparound case: [22:00, 02:00+1)
		return tSec >= startSec || tSec < endSec
	}
}

// ValidatePlatform validates platform configuration
func (c *Config) ValidatePlatform() error {
	if c.Platform.Type == "" {
		return fmt.Errorf("platform.type is required")
	}
	supportedTypes := []string{"anthropic"} // Add more as implemented
	supported := false
	for _, t := range supportedTypes {
		if c.Platform.Type == t {
			supported = true
			break
		}
	}
	if !supported {
		return fmt.Errorf("platform.type '%s' is not supported (supported: %s)",
			c.Platform.Type, strings.Join(supportedTypes, ", "))
	}
	if c.Platform.BaseURL == "" {
		return fmt.Errorf("platform.base_url is required")
	}
	if c.Platform.Options == nil {
		return fmt.Errorf("platform.options is required")
	}
	return nil
}

// isValidTimeFormat checks if the time string is in HH:MM format
func isValidTimeFormat(t string) bool {
	parts := strings.Split(t, ":")
	if len(parts) != 2 {
		return false
	}
	var hour, min int
	_, err := fmt.Sscanf(t, "%d:%d", &hour, &min)
	if err != nil {
		return false
	}
	return hour >= 0 && hour <= 23 && min >= 0 && min <= 59
}

// parseTime parses HH:MM format
func parseTime(t string) (hour, min int) {
	fmt.Sscanf(t, "%d:%d", &hour, &min)
	return
}

// timeOfDay represents a time within a day
type timeOfDay struct {
	hour int
	min  int
}

// toTime converts to a time.Time (date portion is arbitrary)
func (tod timeOfDay) toTime() time.Time {
	return time.Date(2000, 1, 1, tod.hour, tod.min, 0, 0, time.UTC)
}

// formatTime formats a time.Time as HH:MM
func formatTime(t time.Time) string {
	return fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
}
