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

type queueViewModel struct {
	width, height int
	table         table.Model
	spotifyState  *state.SpotifyState
}

func NewQueueView(spotifyState *state.SpotifyState) queueViewModel {

	t := table.New([]table.Column{
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

	m := queueViewModel{
		table:        t,
		spotifyState: spotifyState,
	}

	return m
}

// Init initializes the queue view
func (m queueViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the queue view
func (m queueViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		m.table = m.table.WithTargetWidth(m.width).WithMinimumHeight(m.height).WithPageSize(m.height - 7)
		log.Printf("QueueView width: %d, height: %d", m.width, m.height)

	case state.QueueUpdatedMsg:
		queue := m.spotifyState.GetQueue()
		m.updateTableWithQueue(queue)

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
					queue := m.spotifyState.GetQueue()
					if idx, err := strconv.Atoi(numStr); err == nil && idx > 0 && idx <= len(queue) {
						track := queue[idx-1]
						log.Printf("QueueView: Selected track: %s", track.ID)
						return m, m.spotifyState.PlayTrack(track.ID)
					}
				}
			}
		}
	}

	// Forward all other messages to the table
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the queue view
func (m queueViewModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	var style lipgloss.Style
	/* if m.isFocused {
		style = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color(PrimaryColor))
	} else {
		style = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color(SecondaryColor))
	} */

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(PrimaryColor)).
		MarginLeft(2)

	title := titleStyle.Render("Queue")

	if m.spotifyState.IsQueueEmpty() {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextColor)).
			Align(lipgloss.Center).
			Width(m.width - 4)

		emptyText := emptyStyle.Render("Queue is empty")
		return style.Render(lipgloss.JoinVertical(lipgloss.Left, title, emptyText))
	}

	return style.Render(lipgloss.JoinVertical(lipgloss.Left, title, m.table.View()))
}

// updateTableWithQueue updates the table with the current queue
func (m *queueViewModel) updateTableWithQueue(queue []spotify.FullTrack) {
	if len(queue) == 0 {
		m.table = m.table.WithRows([]table.Row{})
		return
	}

	rows := make([]table.Row, 0, len(queue))

	for i, track := range queue {
		artist := ""
		if len(track.Artists) > 0 {
			artist = track.Artists[0].Name
		}

		album := track.Album.Name

		// Format duration from milliseconds to mm:ss
		durationSec := track.Duration / 1000
		minutes := durationSec / 60
		seconds := durationSec % 60
		duration := fmt.Sprintf("%d:%02d", minutes, seconds)

		row := table.NewRow(table.RowData{
			"#":        strconv.Itoa(i + 1),
			"Title":    track.Name,
			"Artist":   artist,
			"Album":    album,
			"Duration": duration,
		})

		rows = append(rows, row)
	}

	m.table = m.table.WithRows(rows)
}
