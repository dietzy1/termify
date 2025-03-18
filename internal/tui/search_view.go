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

	// Track which list is currently active
	activeList FocusedModel
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

	case state.SearchResultsUpdatedMsg:
		// When search results are updated, update the view
		if msg.Err == nil {
			// Update the search results in the view
			m.UpdateSearchResults()
		} else {
			log.Printf("Search view received error in search results: %v", msg.Err)
		}

	case tea.KeyMsg:
		// Handle key messages based on which list is active
		var cmd tea.Cmd
		// Update only the active list based on the current focus
		switch m.activeList {
		case FocusSearchTracksView:
			m.trackList, cmd = m.trackList.Update(msg)
			cmds = append(cmds, cmd)
		case FocusSearchPlaylistsView:
			m.playlistList, cmd = m.playlistList.Update(msg)
			cmds = append(cmds, cmd)
		case FocusSearchArtistsView:
			m.artistList, cmd = m.artistList.Update(msg)
			cmds = append(cmds, cmd)
		case FocusSearchAlbumsView:
			m.albumList, cmd = m.albumList.Update(msg)
			cmds = append(cmds, cmd)
		default:
			// If no specific list is focused, update all lists
			m.trackList, cmd = m.trackList.Update(msg)
			cmds = append(cmds, cmd)

			m.playlistList, cmd = m.playlistList.Update(msg)
			cmds = append(cmds, cmd)

			m.albumList, cmd = m.albumList.Update(msg)
			cmds = append(cmds, cmd)

			m.artistList, cmd = m.artistList.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the search view
func (m searchViewModel) View() string {
	// Base style for all lists
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(BorderColor)).
		Padding(0, 0)

	// Style for the focused list
	focusedStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(PrimaryColor)).
		Padding(0, 0)

	// Determine which style to use for each list
	tracksStyle := baseStyle
	if m.activeList == FocusSearchTracksView {
		tracksStyle = focusedStyle
	}

	playlistsStyle := baseStyle
	if m.activeList == FocusSearchPlaylistsView {
		playlistsStyle = focusedStyle
	}

	artistsStyle := baseStyle
	if m.activeList == FocusSearchArtistsView {
		artistsStyle = focusedStyle
	}

	albumsStyle := baseStyle
	if m.activeList == FocusSearchAlbumsView {
		albumsStyle = focusedStyle
	}

	// Render all lists with appropriate styles
	topRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		tracksStyle.Render(m.trackList.View()),
		playlistsStyle.Render(m.playlistList.View()),
	)

	bottomRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		artistsStyle.Render(m.artistList.View()),
		albumsStyle.Render(m.albumList.View()),
	)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)
}

// SetFocus sets the focus state of the search view
func (m *searchViewModel) SetFocus(isFocused bool) {
	m.isFocused = isFocused
}

// UpdateSearchResults updates the lists with search results
func (m *searchViewModel) UpdateSearchResults() {
	// Create empty slices for each type of item
	var trackItems []list.Item
	var playlistItems []list.Item
	var albumItems []list.Item
	var artistItems []list.Item

	// Process track results
	for _, track := range m.spotifyState.SearchResults.Tracks {
		artistName := ""
		if len(track.Artists) > 0 {
			artistName = track.Artists[0].Name
		}
		trackItems = append(trackItems, item{
			title: track.Name,
			desc:  artistName,
		})
	}

	// Process playlist results
	for _, playlist := range m.spotifyState.SearchResults.Playlists {

		ownerName := ""
		if playlist.Owner.DisplayName != "" {
			ownerName = playlist.Owner.DisplayName
		}

		playlistItems = append(playlistItems, item{
			title: playlist.Name,
			desc:  ownerName,
		})
	}

	// Process album results
	for _, album := range m.spotifyState.SearchResults.Albums {
		artistName := ""
		if len(album.Artists) > 0 {
			artistName = album.Artists[0].Name
		}
		albumItems = append(albumItems, item{
			title: album.Name,
			desc:  artistName,
		})
	}

	// Process artist results
	for _, artist := range m.spotifyState.SearchResults.Artists {
		// For artists, we don't have a specific description field in the API
		// We could potentially use genres or popularity as a description
		desc := "Artist"
		if len(artist.Genres) > 0 {
			desc = artist.Genres[0]
		}
		artistItems = append(artistItems, item{
			title: artist.Name,
			desc:  desc,
		})
	}

	m.trackList.SetItems(trackItems)
	m.playlistList.SetItems(playlistItems)
	m.albumList.SetItems(albumItems)
	m.artistList.SetItems(artistItems)
}

func createEmptyList(title string) list.Model {
	delegate := list.NewDefaultDelegate()
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	l.Styles.Title = lipgloss.NewStyle().
		Background(lipgloss.Color(BorderColor)).
		Foreground(lipgloss.Color("#ffffff")).
		Padding(0, 1)

	return l
}

func (m *searchViewModel) updateListStyles(itemWidth int) {
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

	// Create a function to get the appropriate selected style based on focus
	getSelectedStyle := func(isFocused bool) lipgloss.Style {
		selectedColor := TextColor
		if isFocused {
			selectedColor = PrimaryColor
		}

		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color(selectedColor)).
			Foreground(lipgloss.Color(selectedColor)).
			Padding(0, 0, 0, 1).
			Bold(true).
			Width(itemWidth - 2).
			MaxWidth(itemWidth - 2)
	}

	// Create a function to get the appropriate selected description style based on focus
	getSelectedDescStyle := func(isFocused bool) lipgloss.Style {
		selectedColor := TextColor
		if isFocused {
			selectedColor = PrimaryColor
		}

		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(selectedColor)).
			Padding(0, 0, 0, 2).
			Width(itemWidth - 2).
			MaxWidth(itemWidth - 2)
	}

	// Create delegates for each list with appropriate focus state
	trackDelegate := list.NewDefaultDelegate()
	trackDelegate.Styles.SelectedTitle = getSelectedStyle(m.activeList == FocusSearchTracksView)
	trackDelegate.Styles.SelectedDesc = getSelectedDescStyle(m.activeList == FocusSearchTracksView)

	playlistDelegate := list.NewDefaultDelegate()
	playlistDelegate.Styles.SelectedTitle = getSelectedStyle(m.activeList == FocusSearchPlaylistsView)
	playlistDelegate.Styles.SelectedDesc = getSelectedDescStyle(m.activeList == FocusSearchPlaylistsView)

	artistDelegate := list.NewDefaultDelegate()
	artistDelegate.Styles.SelectedTitle = getSelectedStyle(m.activeList == FocusSearchArtistsView)
	artistDelegate.Styles.SelectedDesc = getSelectedDescStyle(m.activeList == FocusSearchArtistsView)

	albumDelegate := list.NewDefaultDelegate()
	albumDelegate.Styles.SelectedTitle = getSelectedStyle(m.activeList == FocusSearchAlbumsView)
	albumDelegate.Styles.SelectedDesc = getSelectedDescStyle(m.activeList == FocusSearchAlbumsView)

	// Update the delegates for all lists
	m.trackList.SetDelegate(trackDelegate)
	m.playlistList.SetDelegate(playlistDelegate)
	m.albumList.SetDelegate(albumDelegate)
	m.artistList.SetDelegate(artistDelegate)
}

// SetActiveList sets which list is currently active based on the focused model
func (m *searchViewModel) SetActiveList(focusedModel FocusedModel) {
	m.activeList = focusedModel
	m.updateListStyles(m.width / 2)
}
