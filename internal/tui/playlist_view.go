package tui

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
	"github.com/evertras/bubble-table/table"
	"github.com/zmb3/spotify/v2"
)

type clearQueuedHighlightMsg struct {
	TrackID spotify.ID
}

// PlaylistViewModel represents the table view for playlists and tracks
type playlistViewModel struct {
	ctx            context.Context
	width, height  int
	table          table.Model
	isFocused      bool
	spotifyState   *state.SpotifyState
	queuedTracks   map[spotify.ID]bool // Track which songs have been queued recently
	highlightTimer *time.Timer         // Timer to clear the highlight
}

func newPlaylistView(ctx context.Context, spotifyState *state.SpotifyState) playlistViewModel {
	return playlistViewModel{
		ctx:          ctx,
		table:        createPlaylistTable(),
		spotifyState: spotifyState,
		queuedTracks: make(map[spotify.ID]bool),
	}
}

func (m playlistViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the playlist view
func (m playlistViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		const minTableHeight = 7
		// This check is a panic safeguard
		if m.height-minTableHeight < 0 {
			m.table = m.table.WithTargetWidth(m.width).WithMinimumHeight(m.height).WithPageSize(1)
			return m, nil
		}

		m.table = m.table.WithTargetWidth(m.width).WithMinimumHeight(m.height).WithPageSize(m.height - 6)
		log.Printf("PlaylistView width: %d, height: %d", m.width, m.height)

	case state.TracksUpdatedMsg:
		tracks := m.spotifyState.GetTracks()
		m.updateTableWithTracks(tracks)

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
					tracks := m.spotifyState.GetTracks()
					if idx, err := strconv.Atoi(numStr); err == nil && idx > 0 && idx <= len(tracks) {
						track := tracks[idx-1]
						log.Printf("PlaylistView: Selected track: %s", track.ID)
						return m, m.spotifyState.PlayTrack(m.ctx, track.ID)
					}
				}
			}

			return m, nil

		case key.Matches(msg, DefaultKeyMap.AddToQueue):
			if selected := m.table.HighlightedRow(); selected.Data != nil {
				if numStr, ok := selected.Data["#"].(string); ok {
					tracks := m.spotifyState.GetTracks()
					if idx, err := strconv.Atoi(numStr); err == nil && idx > 0 && idx <= len(tracks) {
						track := tracks[idx-1]
						log.Printf("PlaylistView: Adding track to queue: %s", track.ID)

						m.queuedTracks[track.ID] = true
						m.updateTableWithTracks(tracks)
						m.spotifyState.Queue.Enqueue(track)
						return m, tea.Batch(
							state.UpdateQueue(),
							tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
								return clearQueuedHighlightMsg{TrackID: track.ID}
							}),
						)
					}
				}
			}
		}
	case clearQueuedHighlightMsg:
		delete(m.queuedTracks, msg.TrackID)
		tracks := m.spotifyState.GetTracks()
		m.updateTableWithTracks(tracks)
		return m, nil
	}

	// Forward all other messages to the table
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

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
	playlists := m.spotifyState.GetPlaylists()
	selectedId := m.spotifyState.GetSelectedID()

	name := "Unknown Playlist"
	for _, playlist := range playlists {
		if playlist.ID == spotify.ID(selectedId) {
			name = playlist.Name
			break
		}
	}
	//As fallback check searchResults for the selected playlist and brute force it
	if name == "Unknown Playlist" {
		playlistSearchResults := m.spotifyState.GetSearchResultPlaylists()
		for _, playlist := range playlistSearchResults {
			if playlist.ID == spotify.ID(selectedId) {
				name = playlist.Name
				break
			}
		}

		artistsSearchResults := m.spotifyState.GetSearchResultArtists()
		for _, artist := range artistsSearchResults {
			if artist.ID == spotify.ID(selectedId) {
				name = artist.Name
				break
			}
		}
		albumsSearchResults := m.spotifyState.GetSearchResultAlbums()
		for _, album := range albumsSearchResults {
			if album.ID == spotify.ID(selectedId) {
				name = album.Name
				break
			}
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

func (m *playlistViewModel) SetFocus(isFocused bool) {
	m.isFocused = isFocused
}

func (m *playlistViewModel) updateTableWithTracks(tracks []spotify.SimpleTrack) {
	var rows []table.Row
	for i, track := range tracks {
		// Get primary artist name
		artistName := "Unknown Artist"
		if len(track.Artists) > 0 {
			artistName = track.Artists[0].Name
		}

		// Get album name
		albumName := "Unknown Album"
		if track.Album.Name != "" {
			albumName = track.Album.Name
		}

		// Format duration
		duration := formatTrackDuration(int(track.Duration))

		// Check if this track was recently queued, if so add the "QUEUED" highlight to the title
		title := track.Name
		if m.queuedTracks[track.ID] {
			// Create a highlighted title that shows it's been queued
			queuedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(PrimaryColor)).
				Bold(true)

			title = title + " " + queuedStyle.Render("(Added to queue)")
		}

		rows = append(rows, table.NewRow(table.RowData{
			"#":        fmt.Sprintf("%d", i+1),
			"title":    title,
			"artist":   artistName,
			"album":    albumName,
			"duration": duration,
		}))
	}

	m.table = m.table.WithRows(rows)
}

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

// GetNextTrack returns the ID of the next track to play when autoplay is triggered
// It returns the next track after the currently playing one, or the first track if none is playing
func (m *playlistViewModel) getNextTrack() spotify.ID {

	tracks := m.spotifyState.GetTracks()
	if len(tracks) == 0 {
		log.Println("No tracks in playlist to autoplay")
		return ""
	}

	// If no track is currently playing, return the first track
	playerState := m.spotifyState.GetPlayerState()

	if playerState.Item == nil {
		log.Println("No track currently playing, returning first track")
		return spotify.ID(tracks[0].ID)
	}

	currentTrackID := playerState.Item.ID
	log.Printf("Current track ID: %s", currentTrackID)

	// Find the current track in the playlist
	for i, track := range tracks {
		if string(track.ID) == string(currentTrackID) {
			// If it's the last track, return empty (we'll use recommendations)
			if i >= len(tracks)-1 {
				log.Println("Current track is the last one in playlist")
				return ""
			}
			// Return the next track
			log.Printf("Found current track at index %d, returning next track", i)
			return spotify.ID(tracks[i+1].ID)
		}
	}

	// If current track not found in playlist, return the first track
	log.Println("Current track not found in playlist, returning first track")
	return spotify.ID(tracks[0].ID)
}
