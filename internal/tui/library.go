package tui

import (
	"context"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zmb3/spotify/v2"
)

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
	height        int
	list          list.Model
	spotifyClient *spotify.Client
	err           error
}

// Message type for playlist updates
type playlistsUpdatedMsg struct {
	playlists []list.Item
	err       error
}

func newLibrary(client *spotify.Client) libraryModel {
	delegate := list.NewDefaultDelegate()

	// Set fixed width for titles and descriptions
	//potentially need to change this back to 28
	const itemWidth = 28

	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Padding(0, 0, 0, 2).
		Width(itemWidth).
		MaxWidth(itemWidth)

	//delegate.Styles.NormalDesc = lipgloss.NewStyle().
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
		list:          l,
		spotifyClient: client,
	}
}

// fetchPlaylists fetches playlists from Spotify and returns them as a command
func (m libraryModel) fetchPlaylists() tea.Cmd {
	return func() tea.Msg {
		log.Println("Fetching playlists...")
		if m.spotifyClient == nil {
			log.Println("Error: Spotify client is nil")
			return playlistsUpdatedMsg{err: fmt.Errorf("spotify client not initialized")}
		}

		playlists, err := m.spotifyClient.CurrentUsersPlaylists(context.Background())
		if err != nil {
			log.Printf("Error fetching playlists: %v", err)
			return playlistsUpdatedMsg{err: err}
		}

		if len(playlists.Playlists) == 0 {
			log.Println("No playlists found")
			return playlistsUpdatedMsg{playlists: []list.Item{}}
		}

		// Convert Spotify playlists to our playlist type
		items := make([]list.Item, 0, len(playlists.Playlists))
		for _, p := range playlists.Playlists {
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
			log.Printf("Added playlist: %s by %s", title, desc)
		}
		log.Printf("Successfully fetched %d playlists", len(items))
		return playlistsUpdatedMsg{playlists: items}
	}
}

func (m libraryModel) Init() tea.Cmd {
	// Start fetching playlists immediately
	log.Println("Initializing library model and fetching playlists...")

	return tea.Sequence(
		m.fetchPlaylists(),
		tea.WindowSize(), // This is a bi tof a hack tbf I dont understand why its needed
	)
}

func (m libraryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	log.Printf("Library model received message of type: %T", msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.list.SetHeight(m.height)
		return m, nil

	case playlistsUpdatedMsg:
		if msg.err != nil {
			m.err = msg.err
			log.Printf("Error updating playlists: %v", msg.err)
			return m, nil
		}

		if len(msg.playlists) == 0 {
			log.Println("Warning: Setting empty playlist list")
		} else {
			log.Printf("Setting %d playlists to list", len(msg.playlists))
			for i, item := range msg.playlists {
				if p, ok := item.(playlist); ok {
					log.Printf("  %d: %s by %s", i+1, p.title, p.desc)
				}
			}
		}

		m.list.SetItems(msg.playlists)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "enter" {
			// When enter is pressed, emit a playlist selected message
			if i := m.list.Index(); i != -1 {
				if p, ok := m.list.SelectedItem().(playlist); ok {
					log.Printf("Selected playlist: %s", p.uri)
					return m, func() tea.Msg {
						return playlistSelectedMsg{playlistID: p.uri}
					}
				}
			}
		}
	}

	// Handle list-specific updates
	m.list, cmd = m.list.Update(msg)
	log.Printf("List model was updated by message type: %T", msg)
	return m, cmd
}

func (m libraryModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error loading playlists: %v", m.err)
	}

	return m.list.View()
}
