package authentication

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const redirectURI = "http://127.0.0.1:8080/callback"

// AuthResult represents the current state of the authentication process
type AuthResult struct {
	LoginURL string
	Error    error
	Client   *spotify.Client
}

// SpotifyAuth handles Spotify authentication using PKCE flow
type SpotifyAuth struct {
	clientID string
	state    string
	verifier verifier

	auth         *spotifyauth.Authenticator
	credManager  *credentialManager
	authComplete chan AuthResult
}

func NewSpotifyAuth(cfg spotifyAuthConfig) (*SpotifyAuth, error) {
	if cfg.server == nil {
		return nil, fmt.Errorf("server cannot be nil")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "termify")
	credManager, err := NewCredentialManager(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential manager: %w", err)
	}

	return &SpotifyAuth{
		verifier:     newVerifier(lenMax),
		authComplete: make(chan AuthResult, 1),
		credManager:  credManager,
	}, nil
}

// Initialize sets up the Spotify authenticator with the provided client ID
func (s *SpotifyAuth) initialize(clientID string) {
	// Generate random state
	state := make([]byte, 16)
	for i := range state {
		state[i] = chars[rand.Intn(len(chars))]
	}

	s.state = string(state)
	s.clientID = clientID
	s.auth = spotifyauth.New(
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopePlaylistReadCollaborative,
			spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserModifyPlaybackState,
			spotifyauth.ScopeUserLibraryRead,
		),
	)
}

// StartAuth begins the authentication process
// It returns a channel that will receive auth status updates
func (s *SpotifyAuth) StartAuth(ctx context.Context, clientID string) chan AuthResult {
	log.Printf("StartAuth called with client ID: %s", clientID)
	s.initialize(clientID)

	// First, try to load existing token
	token, err := s.credManager.LoadToken()
	if err != nil {
		log.Printf("Failed to load token: %v", err)
		s.authComplete <- AuthResult{
			Error: fmt.Errorf("failed to load token: %w", err),
		}
		return s.authComplete
	}

	// If we have a valid token, create and return the client immediately
	if token != nil && token.Valid() {
		log.Printf("Found valid token, creating client and returning immediately")
		client := spotify.New(s.auth.Client(ctx, token))
		s.authComplete <- AuthResult{
			Client: client,
		}
		log.Printf("Closing authComplete channel (valid token path)")
		close(s.authComplete)
		log.Printf("Using existing credentials")
		return s.authComplete
	}

	log.Printf("No valid token found, starting PKCE flow")
	// If we have no token or it's expired, start the PKCE flow

	// Generate the challenge from verifier
	challenge := s.verifier.challenge()

	// Generate authorization URL with PKCE challenge
	authParams := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("client_id", s.clientID),
	}
	authParams = append(authParams, challenge.params()...)

	url := s.auth.AuthURL(
		s.state,
		authParams...,
	)

	log.Printf("Generated login URL: %s", url)
	s.authComplete <- AuthResult{
		LoginURL: url,
	}

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		log.Printf("Context cancelled, sending error to authComplete channel")
		s.authComplete <- AuthResult{
			Error: ctx.Err(),
		}
		close(s.authComplete)
	}()
	log.Printf("Returning authComplete channel")
	return s.authComplete
}

//go:embed login_complete.html
var content embed.FS

func (s *SpotifyAuth) completeAuth(w http.ResponseWriter, r *http.Request) {
	log.Println("Callback received, processing authentication completion")
	verifierParams := s.verifier.params()

	tok, err := s.auth.Token(r.Context(), s.state, r, verifierParams...)
	if err != nil {
		log.Printf("Token error: %v", err)
		s.authComplete <- AuthResult{
			Error: fmt.Errorf("token error: %w", err),
		}
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		log.Printf("Failed to get token: %v", err)
		return
	}
	log.Printf("Token successfully obtained")

	if st := r.FormValue("state"); st != s.state {
		log.Printf("State mismatch: expected %s, got %s", s.state, st)
		s.authComplete <- AuthResult{
			Error: fmt.Errorf("state mismatch"),
		}
		http.Error(w, "State verification failed", http.StatusBadRequest)
		return
	}
	log.Printf("State verification successful")

	if err := s.credManager.SaveToken(tok, s.clientID); err != nil {
		log.Printf("Failed to save token: %v", err)
		s.authComplete <- AuthResult{
			Error: fmt.Errorf("failed to save token: %w", err),
		}
		http.Error(w, "Failed to save credentials", http.StatusInternalServerError)
		return
	}
	log.Printf("Token successfully saved")

	client := spotify.New(s.auth.Client(r.Context(), tok))

	log.Println("Sending client to authComplete channel")
	s.authComplete <- AuthResult{
		Client: client,
	}
	log.Println("Closing authComplete channel")
	close(s.authComplete)

	// Read the embedded HTML file
	htmlContent, err := content.ReadFile("login_complete.html")
	if err != nil {
		log.Printf("Failed to load completion page: %v", err)
		http.Error(w, "Failed to load completion page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.Writer.Write(w, htmlContent)
	log.Println("Callback processing complete, HTML response sent")
}
