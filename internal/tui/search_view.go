package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
)

// item implements list.Item interface for display in lists
type item struct {
	title string
	desc  string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// SearchViewModel represents the search view with multiple lists
type searchViewModel struct {
	width, height int
	isFocused     bool
	spotifyState  *state.SpotifyState

	// Lists for different content types
	trackList    list.Model
	playlistList list.Model
	albumList    list.Model
	artistList   list.Model
}

// NewSearchView creates a new search view
func NewSearchView(spotifyState *state.SpotifyState) searchViewModel {
	m := searchViewModel{
		spotifyState: spotifyState,
		trackList:    createEmptyList("Tracks"),
		playlistList: createEmptyList("Playlists"),
		albumList:    createEmptyList("Albums"),
		artistList:   createEmptyList("Artists"),
	}

	return m
}

// Init initializes the search view
func (m searchViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the search view
func (m searchViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		log.Println("Search view received window size message: ", m.width, m.height)

		// Recalculate list dimensions
		listWidth := (m.width / 2)
		upperListsHeight := (m.height / 2) - 1
		lowerListsHeight := m.height - upperListsHeight - 4

		// Update list dimensions
		m.trackList.SetSize(listWidth, upperListsHeight)
		m.playlistList.SetSize(listWidth, upperListsHeight)
		m.albumList.SetSize(listWidth, lowerListsHeight)
		m.artistList.SetSize(listWidth, lowerListsHeight)

		m.trackList.Styles.NoItems = lipgloss.NewStyle().Width(listWidth-2).Padding(0, 0, 0, 2)
		m.playlistList.Styles.NoItems = lipgloss.NewStyle().Width(listWidth-2).Padding(0, 0, 0, 2)
		m.albumList.Styles.NoItems = lipgloss.NewStyle().Width(listWidth-2).Padding(0, 0, 0, 2)
		m.artistList.Styles.NoItems = lipgloss.NewStyle().Width(listWidth-2).Padding(0, 0, 0, 2)

		m.updateListStyles(listWidth)

	case tea.KeyMsg:

		// For now, we'll just forward to all lists
		// TODO: We need to implement focus management for the 4 lists
		var cmd tea.Cmd
		m.trackList, cmd = m.trackList.Update(msg)
		cmds = append(cmds, cmd)

		m.playlistList, cmd = m.playlistList.Update(msg)
		cmds = append(cmds, cmd)

		m.albumList, cmd = m.albumList.Update(msg)
		cmds = append(cmds, cmd)

		m.artistList, cmd = m.artistList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the search view
func (m searchViewModel) View() string {
	// Style the lists
	listStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(getBorderStyle(m.isFocused)).
		Padding(0, 0)

		// Render all lists
	topRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		listStyle.Render(m.trackList.View()),
		listStyle.Render(m.playlistList.View()),
	)

	bottomRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		listStyle.Render(m.artistList.View()),
		listStyle.Render(m.albumList.View()),
	)

	// Render all lists
	/* topRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		listStyle.Render(m.renderListWithFullWidth(m.trackList, m.width/2)),
		listStyle.Render(m.renderListWithFullWidth(m.playlistList, m.width/2)),
	)

	bottomRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		listStyle.Render(m.renderListWithFullWidth(m.artistList, m.width/2)),
		listStyle.Render(m.renderListWithFullWidth(m.albumList, m.width/2)),
	) */

	return lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)
}

// renderListWithFullWidth ensures the list takes full width even when empty
/* func (m searchViewModel) renderListWithFullWidth(l list.Model, width int) string {
	// Get the list's view
	view := l.View()

	// Check if the list is empty (contains "No items.")
	if strings.Contains(view, "No items.") {
		// Calculate the inner width (accounting for borders)
		innerWidth := width - 4 // Subtract 4 for the left and right borders (2 each side)

		// Create a style for the empty state that takes full width
		emptyStyle := lipgloss.NewStyle().
			Width(innerWidth).
			Align(lipgloss.Left)

		// Split the view to get the title and empty message
		lines := strings.Split(view, "\n")

		// If there's a title and empty message
		if len(lines) >= 3 {
			// Reconstruct the view with a full-width empty message
			title := lines[0]
			emptyLine := emptyStyle.Render("No items.")

			// Calculate padding to fill the height
			padding := ""
			for i := 0; i < l.Height()-3; i++ {
				padding += strings.Repeat(" ", innerWidth) + "\n"
			}

			return title + "\n" + emptyLine + "\n" + padding
		}
	}

	return view
} */

// SetFocus sets the focus state of the search view
func (m *searchViewModel) SetFocus(isFocused bool) {
	m.isFocused = isFocused
}

// UpdateSearchResults updates the lists with search results
func (m *searchViewModel) UpdateSearchResults(query string) {
	// This would be implemented to fetch and display actual search results
	// For now, we'll use mock data

	// Mock data for lists
	trackItems := []list.Item{
		item{title: "Bohemian Rhapsody", desc: "Queen"},
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

	// Update the lists with the mock data
	m.trackList.SetItems(trackItems)
	m.playlistList.SetItems(playlistItems)
	m.albumList.SetItems(albumItems)
	m.artistList.SetItems(artistItems)
}

// Helper functions

// createEmptyList creates a new empty list with the given title
func createEmptyList(title string) list.Model {
	delegate := list.NewDefaultDelegate()
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	return l
}

// updateListStyles updates the styles of all lists
func (m *searchViewModel) updateListStyles(itemWidth int) {
	// Create a delegate with the correct styles
	delegate := list.NewDefaultDelegate()

	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Padding(0, 0, 0, 2).
		Width(itemWidth - 2).
		MaxWidth(itemWidth - 2)

	delegate.Styles.NormalDesc = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color(TextColor)).
		Padding(0, 0, 0, 2).
		Width(itemWidth - 2).
		MaxWidth(itemWidth - 2)

	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color(PrimaryColor)).
		Foreground(lipgloss.Color(PrimaryColor)).
		Padding(0, 0, 0, 1).
		Bold(true).
		Width(itemWidth - 2).
		MaxWidth(itemWidth - 2)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor)).
		Padding(0, 0, 0, 2).
		Width(itemWidth - 2).
		MaxWidth(itemWidth - 2)

	// Update the delegates for all lists
	m.trackList.SetDelegate(delegate)
	m.playlistList.SetDelegate(delegate)
	m.albumList.SetDelegate(delegate)
	m.artistList.SetDelegate(delegate)
}
