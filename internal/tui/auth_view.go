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
	"github.com/dietzy1/termify/internal/config"
	"github.com/pkg/browser"
)

var _ tea.Model = (*authModel)(nil)

// Message types for authentication flow
type authState int

const (
	stateAwaitingClientID authState = iota
	stateAwaitingLogin
)

type authModel struct {
	width, height int
	input         textinput.Model
	state         authState
	err           error
	loginURL      string
	copied        bool
	config        *config.Config
	authenticator authenticator
}

func newAuthModel(c *config.Config, authenticator authenticator) authModel {
	textInput := textinput.New()
	textInput.Placeholder = "Enter your Spotify Client ID"
	textInput.Focus()
	textInput.CharLimit = 32
	textInput.Width = 40

	return authModel{
		input:         textInput,
		state:         stateAwaitingClientID,
		err:           nil,
		loginURL:      "",
		copied:        false,
		config:        c,
		authenticator: authenticator,
	}
}

type resetCopyMsg struct{}

// Command to reset copy state after a delay
func resetCopyAfterDelay() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return resetCopyMsg{}
	})
}

var errClientID = fmt.Errorf("client ID must be 32 characters")

func (m authModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case resetCopyMsg:
		m.copied = false
		return m, nil

	case authentication.LoginUrlMsg:
		log.Printf("Received login URL: %s", msg.Url)
		if err := browser.OpenURL(msg.Url); err != nil {
			log.Fatal(err)
		}
		m.loginURL = msg.Url
		m.state = stateAwaitingLogin
		m.copied = false

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Select):
			if m.state == stateAwaitingClientID {
				clientID := m.input.Value()
				if len(clientID) != 32 {
					m.err = errClientID
					return m, nil
				}
				m.err = nil
				return m, m.authenticator.StartPkceAuth(context.TODO(), clientID)

			}
		case key.Matches(msg, DefaultKeyMap.Copy):
			if m.state == stateAwaitingLogin && msg.String() == "c" {
				if err := clipboard.WriteAll(m.loginURL); err != nil {
					log.Printf("Failed to copy to clipboard: %v", err)
					return m, nil
				}
				m.copied = true
				return m, resetCopyAfterDelay()
			}
		}
	}
	// Handle text input updates only if we're awaiting client ID
	if m.state == stateAwaitingClientID {
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m authModel) View() string {
	containerStyle := lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Width(m.width).
		Height(m.height)

	LogoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor))

	// Create a box for the main content
	boxStyle := lipgloss.NewStyle().
		Padding(2).
		Width(72).
		Align(lipgloss.Center)

	inputStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(BorderColor)).
		Padding(1, 2).
		Width(66)

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
		Width(66).
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

	switch m.state {
	case stateAwaitingClientID:
		instructions := []string{
			"Getting started:",
			"1. Visit https://developer.spotify.com/dashboard",
			"2. Create a new application in your Spotify Developer Dashboard",
			fmt.Sprintf("3. Set the redirect URI to http://127.0.0.1%s/callback", m.config.Server.Port),
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
					safelyRenderError(m.err),
				),
				inputStyle.Render(m.input.View()),
				keyHint,
			)),
		))

	case stateAwaitingLogin:
		statusText := "Press c to copy URL"
		if m.copied {
			statusText = "✓ URL copied to clipboard!"
		}

		return containerStyle.Render(lipgloss.JoinVertical(
			lipgloss.Center,
			LogoStyle.Render(logo),
			boxStyle.Render(lipgloss.JoinVertical(
				lipgloss.Center,
				instructionStyle.Render("Please visit this URL to login:"),
				urlStyle.Render(m.loginURL),
				hintStyle.Render(statusText),
			)),
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

func (m authModel) Init() tea.Cmd {
	return m.authenticator.StartStoredTokenAuth(context.TODO())
}
