package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// AnthropicPlatform implements the Platform interface for Anthropic Claude API
type AnthropicPlatform struct {
	baseURL    string
	apiKey     string
	model      string
	timeout    time.Duration
	maxRetries int
}

// AnthropicRequest represents the request payload for Anthropic API
type AnthropicRequest struct {
	Model     string         `json:"model"`
	MaxTokens int            `json:"max_tokens"`
	Stream    bool           `json:"stream"`
	Messages  []AnthropicMsg `json:"messages"`
}

// AnthropicMsg represents a message in Anthropic format
type AnthropicMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NewAnthropicPlatform creates a new Anthropic platform instance
func NewAnthropicPlatform(input *PlatformInput) (*AnthropicPlatform, error) {
	p := &AnthropicPlatform{
		baseURL:    input.BaseURL,
		timeout:    30 * time.Second, // default
		maxRetries: 0,                // default
	}

	// Extract options
	if apiKey, ok := input.Options["api_key"].(string); ok {
		p.apiKey = apiKey
	} else {
		return nil, fmt.Errorf("api_key is required for anthropic platform")
	}

	if model, ok := input.Options["model"].(string); ok {
		p.model = model
	} else {
		p.model = "claude-3-5-sonnet-20241022" // default
	}

	if timeoutSec, ok := input.Options["timeout_seconds"].(int); ok {
		p.timeout = time.Duration(timeoutSec) * time.Second
	}

	if retries, ok := input.Options["max_retries"].(int); ok {
		p.maxRetries = retries
	}

	return p, nil
}

// Name returns the platform name
func (p *AnthropicPlatform) Name() string {
	return "anthropic"
}

// ValidateConfig validates the Anthropic-specific configuration
func (p *AnthropicPlatform) ValidateConfig() error {
	if p.baseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	if p.apiKey == "" {
		return fmt.Errorf("api_key is required")
	}
	if p.model == "" {
		return fmt.Errorf("model is required")
	}
	return nil
}

// Trigger sends a minimal request to activate the Anthropic quota
func (p *AnthropicPlatform) Trigger(ctx context.Context) error {
	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("[Anthropic] Retry attempt %d/%d", attempt, p.maxRetries)
			// Exponential backoff
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		err := p.doTrigger(ctx)
		if err == nil {
			return nil
		}
		lastErr = err
		log.Printf("[Anthropic] Attempt %d failed: %v", attempt+1, err)
	}

	return fmt.Errorf("trigger failed after %d attempts: %w", p.maxRetries+1, lastErr)
}

// doTrigger performs a single trigger attempt
func (p *AnthropicPlatform) doTrigger(ctx context.Context) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Build request payload
	reqBody := AnthropicRequest{
		Model:     p.model,
		MaxTokens: 1,
		Stream:    true,
		Messages: []AnthropicMsg{
			{Role: "user", Content: "hi"},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	reqUrl, err := url.JoinPath(p.baseURL, "v1/messages")
	if err != nil {
		return fmt.Errorf("failed to join path: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", reqUrl, strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read first chunk and discard (we just need to trigger the request)
	buf := make([]byte, 1024)
	if _, err := resp.Body.Read(buf); err != nil && err != io.EOF {
		return fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("[Anthropic] Triggered successfully with model: %s", p.model)
	log.Printf("[Anthropic] Response status: %s", resp.Status)
	log.Printf("[Anthropic] Response body: %s", string(buf))

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
