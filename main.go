package main

import (
	"context"
	"log"

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

	go func() {
		if err := authServer.ListenAndServe(ctx); err != nil && ctx.Err() == nil {
			log.Printf("Server error: %v", err)
			cancel()
		}
	}()

	// Run TUI - this will block until TUI exits
	if err := tui.Run(ctx, config, authService, authService); err != nil {
		log.Printf("TUI error: %v", err)
	}
}
