package tui

import (
	"fmt"
	"log"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
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
	allTracks     []spotify.PlaylistItem // Store all tracks for filtering
	searchInput   textinput.Model        // Text input for search
	searching     bool                   // Whether we're in search mode

	spotifyState *SpotifyState
}

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
	ti := textinput.New()
	ti.Placeholder = "Search tracks..."
	ti.CharLimit = 50
	ti.Width = 30 // Will be adjusted based on window width

	return viewportModel{
		table:        createTable(),
		tracks:       make([]spotify.PlaylistItem, 0),
		allTracks:    make([]spotify.PlaylistItem, 0),
		searchInput:  ti,
		searching:    false,
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

		// Adjust search input width
		m.searchInput.Width = m.width - 2

		// This check is a panic safeguard
		if m.height-7 < 0 { // One more line for search bar
			m.table = m.table.WithTargetWidth(m.width).WithMinimumHeight(m.height).WithPageSize(1)
			return m, nil
		}

		m.table = m.table.WithTargetWidth(m.width).WithMinimumHeight(m.height - 1 - 2).WithPageSize(m.height - 7)
		log.Printf("Viewport width: %d, height: %d", m.width, m.height)

	case TracksUpdatedMsg:
		if msg.Err != nil {
			log.Printf("Viewport: Error loading tracks: %v", msg.Err)
			return m, nil
		}

		log.Printf("Viewport: Converting %d tracks to table rows", len(msg.Tracks))
		m.allTracks = msg.Tracks // Store all tracks for filtering
		m.tracks = msg.Tracks    // Store tracks for selection
		m.updateTableWithTracks(m.tracks)

	// Handle keyboard events for table navigation and search
	case tea.KeyMsg:
		// Handle search mode toggle
		if key.Matches(msg, key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search"))) {
			m.searching = !m.searching
			if m.searching {
				m.searchInput.Focus()
				return m, textinput.Blink
			} else {
				m.searchInput.Blur()
			}
			return m, nil
		}

		// Handle search input when in search mode
		if m.searching {
			switch msg.String() {
			case "esc":
				m.searching = false
				m.searchInput.Blur()
				return m, nil
			case "enter":
				m.searching = false
				m.searchInput.Blur()
				return m, nil
			default:
				var inputCmd tea.Cmd
				m.searchInput, inputCmd = m.searchInput.Update(msg)
				return m, inputCmd
			}
		}

		// Handle regular navigation when not in search mode
		switch {
		case key.Matches(msg, DefaultKeyMap.Up, DefaultKeyMap.Down):
			var tableCmd tea.Cmd
			m.table, tableCmd = m.table.Update(msg)
			return m, tableCmd

		case key.Matches(msg, DefaultKeyMap.Select):
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
		}
	}

	// Forward all other messages to the table
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m viewportModel) View() string {
	// Create search bar style
	searchStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(BorderColor)).
		Padding(0, 1).
		Width(m.width - 2)

	// Search bar with indicator
	searchPrefix := "🔍 "
	if !m.searching {
		searchPrefix = "/ "
	}

	searchBar := searchStyle.Render(searchPrefix + m.searchInput.View())

	// Combine search bar with table
	return lipgloss.JoinVertical(
		lipgloss.Left,
		searchBar,
		m.table.View(),
	)
}

// Helper function to format track duration
func formatTrackDuration(ms int) string {
	seconds := ms / 1000
	minutes := seconds / 60
	remainingSeconds := seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, remainingSeconds)
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
