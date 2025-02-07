package authentication

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
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
	mu     sync.Mutex
	auth   *SpotifyAuth
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
			Addr:    cfg.AppConfig.Port,
			Handler: mux,
		},
		mux:  mux,
		mu:   sync.Mutex{},
		auth: nil,
	}

	// Initialize SpotifyAuth with the service as the CallbackSetter
	auth, err := NewSpotifyAuth(spotifyAuthConfig{
		server: s,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create spotify auth: %w", err)
	}
	mux.HandleFunc("/callback", auth.completeAuth)

	s.auth = auth

	return s, nil
}

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

func (s *service) GetClientID() string {
	if s.auth == nil {
		return ""
	}

	token, err := s.auth.credManager.LoadClientID()
	if err != nil {
		log.Printf("Failed to load client ID: %v", err)
		return ""
	}

	return token
}

func (s *service) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	errChan := make(chan error, 1)
	log.Println("Starting authentication service")
	go func() {
		log.Printf("Starting callback server on %s", s.config.Port)
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
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		log.Println("Stopping callback server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	}
	return nil
}
