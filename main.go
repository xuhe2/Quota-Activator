package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"quota-activator/config"
	"quota-activator/platform"
	"quota-activator/scheduler"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	// Create platform instance
	p, err := platform.NewPlatform(&platform.PlatformInput{
		Type:    cfg.Platform.Type,
		BaseURL: cfg.Platform.BaseURL,
		Options: cfg.Platform.Options,
	})
	if err != nil {
		log.Fatalf("Failed to create platform: %v", err)
	}

	// Validate platform-specific config
	if err := p.ValidateConfig(); err != nil {
		log.Fatalf("Platform config validation failed: %v", err)
	}

	// Create and start scheduler
	sched := scheduler.New(&cfg.Scheduler, p)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutdown signal received, stopping...")
		cancel()
	}()

	// Start the scheduler (blocking call)
	if err := sched.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Scheduler error: %v", err)
	}

	log.Println("Quota-Activator stopped")
}
