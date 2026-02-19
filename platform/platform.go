package platform

import (
	"context"
)

// Platform defines the interface for all quota activation platforms
type Platform interface {
	// Name returns the platform name
	Name() string

	// Trigger sends a minimal request to activate the quota
	Trigger(ctx context.Context) error

	// ValidateConfig validates platform-specific configuration
	ValidateConfig() error
}

// PlatformInput holds the input data needed to create a platform
type PlatformInput struct {
	Type    string
	BaseURL string
	Options map[string]any
}

// NewPlatform creates a platform instance based on the configured type
func NewPlatform(input *PlatformInput) (Platform, error) {
	switch input.Type {
	case "anthropic":
		return NewAnthropicPlatform(input)
	// Future platforms can be added here
	// case "openai":
	//     return NewOpenAIPlatform(input)
	// case "glm":
	//     return NewGLMPlatform(input)
	default:
		return nil, &UnsupportedPlatformError{Type: input.Type}
	}
}

// UnsupportedPlatformError is returned when an unknown platform type is requested
type UnsupportedPlatformError struct {
	Type string
}

func (e *UnsupportedPlatformError) Error() string {
	return "unsupported platform type: " + e.Type
}
