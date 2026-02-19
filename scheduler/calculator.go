package scheduler

import (
	"fmt"
	"sort"
	"time"
)

// DailyTrigger represents a single trigger time for a day
type DailyTrigger struct {
	TargetTime  string    // Original target time (e.g., "14:00")
	TriggerTime time.Time // When to trigger (e.g., 09:01)
}

// CalculateDailyTriggers calculates all trigger times for a given day
//
// For each target time, the trigger time is: target_time - interval_hours + buffer_seconds
// This ensures fresh quota is available at the target time.
func CalculateDailyTriggers(date time.Time, targetTimes []string, intervalHours, bufferSeconds int) []DailyTrigger {
	interval := time.Duration(intervalHours) * time.Hour
	buffer := time.Duration(bufferSeconds) * time.Second

	var triggers []DailyTrigger

	for _, targetStr := range targetTimes {
		// Parse target time (e.g., "14:00")
		var hour, min int
		_, err := fmt.Sscanf(targetStr, "%d:%d", &hour, &min)
		if err != nil {
			continue // Skip invalid times (should have been validated already)
		}

		// Create target time for the given date
		targetTime := time.Date(date.Year(), date.Month(), date.Day(), hour, min, 0, 0, date.Location())

		// Calculate trigger time: target - interval + buffer
		triggerTime := targetTime.Add(-interval).Add(buffer)

		triggers = append(triggers, DailyTrigger{
			TargetTime:  targetStr,
			TriggerTime: triggerTime,
		})
	}

	// Sort by trigger time
	sort.Slice(triggers, func(i, j int) bool {
		return triggers[i].TriggerTime.Before(triggers[j].TriggerTime)
	})

	return triggers
}

// NextTrigger finds the next trigger time from now
// Returns triggers for today and future days as needed
func NextTrigger(now time.Time, targetTimes []string, intervalHours, bufferSeconds int) (trigger DailyTrigger, nextDate time.Time) {
	// Try today first
	today := now.Truncate(24 * time.Hour)
	triggers := CalculateDailyTriggers(today, targetTimes, intervalHours, bufferSeconds)

	for _, t := range triggers {
		if t.TriggerTime.After(now) {
			return t, today
		}
	}

	// If no trigger left today, try tomorrow
	tomorrow := today.Add(24 * time.Hour)
	triggers = CalculateDailyTriggers(tomorrow, targetTimes, intervalHours, bufferSeconds)
	if len(triggers) > 0 {
		return triggers[0], tomorrow
	}

	// Fallback (shouldn't happen)
	return DailyTrigger{
		TriggerTime: now.Add(24 * time.Hour),
	}, tomorrow
}
