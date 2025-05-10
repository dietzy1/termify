package authentication

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	tea "github.com/charmbracelet/bubbletea"
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
func NewService(c *config.Config) (*service, error) {

	configDirPath, err := getUserConfigDir(c.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find valid configuration directory for user due to: %w", err)
	}

	log.Println("Config directory path:", configDirPath)

	credManager, err := NewCredentialManager(configDirPath)
	if err != nil {
		/* return nil, fmt.Errorf("failed to create credential manager: %w", err) */
	}

	s := &service{
		clientID:      "",
		verifier:      newVerifier(lenMax),
		authenticator: nil,

		config:      c,
		credManager: credManager,
	}

	return s, nil
}

func (s *service) SetTeaProgram(p *tea.Program) {
	s.program = p
}

func (s *service) StartPkceAuth(ctx context.Context, clientID string) tea.Cmd {

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
		return func() tea.Msg {
			return LoginErrorMsg{Error: fmt.Errorf("failed to load token: %w", err)}
		}
	}

	log.Println("Loaded token:", token)

	// We must also ensure that we have a valid clientID stored somewhere
	// Either check in storage or use the one found in the config
	/* clientID, err := s.credManager.loadClientID()
	if err != nil {
		log.Println("Failed to load client ID:", err)
		return nil
	}
	s.setAuthenticator(clientID) */

	// We must beforehand ensure we have initialized the authenticator
	s.setAuthenticator("beff5495d8fa419fb4040e4618e838d0")

	//TODO: We need to write some sort of intercepter for the oath2 token since it will be refreshed in the background.
	// This means that our locally stored token will be invalidated on the spotify side and our local token is therefor out of sync and invalid.
	// We must therefor ensure that on refresh then we also update the local token.
	// We are able to check the token by doing client.Token()

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
			return func() tea.Msg {
				return LoginErrorMsg{Error: fmt.Errorf("failed to refresh token: %w", err)}
			}
		}
		log.Println("Token refreshed successfully")
		log.Println("Refreshed token:", refreshedToken)

		return func() tea.Msg {
			return LoginClientMsg{Client: spotify.New(s.authenticator.Client(ctx, refreshedToken))}
		}
	}
	// If we reach here, the token is invalid or not found
	log.Println("Token is invalid or not found")
	return nil
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

func (s *service) saveToken(token *oauth2.Token) error {
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
