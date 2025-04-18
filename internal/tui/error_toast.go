package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errorMsg struct {
	title   string
	message string
}

type ShowToastMsg struct {
	Title   string
	Message string
}

type ErrorTimerExpiredMsg struct{}

func (m applicationModel) renderErrorBar() string {
	if m.errorBar.title == "" || m.errorBar.message == "" {
		return ""
	}

	errorBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff4444")).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#ff4444")).
		Width(m.width-2).
		Height(2).
		Padding(0, 1)

	return errorBar.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().
			Bold(true).
			Render("Error: "+m.errorBar.title),
		"Details: "+m.errorBar.message,
	))
}

func ShowErrorToast(title, message string) tea.Cmd {
	return func() tea.Msg {
		return ShowToastMsg{
			Title:   title,
			Message: message,
		}
	}
}

func (m applicationModel) handleShowErrorMsg(msg ShowToastMsg) (applicationModel, tea.Cmd) {
	m.errorBar.title = msg.Title
	m.errorBar.message = msg.Message
	return m, tea.Batch(
		tea.WindowSize(),
		tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
			return ErrorTimerExpiredMsg{}
		}),
	)
}
