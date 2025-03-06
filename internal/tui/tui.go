package tui

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dietzy1/termify/internal/authentication"
	"github.com/zmb3/spotify/v2"
)

var _ tea.Model = (*model)(nil)

func (m model) Init() tea.Cmd {
	switch m.state {
	case authenticating:
		return tea.Batch(
			tea.SetWindowTitle("Authenticating"),
			m.authModel.Init(),
		)
	case application:
		return tea.Batch(
			tea.SetWindowTitle("Termify"),
			m.applicationModel.Init(),
		)
	}
	return nil
}

type AppState int

const (
	authenticating AppState = iota
	application
)

type model struct {
	width, height int
	state         AppState

	authModel        authModel
	applicationModel applicationModel
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
		state:            authenticating,
		authModel:        newAuthModel(cfg.AuthService),
		applicationModel: newApplication(nil),
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

func transitionToApplication(m model, spotifyClient *spotify.Client) model {
	return model{
		width:            m.width,
		height:           m.height,
		state:            application,
		authModel:        m.authModel,
		applicationModel: newApplication(spotifyClient),
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case authenticating:
		log.Printf("Updating authentication model with message type: %T", msg)
		return m.updateAuth(msg)
	case application:
		log.Printf("Updating application model with message type: %T", msg)
		return m.applicationModel.Update(msg)
	}
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case authenticating:
		return m.viewAuth()
	case application:
		return m.applicationModel.View()
	}
	return "Illegal state"
}
