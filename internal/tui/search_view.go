package tui

import (
	"context"
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
	"github.com/zmb3/spotify/v2"
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
	ctx           context.Context
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
func newSearchView(ctx context.Context, spotifyState *state.SpotifyState) searchViewModel {
	m := searchViewModel{
		ctx:          ctx,
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
		m.UpdateSearchResults()

	case tea.KeyMsg:
		var cmd tea.Cmd
		// Update only the active list based on the current focus
		switch m.activeList {
		case FocusSearchTracksView:
			m.trackList, cmd = m.trackList.Update(msg)
			cmds = append(cmds, cmd)
			// If the user presses enter, play the selected track
			if key.Matches(msg, DefaultKeyMap.Select) {
				if len(m.trackList.Items()) > 0 {
					index := m.trackList.Index()
					id := m.spotifyState.GetSearchResultTracks()[index].ID
					// Play the track and remain on the current view
					return m, tea.Batch(
						m.spotifyState.PlayTrack(m.ctx, id),
						// Stay in search view with current focus
						NavigateCmd(FocusSearchTracksView, false, "", playlistView),
					)
				}
			}
			if key.Matches(msg, DefaultKeyMap.AddToQueue) {
				if len(m.trackList.Items()) > 0 {
					index := m.trackList.Index()
					fullTrack := m.spotifyState.GetSearchResultTracks()[index]

					m.spotifyState.Queue.Enqueue(fullTrack.SimpleTrack)
					return m, state.UpdateQueue()

				}
			}

		case FocusSearchPlaylistsView:
			m.playlistList, cmd = m.playlistList.Update(msg)
			cmds = append(cmds, cmd)
			// If they press select then we need to open the table view for the selected playlist
			if key.Matches(msg, DefaultKeyMap.Select) {
				if len(m.playlistList.Items()) > 0 {
					index := m.playlistList.Index()
					playlistID := m.spotifyState.GetSearchResultPlaylists()[index].ID
					log.Println("Opening playlist view for: ", playlistID)
					m.spotifyState.SetSelectedID(playlistID)

					//We somehow need to pass along that we opened a playlist called XX

					// Use helper function for cleaner code
					return m, navigateToPlaylistView(playlistID, playlistView)
				}
			}

		case FocusSearchArtistsView:
			m.artistList, cmd = m.artistList.Update(msg)
			cmds = append(cmds, cmd)
			// If user selects an artist, navigate to view their tracks/albums
			if key.Matches(msg, DefaultKeyMap.Select) {
				if len(m.artistList.Items()) > 0 {
					index := m.artistList.Index()
					//artistID := m.spotifyState.SearchResults.Artists[index].ID
					artistID := m.spotifyState.GetSearchResultArtists()[index].ID
					log.Println("Opening artist view for: ", artistID)
					m.spotifyState.SetSelectedID(artistID)

					// Use helper function
					return m, navigateToPlaylistView(artistID, artistTopTracksView)
				}
			}

		case FocusSearchAlbumsView:
			m.albumList, cmd = m.albumList.Update(msg)
			cmds = append(cmds, cmd)
			// If user selects an album, navigate to view its tracks
			if key.Matches(msg, DefaultKeyMap.Select) {
				if len(m.albumList.Items()) > 0 {
					index := m.albumList.Index()
					//albumID := m.spotifyState.SearchResults.Albums[index].ID
					albumID := m.spotifyState.GetSearchResultAlbums()[index].ID

					log.Println("Opening album view for: ", albumID)
					m.spotifyState.SetSelectedID(albumID)
					// Use helper function
					return m, navigateToPlaylistView(albumID, albumTracksView)
				}
			}
		}

	}

	return m, tea.Batch(cmds...)
}

// View renders the search view
func (m searchViewModel) View() string {
	// Base style for all lists
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(BorderColor).
		Padding(0, 0)

	// Style for the focused list
	focusedStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(PrimaryColor).
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
	trackSearchResults := m.spotifyState.GetSearchResultTracks()
	for _, track := range trackSearchResults {
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
	playlistSearchResults := m.spotifyState.GetSearchResultPlaylists()
	for _, playlist := range playlistSearchResults {

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
	albumSearchResults := m.spotifyState.GetSearchResultAlbums()
	for _, album := range albumSearchResults {
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
	artistSearchResults := m.spotifyState.GetSearchResultArtists()
	for _, artist := range artistSearchResults {
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
		Background(BorderColor).
		Foreground(WhiteTextColor).
		Padding(0, 1)

	return l
}

func (m *searchViewModel) updateListStyles(itemWidth int) {
	delegate := list.NewDefaultDelegate()

	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(WhiteTextColor).
		Padding(0, 0, 0, 2).
		Width(itemWidth - 2).
		MaxWidth(itemWidth - 2)

	delegate.Styles.NormalDesc = delegate.Styles.NormalTitle.
		Foreground(TextColor).
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
			BorderForeground(selectedColor).
			Foreground(selectedColor).
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
			Foreground(selectedColor).
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

// GetNextTrack returns the ID of the next track to play in the active search view
func (m *searchViewModel) GetNextTrack(focusedModel FocusedModel) spotify.ID {
	// Check which list is active and get tracks from that list
	switch focusedModel {
	case FocusSearchTracksView:
		// For tracks, we should play the next track in the search results
		tracks := m.spotifyState.GetSearchResultTracks()
		return m.getNextTrackFromList(tracks)

	case FocusSearchArtistsView:
		// For artists, we need to look at the tracks loaded in the state
		// Since selecting an artist loads their top tracks
		tracks := m.spotifyState.GetTracks()
		if len(tracks) == 0 {
			return ""
		}

		// Convert SimpleTrack to FullTrack format for the helper method
		fullTracks := make([]spotify.FullTrack, len(tracks))
		for i, track := range tracks {
			fullTracks[i] = spotify.FullTrack{
				SimpleTrack: track,
			}
		}
		return m.getNextTrackFromList(fullTracks)

	case FocusSearchAlbumsView:
		// For albums, we also need to look at the tracks loaded in the state
		// Since selecting an album loads its tracks
		tracks := m.spotifyState.GetTracks()
		if len(tracks) == 0 {
			return ""
		}

		// Convert SimpleTrack to FullTrack format for the helper method
		fullTracks := make([]spotify.FullTrack, len(tracks))
		for i, track := range tracks {
			fullTracks[i] = spotify.FullTrack{
				SimpleTrack: track,
			}
		}
		return m.getNextTrackFromList(fullTracks)

	case FocusSearchPlaylistsView:
		// For playlists, we need to look at the tracks loaded in the state
		// Since selecting a playlist loads its tracks
		tracks := m.spotifyState.GetTracks()
		if len(tracks) == 0 {
			return ""
		}

		// Convert SimpleTrack to FullTrack format for the helper method
		fullTracks := make([]spotify.FullTrack, len(tracks))
		for i, track := range tracks {
			fullTracks[i] = spotify.FullTrack{
				SimpleTrack: track,
			}
		}
		return m.getNextTrackFromList(fullTracks)
	}
	return ""
}

// Helper method to get the next track from a list of tracks
func (m *searchViewModel) getNextTrackFromList(tracks []spotify.FullTrack) spotify.ID {
	if len(tracks) == 0 {
		log.Println("No tracks in search results to autoplay")
		return ""
	}

	// If no track is currently playing, return the first track
	playerState := m.spotifyState.GetPlayerState()
	if playerState.Item == nil {
		log.Println("No track currently playing, returning first track from search")
		return tracks[0].ID
	}

	currentTrackID := playerState.Item.ID

	// Find the current track in the search results
	for i, track := range tracks {
		if string(track.ID) == string(currentTrackID) {
			// If it's the last track, return empty (we'll use recommendations)
			if i >= len(tracks)-1 {
				log.Println("Current track is the last one in search results")
				return ""
			}
			// Return the next track
			log.Printf("Found current track at index %d in search, returning next track", i)
			return tracks[i+1].ID
		}
	}

	// If current track not found in search results, return the first track
	log.Println("Current track not found in search results, returning first track")
	return tracks[0].ID
}
