package tui

import (
	"context"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dietzy1/termify/internal/authentication"
	"github.com/dietzy1/termify/internal/config"
	"github.com/zmb3/spotify/v2"
)

type authenticator interface {
	StartPkceAuth(ctx context.Context, clientID string) tea.Cmd
	StartStoredTokenAuth(ctx context.Context) tea.Cmd
	SetTeaProgram(p *tea.Program)
}

var _ tea.Model = (*model)(nil)

type tuiState int

const (
	authenticating tuiState = iota
	application
)

type model struct {
	width, height int
	state         tuiState

	authModel        authModel
	applicationModel applicationModel
}

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

func Run(ctx context.Context, c *config.Config, authenticator authenticator) error {

	m := model{
		state:            authenticating,
		authModel:        newAuthModel(c, authenticator),
		applicationModel: newApplication(nil),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	authenticator.SetTeaProgram(p)

	// Handle context cancellation
	go func() {
		<-ctx.Done()
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

	switch msg := msg.(type) {
	case authentication.LoginClientMsg:
		log.Printf("Parent model (tui.go): Received LoginClientMsg. Transitioning to application state. Client valid: %v", msg.Client != nil)
		if msg.Client == nil {
			log.Println("Parent model (tui.go): LoginClientMsg has a nil client. Authentication might have failed silently earlier or an error message was missed.")
		}
		updatedModel := transitionToApplication(m, msg.Client)
		return updatedModel, updatedModel.Init()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Quit):
			return m, tea.Quit
		}
	}

	switch m.state {
	case authenticating:
		if updatedAuth, cmd, ok := updateSubmodel(m.authModel, msg, m.authModel); ok {
			m.authModel = updatedAuth
			return m, cmd
		}

	case application:
		if updatedApplication, cmd, ok := updateSubmodel(m.applicationModel, msg, m.applicationModel); ok {
			m.applicationModel = updatedApplication
			return m, cmd
		}
	}
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case authenticating:
		return m.authModel.View()
	case application:
		return m.applicationModel.View()
	}
	return "Illegal state"
}
