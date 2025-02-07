package tui

import (
	"context"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/authentication"
	"github.com/zmb3/spotify/v2"
)

var _ tea.Model = (*model)(nil)

func (m model) Init() tea.Cmd {
	switch m.state {
	case authenticating:
		return tea.Batch(
			tea.SetWindowTitle("✨Authenticating✨"),
			textinput.Blink,
			m.authModel.Init(),
		)
	case ready:
		m.library = newLibrary(m.spotifyClient)
		return tea.Batch(
			tea.SetWindowTitle("✨Termify✨"),
			m.audioPlayer.Init(),
			m.library.Init(),
		)
	}
	return nil
}

type AppState int

const (
	authenticating AppState = iota
	ready
)

type model struct {
	width, height int
	state         AppState

	// Auth related fields
	authModel authModel

	// Spotify client and state
	spotifyClient *spotify.Client
	spotifyState  *SpotifyState

	// UI Components
	focusedModel    FocusedModel
	navbar          navbarModel
	searchBar       searchbarModel
	library         libraryModel
	viewport        viewportModel
	playbackControl playbackControlsModel
	audioPlayer     audioPlayerModel
}

// Config holds the TUI configuration
type Config struct {
	Ctx         context.Context
	AuthService authentication.Service
}

func Run(cfg Config) error {
	if cfg.Ctx == nil {
		return fmt.Errorf("context is required")
	}

	m := model{
		state:           authenticating,
		authModel:       newAuthModel(cfg.AuthService),
		focusedModel:    FocusViewport,
		navbar:          newNavbar(),
		searchBar:       newSearchbar(),
		library:         newLibrary(nil),
		viewport:        newViewport(),
		playbackControl: newPlaybackControlsModel(),
		audioPlayer:     newAudioPlayer(),
		spotifyClient:   nil,
		spotifyState:    nil,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	// Handle context cancellation
	go func() {
		<-cfg.Ctx.Done()
		p.Quit()
	}()

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle global window size messages first
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		return m.handleWindowSizeMsg(msg)
	}

	// Handle state-specific updates
	switch m.state {
	case authenticating:
		return m.updateAuth(msg)
	case ready:
		log.Printf("Updating primary model with message type: %T", msg)
		// Update the application and collect any commands
		updatedModel, cmd := m.updateApplication(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Type assert back to our model type
		if m, ok := updatedModel.(model); ok {
			return m, tea.Batch(cmds...)
		}
		return updatedModel, tea.Batch(cmds...)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	switch m.state {
	case authenticating:
		return m.viewAuth()
	case ready:
		return m.viewMain()
	}
	return ""
}

// Helper function to update submodels
func updateSubmodel[T any](model tea.Model, msg tea.Msg, targetType T) (T, tea.Cmd, bool) {
	updatedModel, cmd := model.Update(msg)

	typedModel, ok := updatedModel.(T)
	if !ok {
		return targetType, tea.Quit, false
	}

	return typedModel, cmd, true
}

func (m model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (model, tea.Cmd) {
	m.width, m.height = msg.Width, msg.Height

	var cmds []tea.Cmd
	msg.Height -= lipgloss.Height(m.navbar.View()) + lipgloss.Height(m.playbackControl.View()) + lipgloss.Height(m.audioPlayer.View()) + 4

	if updatedNavbar, cmd, ok := updateSubmodel(m.navbar, tea.WindowSizeMsg{
		Width: msg.Width - 2,
	}, m.navbar); ok {
		m.navbar = updatedNavbar
		cmds = append(cmds, cmd)
	} else {
		return m, tea.Quit
	}

	if updatedLibrary, cmd, ok := updateSubmodel(m.library, tea.WindowSizeMsg{
		Width:  28,
		Height: msg.Height,
	}, m.library); ok {
		m.library = updatedLibrary
		cmds = append(cmds, cmd)
	} else {
		return m, tea.Quit
	}

	if updatedViewport, cmd, ok := updateSubmodel(m.viewport, tea.WindowSizeMsg{
		Width:  msg.Width - 4 - m.library.list.Width(),
		Height: msg.Height,
	}, m.viewport); ok {
		m.viewport = updatedViewport
		cmds = append(cmds, cmd)
	} else {
		return m, tea.Quit
	}

	if updatedPlaybackControl, cmd, ok := updateSubmodel(m.playbackControl, tea.WindowSizeMsg{
		Width: msg.Width - 2,
	}, m.playbackControl); ok {
		m.playbackControl = updatedPlaybackControl
		cmds = append(cmds, cmd)
	} else {
		return m, tea.Quit
	}

	if updatedAudioPlayer, cmd, ok := updateSubmodel(m.audioPlayer, tea.WindowSizeMsg{
		Width: msg.Width - 2,
	}, m.audioPlayer); ok {
		m.audioPlayer = updatedAudioPlayer
		cmds = append(cmds, cmd)
	} else {
		return m, tea.Quit
	}

	return m, tea.Batch(cmds...)
}
