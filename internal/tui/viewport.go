package tui

import (
	"fmt"
	"log"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/zmb3/spotify/v2"
)

var _ tea.Model = (*viewportModel)(nil)

type viewportModel struct {
	width, height int
	table         table.Model
	tracks        []spotify.PlaylistItem // Store tracks for selection

	spotifyState *SpotifyState
}

// "https://api.spotify.com/v1/playlists/3bNsJuQ0M60iaK7fuwyKwS/tracks"

// Game plan is that we are going to request everything in a single call and hold that information in memory.
// We are then going to swap out the data found in the table based on the user's navigation.

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
			BorderForeground(lipgloss.Color(BackgroundColor)).
			Underline(true),
	).WithBaseStyle(
		lipgloss.NewStyle().
			Align(lipgloss.Left).
			BorderForeground(lipgloss.Color(BackgroundColor)),
	).Focused(true).HighlightStyle(
		lipgloss.NewStyle().
			Background(lipgloss.Color(BackgroundColor)).
			Foreground(lipgloss.Color(PrimaryColor)).
			Padding(0, 0, 0, 1).Bold(true),
	)
}

func newViewport(spotifyState *SpotifyState) viewportModel {
	return viewportModel{
		table:        createTable(),
		tracks:       make([]spotify.PlaylistItem, 0),
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
		m.table = m.table.WithTargetWidth(m.width).WithMinimumHeight(m.height).WithPageSize(m.height - 6)
		log.Printf("Viewport width: %d, height: %d", m.width, m.height)

	case TracksUpdatedMsg:
		if msg.Err != nil {
			log.Printf("Viewport: Error loading tracks: %v", msg.Err)
			return m, nil
		}

		log.Printf("Viewport: Converting %d tracks to table rows", len(msg.Tracks))
		var rows []table.Row
		for i, track := range msg.Tracks {
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

			m.tracks = msg.Tracks // Store tracks for selection
			m.table = m.table.WithRows(rows)

		}

	// Handle keyboard events for table navigation
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// When enter is pressed, get the selected track and emit a selection message
			if selected := m.table.HighlightedRow(); selected.Data != nil {
				if numStr, ok := selected.Data["#"].(string); ok {
					if idx, err := strconv.Atoi(numStr); err == nil && idx > 0 && idx <= len(m.tracks) {
						track := m.tracks[idx-1]
						if track.Track.Track != nil {
							log.Printf("Viewport: Selected track: %s", track.Track.Track.ID)
							return m, m.spotifyState.PlayTrack(track.Track.Track.ID)
						}
					}
				}
			}
			return m, nil
		case "up", "down":
			var tableCmd tea.Cmd
			m.table, tableCmd = m.table.Update(msg)
			return m, tableCmd
		}
	}

	// Forward all other messages to the table
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m viewportModel) View() string {
	return m.table.View()
}

// Helper function to format track duration
func formatTrackDuration(ms int) string {
	seconds := ms / 1000
	minutes := seconds / 60
	remainingSeconds := seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, remainingSeconds)
}
