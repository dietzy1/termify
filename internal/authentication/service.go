package authentication

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/dietzy1/termify/internal/config"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type LoginUrlMsg struct {
	Url string
}

type LoginClientMsg struct {
	Client *spotify.Client
}

type LoginErrorMsg struct {
	Error error
}

type service struct {
	clientID      string
	state         string
	verifier      verifier
	authenticator *spotifyauth.Authenticator

	config      *config.Config
	credManager *credentialManager
	program     *tea.Program
}

// NewService creates a new authentication service
func NewService(c *config.Config, credManager *credentialManager) (*service, error) {

	s := &service{
		clientID:      "",
		verifier:      newVerifier(lenMax),
		authenticator: nil,
		config:        c,
		credManager:   credManager,
	}

	return s, nil
}

func (s *service) SetTeaProgram(p *tea.Program) {
	s.program = p
}

func (s *service) StartPkceAuth(clientID string) tea.Cmd {
	log.Println("Token is invalid or not found, generating new URL")
	url := s.generateUrl(clientID)
	return func() tea.Msg {
		return LoginUrlMsg{Url: url}
	}
}

// https://github.com/zmb3/spotify/issues/167
func (s *service) StartStoredTokenAuth(ctx context.Context) tea.Cmd {

	token, err := s.credManager.loadToken()
	if err != nil {
		log.Println("Failed to load token:", err)
	}

	clientID, err := s.credManager.loadClientID()
	if err != nil {
		log.Println("Failed to load client ID:", err)
		return nil
	}
	// Verify that its length is 32
	if len(clientID) != 32 {
		log.Println("Client ID is not valid: ", clientID)
		return nil
	}

	s.setAuthenticator(clientID)

	// Exit early if we have a valid token
	if token != nil && token.Valid() {
		log.Println("Token is valid, returning client")
		return func() tea.Msg {
			return LoginClientMsg{Client: spotify.New(
				s.authenticator.Client(ctx, token),
			)}
		}
	}

	// If the token is invalid, we need to refresh it
	if token != nil && token.RefreshToken != "" {
		log.Println("Token is invalid, attempting to refresh")

		// Create a TokenSource using the authenticator and our existing token
		refreshedToken, err := s.authenticator.RefreshToken(ctx, token)
		if err != nil {
			log.Println("Failed to refresh token:", err)
		}
		if refreshedToken != nil {
			log.Println("Token refreshed successfully")
			log.Println("Refreshed token:", refreshedToken)
			return func() tea.Msg {
				return LoginClientMsg{Client: spotify.New(s.authenticator.Client(ctx, refreshedToken))}
			}
		}
	}

	// If we reach here, the token is invalid or not found
	log.Println("Token is invalid or not found")

	// Generate a new URL for authentication
	url := s.generateUrl(clientID)
	log.Println("Generated URL:", url)
	return func() tea.Msg {
		return LoginUrlMsg{Url: url}
	}
}

func (s *service) generateUrl(clientID string) string {
	// Generate random state
	state := make([]byte, 16)
	for i := range state {
		state[i] = chars[rand.Intn(len(chars))]
	}

	s.clientID = clientID
	s.state = string(state)
	s.setAuthenticator(clientID)

	challenge := s.verifier.challenge()

	authParams := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("client_id", s.clientID),
	}
	authParams = append(authParams, challenge.params()...)

	url := s.authenticator.AuthURL(
		s.state,
		authParams...,
	)

	log.Printf("Generated login URL: %s", url)
	return url
}

func (s *service) setAuthenticator(clientID string) {
	s.authenticator = spotifyauth.New(
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithRedirectURL(fmt.Sprintf("http://127.0.0.1%s/callback", s.config.Server.Port)),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopePlaylistReadCollaborative,
			spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserModifyPlaybackState,
			spotifyauth.ScopeUserLibraryRead,
		),
	)
}

func (s *service) SaveToken(token *oauth2.Token) error {
	log.Println("Saving token")

	if err := s.credManager.saveToken(token); err != nil {
		log.Printf("Failed to save token: %v", err)
		return fmt.Errorf("failed to save token: %w", err)
	}
	log.Printf("Token successfully saved")
	return nil
}

func (s *service) saveClientID(clientID string) error {
	if err := s.credManager.saveClientID(clientID); err != nil {
		log.Printf("Failed to save client ID: %v", err)
		return fmt.Errorf("failed to save client ID: %w", err)
	}
	log.Printf("Client ID successfully saved")
	return nil
}
