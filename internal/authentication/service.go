package authentication

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dietzy1/termify/internal/config"
)

type ServiceConfig struct {
	AppConfig *config.Config
}

type spotifyAuthConfig struct {
	server *service
}

// Service defines the interface for authentication operations
type Service interface {
	StartAuth(ctx context.Context, clientID string) chan AuthResult
	GetClientID() string
	// Server management methods
	Start(ctx context.Context) error
	Stop() error
}

type service struct {
	config *config.Config
	server *http.Server
	mux    *http.ServeMux
	auth   *SpotifyAuth
}

// getUserConfigDir returns the user's config directory
func getUserConfigDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(userConfigDir, "termify")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return configDir, nil
}

// NewService creates a new authentication service
func NewService(cfg ServiceConfig) (Service, error) {
	if cfg.AppConfig == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	mux := http.NewServeMux()
	s := &service{
		config: cfg.AppConfig,
		server: &http.Server{
			Addr:    cfg.AppConfig.GetPort(),
			Handler: mux,
		},
		mux: mux,
	}

	// Initialize Spotify auth with the credential manager
	auth, err := NewSpotifyAuth(spotifyAuthConfig{
		server: s,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create spotify auth: %w", err)
	}
	s.auth = auth

	mux.HandleFunc("/callback", s.auth.completeAuth)

	return s, nil
}

// StartAuth initiates the Spotify authentication process
func (s *service) StartAuth(ctx context.Context, clientID string) chan AuthResult {
	if s.auth == nil {
		resultChan := make(chan AuthResult, 1)
		resultChan <- AuthResult{
			Error: fmt.Errorf("authentication service not properly initialized"),
		}
		close(resultChan)
		return resultChan
	}
	log.Printf("Starting auth with client ID: %s", clientID)
	return s.auth.StartAuth(ctx, clientID)
}

// GetClientID returns the Spotify client ID from config or credentials
func (s *service) GetClientID() string {
	// First try to get from config
	clientID := s.config.GetClientID()
	if clientID != "" {
		return clientID
	}

	// If not in config, try to load from credentials
	var err error
	clientID, err = s.auth.credManager.LoadClientID()
	if err != nil {
		log.Printf("Failed to load client ID from credentials: %v", err)
		return ""
	}

	// If found in credentials, update the config
	if clientID != "" {
		if err := s.config.SetClientID(clientID); err != nil {
			log.Printf("Failed to save client ID to config: %v", err)
		}
	}

	return clientID
}

func (s *service) Start(ctx context.Context) error {
	errChan := make(chan error, 1)
	log.Println("Starting authentication service")
	go func() {
		log.Printf("Starting callback server on %s", s.config.GetPort())
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server error: %w", err)
		}
		close(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return s.Stop()
	}
}

func (s *service) Stop() error {
	if s.server != nil {
		log.Println("Stopping callback server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	}
	return nil
}
