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
		verifier:     newVerifier(LenMax),
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
	s.initialize(clientID)

	// First, try to load existing token
	token, err := s.credManager.LoadToken()
	if err != nil {
		s.authComplete <- AuthResult{
			Error: fmt.Errorf("failed to load token: %w", err),
		}
		log.Printf("Failed to load token: %v", err)
		return s.authComplete
	}

	// If we have a valid token, create and return the client immediately
	if token != nil && token.Valid() {
		client := spotify.New(s.auth.Client(ctx, token))
		s.authComplete <- AuthResult{
			Client: client,
		}
		close(s.authComplete)
		//TODO:  Should we also just close the webserver here?

		log.Printf("Using existing credentials")
		return s.authComplete
	}

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

	s.authComplete <- AuthResult{
		LoginURL: url,
	}

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		s.authComplete <- AuthResult{
			Error: ctx.Err(),
		}
		close(s.authComplete)
	}()
	return s.authComplete
}

//go:embed login_complete.html
var content embed.FS

func (s *SpotifyAuth) completeAuth(w http.ResponseWriter, r *http.Request) {
	verifierParams := s.verifier.params()

	tok, err := s.auth.Token(r.Context(), s.state, r, verifierParams...)
	if err != nil {
		s.authComplete <- AuthResult{
			Error: fmt.Errorf("token error: %w", err),
		}
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		log.Printf("Failed to get token: %v", err)
		return
	}

	if st := r.FormValue("state"); st != s.state {
		s.authComplete <- AuthResult{
			Error: fmt.Errorf("state mismatch"),
		}
		http.Error(w, "State verification failed", http.StatusBadRequest)
		return
	}

	if err := s.credManager.SaveToken(tok, s.clientID); err != nil {
		s.authComplete <- AuthResult{
			Error: fmt.Errorf("failed to save token: %w", err),
		}
		http.Error(w, "Failed to save credentials", http.StatusInternalServerError)
		return
	}

	client := spotify.New(s.auth.Client(r.Context(), tok))

	log.Println("Sending client to authComplete channel")
	s.authComplete <- AuthResult{
		Client: client,
	}

	// Read the embedded HTML file
	htmlContent, err := content.ReadFile("login_complete.html")
	if err != nil {
		http.Error(w, "Failed to load completion page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.Writer.Write(w, htmlContent)
}
