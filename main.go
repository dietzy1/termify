package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dietzy1/termify/internal/authentication"
	"github.com/dietzy1/termify/internal/config"
	"github.com/dietzy1/termify/internal/tui"
)

func main() {
	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up logging if enabled
	if config.IsLoggingEnabled() {
		logPath := config.GetLogFilePath()
		f, err := tea.LogToFile(logPath, "debug")
		if err != nil {
			log.Printf("Failed to create log file: %v", err)
		} else {
			defer f.Close()
			log.Printf("Logging enabled, writing to %s", logPath)
		}
	} else {
		log.Println("Logging is disabled")
	}

	// Setup credential manager
	credManager, err := authentication.NewCredentialManager(config.GetClientID(), config.ConfigPath)
	if err != nil {
		log.Fatalf("Failed to create credential manager: %v", err)
	}

	// Setup authentication service client
	authService, err := authentication.NewService(config, credManager)
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}

	// Setup authentication server
	authServer := authentication.NewServer(config, authService)
	if err != nil {
		log.Fatalf("Failed to create authentication server: %v", err)
	}

	var authServiceErr error
	go func() {
		if err := authServer.ListenAndServe(ctx); err != nil && ctx.Err() == nil {
			log.Printf("Server error: %v", err)
			authServiceErr = fmt.Errorf("authentication server error: %w", err)
			cancel()
		}
	}()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				log.Printf("Memory usage - Alloc: %d KB, Sys: %d KB, NumGC: %d",
					m.Alloc/1024, m.Sys/1024, m.NumGC)
			}
		}
	}()

	// Run TUI - this will block until TUI exits
	if err := tui.Run(ctx, config, authService, authService); err != nil {
		log.Printf("TUI error: %v", err)
	}

	if authServiceErr != nil {
		fmt.Println("Termify encountered an error:", authServiceErr)
	}
}

// Track memory usage and other metrics
