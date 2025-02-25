package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ tea.Model = (*searchbarModel)(nil)

type searchbarModel struct {
	width     int
	textInput textinput.Model
}

func newSearchbar() searchbarModel {
	ti := textinput.New()
	ti.Placeholder = "Search for a song"
	ti.CharLimit = 156
	ti.Width = 100

	// Apply styles
	ti.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor))

	ti.TextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor))

	ti.PlaceholderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor))

	// Wrapper for the text input to add background and border

	return searchbarModel{
		textInput: ti,
	}
}

func (m searchbarModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m searchbarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m searchbarModel) View() string {
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(BorderColor))

	return borderStyle.Render(m.textInput.View())
}
