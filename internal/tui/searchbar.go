package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
)

var _ tea.Model = (*searchbarModel)(nil)

const debounceTime = time.Millisecond * 300

type debouncedSearch struct {
	searchTerm string
}

type searchbarModel struct {
	ctx   context.Context
	width int

	textInput textinput.Model
	searching bool
	isFocused bool

	spotifyState *state.SpotifyState
}

func newSearchbar(ctx context.Context, spotifyState *state.SpotifyState) searchbarModel {
	ti := textinput.New()
	ti.Placeholder = "What do you want to play?"
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
		ctx:          ctx,
		textInput:    ti,
		searching:    false,
		isFocused:    false,
		spotifyState: spotifyState,
	}
}

func (m searchbarModel) Init() tea.Cmd {
	return textinput.Blink
}

// EnterSearchMode enters search mode and returns appropriate commands
func (m *searchbarModel) EnterSearchMode() {
	if m.searching {
		return
	}
	m.searching = true
	m.textInput.Focus()
}

// ExitSearchMode exits search mode and returns appropriate commands
func (m *searchbarModel) ExitSearchMode() tea.Cmd {
	if !m.searching {
		return nil
	}
	m.searching = false
	m.textInput.SetValue("")
	m.textInput.Blur()
	return nil
}

func (m searchbarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.textInput.Width = m.width

	case debouncedSearch:
		if m.textInput.Value() == msg.searchTerm {
			return m, m.spotifyState.SearchEverything(m.ctx, msg.searchTerm)
		}

	case tea.KeyMsg:
		// Handle search input when in search mode
		if m.searching {
			switch {
			case key.Matches(msg, DefaultKeyMap.Return):
				m.searching = false
				m.textInput.SetValue("")
				m.textInput.Blur()
				return m, nil
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, tea.Sequence(cmd, tea.Tick(debounceTime, func(_ time.Time) tea.Msg {
					return debouncedSearch{
						searchTerm: m.textInput.Value(),
					}
				}))
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
