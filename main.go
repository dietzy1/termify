package main

import (
	"context"
	"log"
	"sync"

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

	var wg sync.WaitGroup

	appConfig := config.New()

	// Create the auth service with config
	authService, err := authentication.NewService(authentication.ServiceConfig{
		AppConfig: appConfig,
	})
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}
	// Start the server in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := authService.Start(ctx); err != nil && ctx.Err() == nil {
			log.Printf("Server error: %v", err)
			cancel()
		}
	}()

	// Run TUI - this will block until TUI exits
	if err := tui.Run(tui.Config{
		Ctx:         ctx,
		AuthService: authService,
	}); err != nil {
		log.Printf("TUI error: %v", err)
	}

	// TUI has exited, trigger shutdown
	cancel()

	//wg.Wait()
}
