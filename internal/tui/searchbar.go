package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
)

var _ tea.Model = (*searchbarModel)(nil)

type searchbarModel struct {
	width int

	textInput textinput.Model
	searching bool
	isFocused bool

	spotifyState *state.SpotifyState
}

func newSearchbar(spotifyState *state.SpotifyState) searchbarModel {
	ti := textinput.New()
	ti.Placeholder = "Search tracks..."
	ti.CharLimit = 25
	ti.Width = 30

	// Apply styles
	ti.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor))

	ti.TextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor))

	ti.PlaceholderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor))

	return searchbarModel{
		textInput:    ti,
		searching:    false,
		isFocused:    false,
		spotifyState: spotifyState,
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
		m.textInput.Width = m.width

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

				// Perform search
				searchTerm := m.textInput.Value()
				log.Printf("Viewport: Searching for tracks with term: %s", searchTerm)
				return m, m.spotifyState.SearchEverything(searchTerm)

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
	searchStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Width(m.width - 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(getBorderStyle(m.isFocused))

	searchPrefix := "🔍 "
	if !m.searching {
		searchPrefix = "/ "
	}

	return searchStyle.Render(searchPrefix + m.textInput.View())
}
