package tui

import (
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ShowToastMsg struct {
	Title   string
	Message string
}

type ErrorTimerExpiredMsg struct{}

type errorToastModel struct {
	width   int
	title   string
	message string
	active  bool
	timer   *time.Timer
}

func newErrorToast() errorToastModel {
	return errorToastModel{}
}

func (m errorToastModel) Init() tea.Cmd {
	return nil
}

func (m errorToastModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ShowToastMsg:
		log.Println("Received ShowToastMsg:", msg.Title, msg.Message)
		m.title = msg.Title
		m.message = msg.Message
		m.active = true

		if m.timer != nil {
			m.timer.Stop()
		}

		return m, tea.Batch(
			tea.WindowSize(),
			tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
				return ErrorTimerExpiredMsg{}
			}))

	case ErrorTimerExpiredMsg:
		m.active = false
		m.title = ""
		m.message = ""
		if m.timer != nil {
			m.timer = nil
		}

		return m, tea.WindowSize()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	}
	return m, nil
}

// View renders the error toast.
func (m errorToastModel) View() string {
	if !m.active || m.title == "" || m.message == "" {
		return ""
	}

	renderWidth := m.width - 2
	errorBar := lipgloss.NewStyle().
		Foreground(DangerColor).
		Border(lipgloss.NormalBorder()).
		BorderForeground(DangerColor).
		Width(renderWidth).
		Padding(0, 1).
		MaxHeight(4)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().
			Bold(true).
			Render("Error: "+m.title),
		"Details: "+m.message,
	)

	return errorBar.Render(content)
}

func (m errorToastModel) IsActive() bool {
	return m.active
}

func (m errorToastModel) Height() int {
	log.Println("Height in error toast model:", lipgloss.Height(m.View()))
	if !m.IsActive() {
		return 0
	}
	return 4
}

// ShowErrorToast is a helper function to create a command that shows the error toast.
func showErrorToast(title, message string) tea.Cmd {
	return func() tea.Msg {
		return ShowToastMsg{
			Title:   title,
			Message: message,
		}
	}
}

func clearErrorToastCmd() tea.Cmd {
	return func() tea.Msg {
		return ErrorTimerExpiredMsg{}
	}
}
