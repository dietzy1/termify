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

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer f.Close()

	// Load configuration
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup authentication service client
	authService, err := authentication.NewService(config)
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
	if err := tui.Run(ctx, config, authService); err != nil {
		log.Printf("TUI error: %v", err)
	}
}
