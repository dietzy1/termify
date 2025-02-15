package tui

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/authentication"
	"github.com/pkg/browser"
	"github.com/zmb3/spotify/v2"
)

// Message types for authentication flow
type AuthState int

const (
	StateAwaitingClientID AuthState = iota
	StateAwaitingLogin
	StateAuthComplete
)

type authModel struct {
	input    textinput.Model
	auth     authentication.Service
	state    AuthState
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
		state:    StateAwaitingClientID,
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
		result := <-authChan
		log.Println("Waited for auth result:", result)
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
		if msg.Error != nil {
			log.Printf("Authentication error: %v", msg.Error)
			m.authModel.err = msg.Error
			m.authModel.state = StateAwaitingClientID
			return m, nil
		}
		if msg.LoginURL != "" {
			if err := browser.OpenURL(msg.LoginURL); err != nil {
				log.Fatal(err)
			}

			log.Printf("Received login URL: %s", msg.LoginURL)
			m.authModel.loginURL = msg.LoginURL
			m.authModel.state = StateAwaitingLogin
			m.authModel.copied = false // Reset copied state when showing new URL
			return m, waitForAuth(m.authModel.authChan)
		}
		if msg.Client != nil {
			log.Printf("Authentication successful")
			m.authModel.state = StateAuthComplete
			m = transitionToReady(m, msg.Client)
			return m, m.Init()
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.authModel.state == StateAwaitingClientID {
				clientID := m.authModel.input.Value()

				//Manually check if client ID is valid
				if len(clientID) != 32 {
					m.authModel.err = errClientID
					return m, nil
				}

				m.authModel.err = nil

				// Store the auth channel and start listening
				m.authModel.authChan = m.authModel.auth.StartAuth(context.Background(), clientID)
				return m, waitForAuth(m.authModel.authChan)
			}
		case tea.KeyRunes:
			if m.authModel.state == StateAwaitingLogin && msg.String() == "c" {
				// Copy URL to clipboard and set copied state
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
	if m.authModel.state == StateAwaitingClientID {
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
		Foreground(lipgloss.Color("#FFFFFF")).
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
	case StateAwaitingClientID:
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

	case StateAwaitingLogin:
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

	case StateAuthComplete:
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
	clientID := m.auth.GetClientID()
	if clientID == "" {
		return nil
	}
	log.Printf("Found stored client ID: %s", clientID)

	m.authChan = m.auth.StartAuth(context.Background(), clientID)
	return waitForAuth(m.authChan)
}

func safelyRenderError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
