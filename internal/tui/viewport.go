package tui

import (
	"fmt"
	"log"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
	"github.com/evertras/bubble-table/table"
	"github.com/zmb3/spotify/v2"
)

var _ tea.Model = (*viewportModel)(nil)

type viewportModel struct {
	width, height int
	table         table.Model
	isFocused     bool

	spotifyState *state.SpotifyState
}

// item implements list.Item interface for display in lists
type item struct {
	title string
	desc  string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func createTable() table.Model {

	return table.New([]table.Column{
		table.NewColumn("#", "#", 4).WithStyle(lipgloss.NewStyle().Align(lipgloss.Center)),
		table.NewFlexColumn("title", "Title", 1),   // Flex column with weight 3 4
		table.NewFlexColumn("artist", "Artist", 1), // Flex column with weight 2 4
		table.NewFlexColumn("album", "Album", 1),   // Flex column with weight 2 3
		table.NewColumn("duration", "Duration", 8).WithStyle(lipgloss.NewStyle().Align(lipgloss.Center)),
	}).WithRows([]table.Row{}).HeaderStyle(
		lipgloss.NewStyle().
			Bold(true).
			BorderForeground(lipgloss.Color(BorderColor)).
			Underline(true),
	).WithBaseStyle(
		lipgloss.NewStyle().
			Align(lipgloss.Left).
			BorderForeground(lipgloss.Color(BorderColor)),
	).Focused(true).HighlightStyle(
		lipgloss.NewStyle().
			Background(lipgloss.Color(BackgroundColor)).
			Foreground(lipgloss.Color(PrimaryColor)).
			Padding(0, 0, 0, 1).Bold(true),
	).Border(
		RoundedTableBorders,
	)
}

func newViewport(spotifyState *state.SpotifyState) viewportModel {
	return viewportModel{
		table:        createTable(),
		spotifyState: spotifyState,
	}
}

func (m viewportModel) Init() tea.Cmd {
	return nil
}

func (m viewportModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

		// This check is a panic safeguard
		if m.height-7 < 0 {
			m.table = m.table.WithTargetWidth(m.width).WithMinimumHeight(m.height).WithPageSize(1)
			return m, nil
		}

		m.table = m.table.WithTargetWidth(m.width).WithMinimumHeight(m.height).WithPageSize(m.height - 9)
		log.Printf("Viewport width: %d, height: %d", m.width, m.height)

	case state.TracksUpdatedMsg:
		if msg.Err != nil {
			log.Printf("Viewport: Error loading tracks: %v", msg.Err)
			return m, nil
		}

		log.Printf("Viewport: Converting %d tracks to table rows", len(m.spotifyState.Tracks))
		m.updateTableWithTracks(m.spotifyState.Tracks)

	// Handle keyboard events for table navigation and search
	case tea.KeyMsg:

		// Handle regular navigation when not in search mode
		switch {
		case key.Matches(msg, DefaultKeyMap.Up, DefaultKeyMap.Down):
			var tableCmd tea.Cmd
			m.table, tableCmd = m.table.Update(msg)
			return m, tableCmd

		case key.Matches(msg, DefaultKeyMap.Select):
			if selected := m.table.HighlightedRow(); selected.Data != nil {
				if numStr, ok := selected.Data["#"].(string); ok {
					if idx, err := strconv.Atoi(numStr); err == nil && idx > 0 && idx <= len(m.spotifyState.Tracks) {
						track := m.spotifyState.Tracks[idx-1]
						if track.Track.Track != nil {
							log.Printf("Viewport: Selected track: %s", track.Track.Track.ID)
							return m, m.spotifyState.PlayTrack(track.Track.Track.ID)
						}
					}
				}
			}
			return m, nil
		}
	}

	// Forward all other messages to the table
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// Viewport table schema:
// Tracks
// Albums
// Artists
// Playlists
func createLists(width, height int) string {
	listWidth := (width / 2)
	upperListsHeight := (height / 2) - 1
	lowerListsHeight := height - upperListsHeight - 4

	// Mock data for lists
	trackItems := []list.Item{
		item{title: "Bohemian Rhapsody and very long title to test", desc: "Queen"},
		item{title: "Imagine", desc: "John Lennon"},
		item{title: "Billie Jean", desc: "Michael Jackson"},
	}

	playlistItems := []list.Item{
		item{title: "My Mix", desc: "Custom playlist"},
		item{title: "Workout Mix", desc: "Energetic songs"},
		item{title: "Chill Vibes", desc: "Relaxing tunes"},
	}

	albumItems := []list.Item{
		item{title: "A Night at the Opera", desc: "Queen"},
		item{title: "Thriller", desc: "Michael Jackson"},
		item{title: "Abbey Road", desc: "The Beatles"},
	}

	artistItems := []list.Item{
		item{title: "Queen", desc: "Rock band"},
		item{title: "Michael Jackson", desc: "Pop artist"},
		item{title: "The Beatles", desc: "Rock band"},
	}

	// Create delegates
	delegate := list.NewDefaultDelegate()

	var itemWidth = listWidth - 2

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
		Padding(0, 0, 0, 2).
		Width(itemWidth).
		MaxWidth(itemWidth)

	// Create lists
	trackList := list.New(trackItems, delegate, listWidth, upperListsHeight)
	trackList.Title = "Tracks"
	trackList.SetShowStatusBar(false)
	trackList.SetFilteringEnabled(false)
	trackList.SetShowHelp(false)
	trackList.DisableQuitKeybindings()

	playlistList := list.New(playlistItems, delegate, listWidth, upperListsHeight)
	playlistList.Title = "Playlists"
	playlistList.SetShowStatusBar(false)
	playlistList.SetFilteringEnabled(false)
	playlistList.SetShowHelp(false)
	playlistList.DisableQuitKeybindings()

	albumList := list.New(albumItems, delegate, listWidth, lowerListsHeight)
	albumList.Title = "Albums"
	albumList.SetShowStatusBar(false)
	albumList.SetFilteringEnabled(false)
	albumList.SetShowHelp(false)
	albumList.DisableQuitKeybindings()

	artistList := list.New(artistItems, delegate, listWidth, lowerListsHeight)
	artistList.Title = "Artists"
	artistList.SetShowStatusBar(false)
	artistList.SetFilteringEnabled(false)
	artistList.SetShowHelp(false)
	artistList.DisableQuitKeybindings()

	// Style the lists
	listStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(BorderColor)).
		Padding(0, 0)

	// Render all lists
	topRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		listStyle.Render(trackList.View()),
		listStyle.Render(playlistList.View()),
	)

	bottomRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		listStyle.Render(artistList.View()),
		listStyle.Render(albumList.View()),
	)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)
}

func (m viewportModel) View() string {
	// Determine which view to show based on search mode
	var contentView string
	if false {
		// Show lists when in search mode
		contentView = createLists(m.width, m.height-3) // Subtract height for search bar
	} else {
		// Show table when not in search mode
		m.table = m.table.Border(RoundedTableBorders).
			HeaderStyle(
				lipgloss.NewStyle().
					BorderForeground(getBorderStyle(m.isFocused)))
		m.table = m.table.WithBaseStyle(
			lipgloss.NewStyle().BorderForeground(getBorderStyle(m.isFocused)),
		)

		contentView = m.table.View()
	}

	// Combine search bar with content
	return lipgloss.JoinVertical(lipgloss.Left, contentView)
}

// Update table with tracks
func (m *viewportModel) updateTableWithTracks(tracks []spotify.PlaylistItem) {
	var rows []table.Row
	for i, track := range tracks {
		if track.Track.Track == nil {
			log.Printf("Viewport: Warning - Track %d is nil", i+1)
			continue
		}

		// Get primary artist name
		artistName := "Unknown Artist"
		if len(track.Track.Track.Artists) > 0 {
			artistName = track.Track.Track.Artists[0].Name
		}

		// Get album name
		albumName := "Unknown Album"
		if track.Track.Track.Album.Name != "" {
			albumName = track.Track.Track.Album.Name
		}

		// Format duration
		duration := formatTrackDuration(int(track.Track.Track.Duration))

		rows = append(rows, table.NewRow(table.RowData{
			"#":        fmt.Sprintf("%d", i+1),
			"title":    track.Track.Track.Name,
			"artist":   artistName,
			"album":    albumName,
			"duration": duration,
		}))
	}

	m.table = m.table.WithRows(rows)
}
