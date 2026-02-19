package scheduler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"quota-activator/config"
	"quota-activator/platform"
)

// Scheduler handles the timing and execution of quota refresh triggers
type Scheduler struct {
	schedulerCfg *config.SchedulerConfig
	platform     platform.Platform
}

// New creates a new scheduler instance
func New(schedulerCfg *config.SchedulerConfig, p platform.Platform) *Scheduler {
	return &Scheduler{
		schedulerCfg: schedulerCfg,
		platform:     p,
	}
}

// Start begins the scheduling loop
// It calculates the next trigger time, waits until then, executes the trigger,
// and repeats until the context is cancelled.
func (s *Scheduler) Start(ctx context.Context) error {
	log.Printf("Scheduler started for platform: %s", s.platform.Name())
	log.Printf("Target times: [%s], Interval: %dh, Safety buffer: %ds",
		strings.Join(s.schedulerCfg.TargetTimes, ", "),
		s.schedulerCfg.IntervalHours,
		s.schedulerCfg.SafetyBufferSeconds)

	// Calculate first trigger time
	nextTrigger, _ := NextTrigger(
		time.Now(),
		s.schedulerCfg.TargetTimes,
		s.schedulerCfg.IntervalHours,
		s.schedulerCfg.SafetyBufferSeconds,
	)
	log.Printf("First trigger scheduled at: %s (for target: %s)",
		nextTrigger.TriggerTime.Format("2006-01-02 15:04:05"),
		nextTrigger.TargetTime)

	// Main loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		now := time.Now()
		if nextTrigger.TriggerTime.After(now) {
			duration := nextTrigger.TriggerTime.Sub(now)
			log.Printf("Waiting %s until next trigger...", duration.Round(time.Second))

			// Sleep until trigger time or context cancellation
			select {
			case <-time.After(duration):
				// Time to trigger
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Trigger the platform request
		log.Printf("[%s] Triggering quota refresh (for target: %s)...",
			s.platform.Name(), nextTrigger.TargetTime)
		if err := s.platform.Trigger(ctx); err != nil {
			log.Printf("[ERROR] Trigger failed: %v", err)
		} else {
			log.Printf("[SUCCESS] Trigger completed")
		}

		// Calculate next trigger time
		nextTrigger, _ = NextTrigger(
			time.Now(),
			s.schedulerCfg.TargetTimes,
			s.schedulerCfg.IntervalHours,
			s.schedulerCfg.SafetyBufferSeconds,
		)
		log.Printf("Next trigger: %s (for target: %s)",
			nextTrigger.TriggerTime.Format("2006-01-02 15:04:05"),
			nextTrigger.TargetTime)
	}
}

// NextTrigger returns the next scheduled trigger
func (s *Scheduler) NextTrigger() DailyTrigger {
	nextTrigger, _ := NextTrigger(
		time.Now(),
		s.schedulerCfg.TargetTimes,
		s.schedulerCfg.IntervalHours,
		s.schedulerCfg.SafetyBufferSeconds,
	)
	return nextTrigger
}

// String returns a string representation of the scheduler
func (s *Scheduler) String() string {
	return fmt.Sprintf("Scheduler{platform=%s, interval=%dh, targets=[%s]}",
		s.platform.Name(),
		s.schedulerCfg.IntervalHours,
		strings.Join(s.schedulerCfg.TargetTimes, ", "))
}
