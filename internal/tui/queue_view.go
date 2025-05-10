package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
)

// TODO: Requirements for the queue:
// Show current song playing
// Show next in line songs in the queue
// Have a clear button // Requires queue redesign since no API support
// Ability to play next songs by pressing select

var _ tea.Model = (*queueModel)(nil)

type queueItem struct {
	title string
	desc  string
	uri   string
}

func (p queueItem) Title() string       { return p.title }
func (p queueItem) Description() string { return p.desc }
func (p queueItem) FilterValue() string { return p.title }

type queueModel struct {
	height       int
	list         list.Model
	spotifyState *state.SpotifyState
	isFocused    bool
}

func newQueue(spotifyState *state.SpotifyState) queueModel {
	delegate := list.NewDefaultDelegate()

	const itemWidth = 28

	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(WhiteTextColor).
		Padding(0, 0, 0, 2).
		Width(itemWidth).
		MaxWidth(itemWidth)

	delegate.Styles.NormalDesc = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color(TextColor)).
		Padding(0, 0, 0, 2).
		Width(itemWidth).
		MaxWidth(itemWidth)

	l := list.New([]list.Item{}, delegate, itemWidth+2, 0) // +2 for borders
	l.Title = "Next in Queue"
	l.Styles.TitleBar = lipgloss.NewStyle().
		Padding(0, 0, 1, 2).
		Width(itemWidth + 2)
	l.Styles.Title = lipgloss.NewStyle().
		Background(lipgloss.Color(BorderColor)).
		Foreground(lipgloss.Color(WhiteTextColor)).
		Padding(0, 1)

	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	return queueModel{
		list:         l,
		spotifyState: spotifyState,
		isFocused:    false,
	}
}

func (m queueModel) Init() tea.Cmd {
	return nil
}

func (m queueModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.list.SetHeight(m.height - 3)
		return m, nil

	case state.QueueUpdatedMsg:
		m.list.SetItems(m.convertQueueToItems())

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Up, DefaultKeyMap.Down):

		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m queueModel) View() string {
	// Update delegate styles based on focus
	delegate := list.NewDefaultDelegate()
	const itemWidth = 28

	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(WhiteTextColor).
		Padding(0, 0, 0, 2).
		Width(itemWidth).
		MaxWidth(itemWidth)

	delegate.Styles.NormalDesc = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color(TextColor)).
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
		BorderForeground(lipgloss.Color(selectedTitleColor)).
		Foreground(lipgloss.Color(selectedTitleColor)).
		Padding(0, 0, 0, 1).
		Bold(true).
		Width(itemWidth).
		MaxWidth(itemWidth)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color(selectedDescColor)).
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

func (m queueModel) convertQueueToItems() []list.Item {

	queueItems := m.spotifyState.GetQueue()
	log.Println("Queue items found in queue model: ", queueItems)

	items := make([]list.Item, 0, len(queueItems))
	for _, item := range queueItems {
		title := item.Name
		if title == "" {
			title = "Untitled Playlist"
		}
		desc := item.SimpleTrack.Artists[0].Name
		if desc == "" {
			desc = "Unknown Owner"
		}

		items = append(items, queueItem{
			title: title,
			desc:  desc,
			uri:   string(item.URI),
		})
	}
	return items
}
