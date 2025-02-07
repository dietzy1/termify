package tui

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

var _ tea.Model = (*viewportModel)(nil)

type viewportModel struct {
	width, height int
	table         table.Model
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
		//Background(lipgloss.Color(PrimaryColor)),
	)
}

func newViewport() viewportModel {
	return viewportModel{
		table: createTable(),
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
		m.table = m.table.WithTargetWidth(m.width).WithMinimumHeight(m.height)
		log.Printf("Viewport width: %d, height: %d", m.width, m.height)

	case tracksLoadedMsg:
		if msg.err != nil {
			log.Printf("Error loading tracks into viewport: %v", msg.err)
			return m, nil
		}

		// Convert tracks to table rows
		var rows []table.Row
		for _, track := range msg.tracks {
			// Get primary artist name
			artistName := "Unknown Artist"
			if len(track.Track.Artists) > 0 {
				artistName = track.Track.Artists[0].Name
			}

			// Get album name
			albumName := "Unknown Album"
			if track.Track.Album.Name != "" {
				albumName = track.Track.Album.Name
			}

			// Format duration
			duration := formatTrackDuration(int(track.Track.Duration))

			rows = append(rows, table.NewRow(table.RowData{
				"title":    track.Track.Name,
				"artist":   artistName,
				"album":    albumName,
				"duration": duration,
			}))
		}

		m.table = m.table.WithRows(rows)
		return m, nil

	// Handle keyboard events for table navigation
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "down", "enter", "esc":
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
