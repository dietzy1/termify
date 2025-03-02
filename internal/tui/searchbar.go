package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ tea.Model = (*searchbarModel)(nil)

type searchbarModel struct {
	width     int
	textInput textinput.Model
	searching bool
}

func newSearchbar() searchbarModel {
	ti := textinput.New()
	ti.Placeholder = "Search tracks..."
	ti.CharLimit = 50
	ti.Width = 30

	// Apply styles
	ti.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor))

	ti.TextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor))

	ti.PlaceholderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor))

	return searchbarModel{
		textInput: ti,
		searching: false,
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
		m.textInput.Width = m.width - 2

	case tea.KeyMsg:
		// Handle search mode toggle
		if key.Matches(msg, key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search"))) {
			m.searching = !m.searching
			if m.searching {
				m.textInput.Focus()
				return m, textinput.Blink
			} else {
				m.textInput.Blur()
			}
			return m, nil
		}

		// Handle search input when in search mode
		if m.searching {
			switch msg.String() {
			case "esc":
				m.searching = false
				m.textInput.Blur()
				return m, nil
			case "enter":
				m.searching = false
				m.textInput.Blur()
				return m, nil
			default:
				var inputCmd tea.Cmd
				m.textInput, inputCmd = m.textInput.Update(msg)
				return m, inputCmd
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m searchbarModel) View() string {
	// Create search bar style
	searchStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(BorderColor)).
		Padding(0, 1).
		Width(m.width - 2)

	// Search bar with indicator
	searchPrefix := "🔍 "
	if !m.searching {
		searchPrefix = "/ "
	}

	return searchStyle.Render(searchPrefix + m.textInput.View())
}
