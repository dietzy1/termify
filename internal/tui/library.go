package tui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	height int
	list   list.Model
	err    error
}

func newLibrary() libraryModel {
	delegate := list.NewDefaultDelegate()

	const itemWidth = 28

	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Padding(0, 0, 0, 2).
		Width(itemWidth).
		MaxWidth(itemWidth)

	delegate.Styles.NormalDesc = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color(TextColor)).
		Padding(0, 0, 0, 2).
		Width(itemWidth).
		MaxWidth(itemWidth)

	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color(PrimaryColor)).
		Foreground(lipgloss.Color(PrimaryColor)).
		Padding(0, 0, 0, 1).
		Bold(true).
		Width(itemWidth).
		MaxWidth(itemWidth)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor)).
		Padding(0, 0, 0, 1).
		Width(itemWidth).
		MaxWidth(itemWidth)

	// Initialize empty list with fixed width
	l := list.New([]list.Item{}, delegate, itemWidth+2, 0) // +2 for borders
	l.Title = "Your Library"
	l.Styles.TitleBar = lipgloss.NewStyle().
		Padding(0, 0, 1, 2).
		Width(itemWidth + 2)
	l.Styles.Title = lipgloss.NewStyle().
		Background(lipgloss.Color(BorderColor)).
		Foreground(lipgloss.Color("#ffffff")).
		Padding(0, 1)

	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return libraryModel{
		list: l,
	}
}

func (m libraryModel) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m libraryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		log.Printf("Library: Updating window size to height: %d", msg.Height)
		m.height = msg.Height
		m.list.SetHeight(m.height)
		return m, nil

	case PlaylistsUpdatedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			log.Printf("Library: Error updating playlists: %v", msg.Err)
			return m, nil
		}
		log.Printf("Application: Converting %d playlists to list items", len(msg.Playlists))
		items := make([]list.Item, 0, len(msg.Playlists))
		for _, p := range msg.Playlists {
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

		m.list.SetItems(items)
		log.Println("Library: Successfully updated playlist items")
		return m, m.emitSelectedPlaylist()

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "down":
			m.list, cmd = m.list.Update(msg)
			return m, tea.Batch(cmd, m.emitSelectedPlaylist())
		}
	}

	// Handle list-specific updates
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m libraryModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error loading playlists: %v", m.err)
	}
	return m.list.View()
}

func (m libraryModel) emitSelectedPlaylist() tea.Cmd {
	if i := m.list.Index(); i != -1 {
		if p, ok := m.list.SelectedItem().(playlist); ok {
			log.Printf("Library: Selected playlist: %s", p.uri)
			return func() tea.Msg {
				return playlistSelectedMsg{playlistID: p.uri}
			}
		}
	}
	return nil
}
