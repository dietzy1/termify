package tui

import (
	"fmt"
	"log"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
	"github.com/evertras/bubble-table/table"
	"github.com/zmb3/spotify/v2"
)

// PlaylistViewModel represents the table view for playlists and tracks
type playlistViewModel struct {
	width, height int
	table         table.Model
	isFocused     bool
	spotifyState  *state.SpotifyState
}

// NewPlaylistView creates a new playlist view
func NewPlaylistView(spotifyState *state.SpotifyState) playlistViewModel {
	return playlistViewModel{
		table:        createPlaylistTable(),
		spotifyState: spotifyState,
	}
}

// Init initializes the playlist view
func (m playlistViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the playlist view
func (m playlistViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		log.Printf("PlaylistView width: %d, height: %d", m.width, m.height)

	case state.TracksUpdatedMsg:
		if msg.Err != nil {
			ShowError("Error loading tracks", msg.Err.Error())
			log.Printf("PlaylistView: Error loading tracks: %v", msg.Err)
			return m, nil
		}

		log.Printf("PlaylistView: Converting %d tracks to table rows", len(m.spotifyState.Tracks))
		m.updateTableWithTracks(m.spotifyState.Tracks)

	// Handle keyboard events for table navigation
	case tea.KeyMsg:
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
							log.Printf("PlaylistView: Selected track: %s", track.Track.Track.ID)
							return m, m.spotifyState.PlayTrack(track.Track.Track.ID)
						}
					}
				}
			}

			return m, nil

		case key.Matches(msg, DefaultKeyMap.AddToQueue):
			if selected := m.table.HighlightedRow(); selected.Data != nil {
				if numStr, ok := selected.Data["#"].(string); ok {
					if idx, err := strconv.Atoi(numStr); err == nil && idx > 0 && idx <= len(m.spotifyState.Tracks) {
						track := m.spotifyState.Tracks[idx-1]
						if track.Track.Track != nil {
							log.Printf("PlaylistView: Adding track to queue: %s", track.Track.Track.ID)
							return m, m.spotifyState.AddToQueue(track.Track.Track.ID)
						}
					}
				}
			}
		}
	}

	// Forward all other messages to the table
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the playlist view
func (m playlistViewModel) View() string {
	m.table = m.table.Border(RoundedTableBorders).
		HeaderStyle(
			lipgloss.NewStyle().
				BorderForeground(getBorderStyle(m.isFocused)))
	m.table = m.table.WithBaseStyle(
		lipgloss.NewStyle().BorderForeground(getBorderStyle(m.isFocused)),
	)

	currentPage := m.table.CurrentPage()
	maxPage := m.table.MaxPages()

	// loop through m.spotifyState.Playlists to find the selected playlist
	// then get the name of the selected playlist
	// then render the name of the selected playlist in the footer

	//TODO: this doesn't support the case where the selected playlist is not in the list of playlists and we are searching instead
	name := "Unknown Playlist"
	for _, playlist := range m.spotifyState.Playlists {
		if playlist.ID == spotify.ID(m.spotifyState.SelectedPlaylistID) {
			name = playlist.Name
			break
		}
	}

	styledName := lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor)).
		Padding(0, 1).
		Render(name)

	styledPage := lipgloss.NewStyle().
		Foreground(lipgloss.Color(WhiteTextColor)).
		Padding(0, 1).
		Render(fmt.Sprintf("| Page %d/%d", currentPage, maxPage))

	m.table = m.table.WithStaticFooter(
		styledName + styledPage,
	)

	return m.table.View()
}

// SetFocus sets the focus state of the playlist view
func (m *playlistViewModel) SetFocus(isFocused bool) {
	m.isFocused = isFocused
}

// Update table with tracks
func (m *playlistViewModel) updateTableWithTracks(tracks []spotify.PlaylistItem) {
	var rows []table.Row
	for i, track := range tracks {
		if track.Track.Track == nil {
			log.Printf("PlaylistView: Warning - Track %d is nil", i+1)
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

// createPlaylistTable creates a new table for the playlist view
func createPlaylistTable() table.Model {
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
