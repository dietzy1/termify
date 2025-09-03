package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
)

var _ tea.Model = (*libraryModel)(nil)

// playlist implements list.Item interface
type playlist struct {
	title string
	desc  string
	uri   string
}

func (p playlist) Title() string       { return p.title }
func (p playlist) Description() string { return p.desc }
func (p playlist) FilterValue() string { return p.title }

type libraryModel struct {
	height       int
	list         list.Model
	spotifyState *state.SpotifyState
	isFocused    bool
}

func newLibrary(spotifyState *state.SpotifyState) libraryModel {
	delegate := list.NewDefaultDelegate()

	const itemWidth = 28

	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(WhiteTextColor).
		Padding(0, 0, 0, 2).
		Width(itemWidth).
		MaxWidth(itemWidth)

	delegate.Styles.NormalDesc = delegate.Styles.NormalTitle.
		Foreground(TextColor).
		Padding(0, 0, 0, 2).
		Width(itemWidth).
		MaxWidth(itemWidth)

	l := list.New([]list.Item{}, delegate, itemWidth+2, 0) // +2 for borders
	l.Title = "Your Library"
	l.Styles.TitleBar = lipgloss.NewStyle().
		Padding(0, 0, 1, 2).
		Width(itemWidth + 2)
	l.Styles.Title = lipgloss.NewStyle().
		Background(BorderColor).
		Foreground(WhiteTextColor).
		Padding(0, 1)

	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	return libraryModel{
		list:         l,
		spotifyState: spotifyState,
	}
}

func (m libraryModel) Init() tea.Cmd {
	return nil
}

func (m libraryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.list.SetHeight(m.height - 3)
		return m, nil

	case state.PlaylistsUpdatedMsg:
		m.list.SetItems(m.convertPlaylistsToItems())
		return m, m.spotifyState.SelectPlaylist(string(m.list.SelectedItem().(playlist).uri))

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Up, DefaultKeyMap.Down, DefaultKeyMap.Left, DefaultKeyMap.Right):
			m.list, cmd = m.list.Update(msg)
			return m, tea.Batch(cmd, m.spotifyState.SelectPlaylist(string(m.list.SelectedItem().(playlist).uri)))
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m libraryModel) View() string {
	// Update delegate styles based on focus
	delegate := list.NewDefaultDelegate()
	const itemWidth = 28

	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(WhiteTextColor).
		Padding(0, 0, 0, 2).
		Width(itemWidth).
		MaxWidth(itemWidth)

	delegate.Styles.NormalDesc = delegate.Styles.NormalTitle.
		Foreground(TextColor).
		Padding(0, 0, 0, 2).
		Width(itemWidth).
		MaxWidth(itemWidth)

	selectedTitleColor := WhiteTextColor
	if m.isFocused {
		selectedTitleColor = PrimaryColor
	}
	selectedDescColor := TextColor
	if m.isFocused {
		selectedDescColor = PrimaryColor
	}

	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(selectedTitleColor).
		Foreground(selectedTitleColor).
		Padding(0, 0, 0, 1).
		Bold(true).
		Width(itemWidth).
		MaxWidth(itemWidth)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(selectedDescColor).
		Padding(0, 0, 0, 2).
		Width(itemWidth).
		MaxWidth(itemWidth)

	m.list.SetDelegate(delegate)

	return lipgloss.NewStyle().
		Height(m.height - 2).
		MaxHeight(m.height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(getBorderStyle(m.isFocused)).
		Render(m.list.View())
}

func (m libraryModel) convertPlaylistsToItems() []list.Item {

	playlists := m.spotifyState.GetPlaylists()

	items := make([]list.Item, 0, len(playlists))
	for _, p := range playlists {
		title := p.Name
		if title == "" {
			title = "Untitled Playlist"
		}
		desc := p.Owner.DisplayName
		if desc == "" {
			desc = "Unknown Owner"
		}

		items = append(items, playlist{
			title: title,
			desc:  desc,
			uri:   string(p.URI),
		})
	}
	return items
}
