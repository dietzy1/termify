package tui

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/authentication"
	"github.com/pkg/browser"
	"github.com/zmb3/spotify/v2"
)

// Message types for authentication flow
type authState int

const (
	stateAwaitingClientID authState = iota
	stateAwaitingLogin
	stateAuthComplete
)

type authModel struct {
	input    textinput.Model
	auth     authentication.Service
	state    authState
	err      error
	loginURL string
	authChan <-chan authentication.AuthResult
	copied   bool
}

func newAuthModel(auth authentication.Service) authModel {
	textInput := textinput.New()
	textInput.Placeholder = "Enter your Spotify Client ID"
	textInput.Focus()
	textInput.CharLimit = 32
	textInput.Width = 40

	return authModel{
		auth:     auth,
		input:    textInput,
		state:    stateAwaitingClientID,
		err:      nil,
		loginURL: "",
		authChan: nil,
		copied:   false,
	}
}

// AuthChannelMsg represents a message from the auth channel
type AuthChannelMsg struct {
	LoginURL string
	Client   *spotify.Client
	Error    error
}

// Message type for resetting copy state
type resetCopyMsg struct{}

// Command to reset copy state after a delay
func resetCopyAfterDelay() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return resetCopyMsg{}
	})
}

// waitForAuth returns a command that listens to the auth channel
func waitForAuth(authChan <-chan authentication.AuthResult) tea.Cmd {
	return func() tea.Msg {
		log.Println("Waiting for auth result from channel...")
		result, ok := <-authChan
		if !ok {
			log.Println("Auth channel closed without receiving any result")
			// Return an empty AuthChannelMsg to avoid blocking
			return AuthChannelMsg{}
		}

		log.Println("Received auth result:", result)

		// If we got a login URL, we need to continue waiting for the client
		if result.LoginURL != "" && result.Client == nil && result.Error == nil {
			log.Println("Auth result contains only a login URL, returning it and continuing to wait")
			// Return the login URL message but also set up a command to continue waiting
			return tea.Batch(
				func() tea.Msg {
					return AuthChannelMsg{
						LoginURL: result.LoginURL,
					}
				},
				waitForNextAuthResult(authChan),
			)()
		}

		// Otherwise, return the result as is
		if result.Client != nil {
			log.Println("Auth result contains a client")
		}
		if result.Error != nil {
			log.Println("Auth result contains an error:", result.Error)
		}

		return AuthChannelMsg{
			LoginURL: result.LoginURL,
			Client:   result.Client,
			Error:    result.Error,
		}
	}
}

// waitForNextAuthResult continues waiting for the next message from the auth channel
func waitForNextAuthResult(authChan <-chan authentication.AuthResult) tea.Cmd {
	return func() tea.Msg {
		log.Println("Continuing to wait for auth result...")
		result, ok := <-authChan
		if !ok {
			log.Println("Auth channel closed without receiving client or error")
			// Return an empty AuthChannelMsg to avoid blocking
			return AuthChannelMsg{}
		}

		log.Println("Received next auth result:", result)
		if result.Client != nil {
			log.Println("Auth result contains a client")
		}
		if result.Error != nil {
			log.Println("Auth result contains an error:", result.Error)
		}

		return AuthChannelMsg{
			LoginURL: result.LoginURL,
			Client:   result.Client,
			Error:    result.Error,
		}
	}
}

var errClientID = fmt.Errorf("client ID must be 32 characters")

// TODO: I think this should be refactored into using m authModel - For consistency issues
func (m model) updateAuth(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case resetCopyMsg:
		m.authModel.copied = false
		return m, nil
	case AuthChannelMsg:
		log.Printf("Processing AuthChannelMsg: Error=%v, Client=%v, LoginURL=%v",
			msg.Error != nil,
			msg.Client != nil,
			msg.LoginURL != "")
		if msg.Error != nil {
			log.Printf("Authentication error: %v", msg.Error)
			m.authModel.err = msg.Error
			m.authModel.state = stateAwaitingClientID
			return m, nil
		}
		if msg.Client != nil {
			log.Printf("Authentication successful, transitioning to application")
			m.authModel.state = stateAuthComplete
			m = transitionToApplication(m, msg.Client)
			return m, m.Init()
		}
		if msg.LoginURL != "" {
			if err := browser.OpenURL(msg.LoginURL); err != nil {
				log.Fatal(err)
			}
			log.Printf("Received login URL: %s", msg.LoginURL)
			m.authModel.loginURL = msg.LoginURL

			// If we're already in the StateAwaitingLogin state, we don't need to set up another waitForAuth command
			// This happens when we're using a stored client ID and we've already set up the waitForAuth command in Init()
			alreadyWaiting := m.authModel.state == stateAwaitingLogin

			m.authModel.state = stateAwaitingLogin
			m.authModel.copied = false // Reset copied state when showing new URL

			if alreadyWaiting {
				log.Println("Already in StateAwaitingLogin, not setting up another waitForAuth command")
				return m, nil
			}

			log.Println("Setting up waitForAuth command")
			return m, waitForAuth(m.authModel.authChan)
		}

		log.Printf("AuthChannelMsg didn't match any condition, no action taken")
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Select):
			if m.authModel.state == stateAwaitingClientID {
				clientID := m.authModel.input.Value()
				if len(clientID) != 32 {
					m.authModel.err = errClientID
					return m, nil
				}
				m.authModel.err = nil
				// Store the auth channel and start listening
				m.authModel.authChan = m.authModel.auth.StartAuth(context.Background(), clientID)
				return m, waitForAuth(m.authModel.authChan)
			}
		case key.Matches(msg, DefaultKeyMap.Copy):
			if m.authModel.state == stateAwaitingLogin && msg.String() == "c" {
				if err := clipboard.WriteAll(m.authModel.loginURL); err != nil {
					log.Printf("Failed to copy to clipboard: %v", err)
					return m, nil
				}
				m.authModel.copied = true
				return m, resetCopyAfterDelay()
			}
		}
	}
	// Handle text input updates only if we're awaiting client ID
	if m.authModel.state == stateAwaitingClientID {
		m.authModel.input, cmd = m.authModel.input.Update(msg)
	}
	return m, cmd
}

func (m model) viewAuth() string {
	containerStyle := lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Width(m.width).
		Height(m.height)

	LogoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor))

	// Create a box for the main content
	boxStyle := lipgloss.NewStyle().
		Padding(2).
		Width(72). // Fixed width to ensure consistent alignment
		Align(lipgloss.Center)

	inputStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(BorderColor)).
		Padding(1, 2).
		Width(66)

	textStyle := lipgloss.NewStyle().
		PaddingBottom(1)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor)).
		Bold(true).
		MarginTop(1).
		MarginBottom(1)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#999999")).
		Italic(true).
		MarginBottom(1)

	instructionStyle := lipgloss.NewStyle().
		Foreground(WhiteTextColor).
		Width(66). // Fixed width for instructions
		MarginTop(1).
		MarginBottom(1)

	bulletStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor)).
		SetString("• ")

	urlStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor))

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		MarginTop(1).
		Italic(true)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		MarginTop(1)

	// Define styles for key hints
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor))

	keyHint := lipgloss.JoinHorizontal(
		lipgloss.Top,
		"Press ",
		keyStyle.Render("Enter"),
		" to continue | ",
		keyStyle.Render("Ctrl+C"),
		" to exit",
	)

	switch m.authModel.state {
	case stateAwaitingClientID:
		instructions := []string{
			"Getting started:",
			"1. Visit https://developer.spotify.com/dashboard",
			"2. Create a new application in your Spotify Developer Dashboard",
			"3. Set the redirect URI to http://127.0.0.1:8080/callback",
			"4. Copy your Client ID from the dashboard",
		}

		// Create bullet points with the instructions
		formattedInstructions := make([]string, len(instructions))
		for i, instruction := range instructions {
			if i == 0 {
				formattedInstructions[i] = instructionStyle.Render(instruction)
			} else {
				formattedInstructions[i] = lipgloss.JoinHorizontal(
					lipgloss.Left,
					"  ",
					bulletStyle.String(),
					instruction,
				)
			}
		}

		return containerStyle.Render(lipgloss.JoinVertical(
			lipgloss.Center,
			LogoStyle.Render(logo),
			boxStyle.Render(lipgloss.JoinVertical(
				lipgloss.Center,
				titleStyle.Render("Welcome to Termify!"),
				subtitleStyle.Render("Your terminal-based Spotify client"),
				lipgloss.JoinVertical(
					lipgloss.Left,
					formattedInstructions...,
				),

				errorStyle.Render(
					safelyRenderError(m.authModel.err),
				),
				inputStyle.Render(m.authModel.input.View()),
				keyHint,
			)),
		))

	case stateAwaitingLogin:
		statusText := "Press c to copy URL"
		if m.authModel.copied {
			statusText = "✓ URL copied to clipboard!"
		}

		return containerStyle.Render(lipgloss.JoinVertical(
			lipgloss.Center,
			LogoStyle.Render(logo),
			boxStyle.Render(lipgloss.JoinVertical(
				lipgloss.Center,
				instructionStyle.Render("Please visit this URL to login:"),
				urlStyle.Render(m.authModel.loginURL),
				hintStyle.Render(statusText),
			)),
		))

	case stateAuthComplete:
		return containerStyle.Render(lipgloss.JoinVertical(
			lipgloss.Center,
			LogoStyle.Render(logo),
			boxStyle.Render(
				textStyle.Render("Authentication successful!"),
			),
		))

	default:
		return containerStyle.Render(lipgloss.JoinVertical(
			lipgloss.Center,
			LogoStyle.Render(logo),
			boxStyle.Render(
				errorStyle.Render("Unknown authentication state"),
			),
		))
	}
}

// The change we need to make here is that it needs to check the credential manager for a stored client ID
// If that exists then we can change the state to StateAwaitingLogin and start the auth process
func (m *authModel) Init() tea.Cmd {
	log.Println("Initializing authModel")
	clientID := m.auth.GetClientID()
	if clientID == "" {
		log.Println("No stored client ID found, waiting for user input")
		return nil
	}
	log.Printf("Found stored client ID: %s", clientID)
	log.Println("Setting state to StateAwaitingLogin")
	m.state = stateAwaitingLogin

	log.Println("Starting auth with stored client ID")
	m.authChan = m.auth.StartAuth(context.Background(), clientID)
	log.Println("Returning waitForAuth command")
	return waitForAuth(m.authChan)
}
