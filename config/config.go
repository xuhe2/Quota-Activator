package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Scheduler SchedulerConfig  `yaml:"scheduler"`
	Platform  PlatformConfig   `yaml:"platform"`
}

// SchedulerConfig holds scheduler-specific configuration
type SchedulerConfig struct {
	IntervalHours       int      `yaml:"interval_hours"`
	TargetTimes         []string `yaml:"target_times"`      // Array of target times (e.g., ["09:00", "14:00", "19:00"])
	SafetyBufferSeconds int      `yaml:"safety_buffer_seconds"`
}

// PlatformConfig holds platform-specific configuration
type PlatformConfig struct {
	Type    string         `yaml:"type"`     // "anthropic", "openai", "glm", etc.
	BaseURL string         `yaml:"base_url"`
	Options map[string]any `yaml:"options"` // platform-specific options
}

// Load reads and parses the YAML configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	if cfg.Scheduler.SafetyBufferSeconds == 0 {
		cfg.Scheduler.SafetyBufferSeconds = 60
	}
	if cfg.Platform.Options == nil {
		cfg.Platform.Options = make(map[string]any)
	}

	return &cfg, nil
}
