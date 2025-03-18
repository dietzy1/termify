package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FocusedModel int

const (
	FocusLibrary FocusedModel = iota
	FocusSearchBar

	FocusPlaylistView
	FocusSearchTracksView
	FocusSearchPlaylistsView
	FocusSearchArtistsView
	FocusSearchAlbumsView
)

func (m applicationModel) isSearchViewFocus() bool {
	return m.focusedModel == FocusSearchTracksView ||
		m.focusedModel == FocusSearchPlaylistsView ||
		m.focusedModel == FocusSearchArtistsView ||
		m.focusedModel == FocusSearchAlbumsView
}

func (m *applicationModel) getDefaultSearchView() FocusedModel {
	return FocusSearchTracksView
}

// cycleSearchViews cycles through the search views
func (m *applicationModel) cycleSearchViews() {
	switch m.focusedModel {
	case FocusSearchTracksView:
		m.focusedModel = FocusSearchPlaylistsView
	case FocusSearchPlaylistsView:
		m.focusedModel = FocusSearchArtistsView
	case FocusSearchArtistsView:
		m.focusedModel = FocusSearchAlbumsView
	case FocusSearchAlbumsView:
		m.focusedModel = FocusSearchTracksView
	default:
		m.focusedModel = FocusSearchTracksView
	}
}

// cycleSearchViewsBackward cycles through the search views in reverse
func (m *applicationModel) cycleSearchViewsBackward() {
	switch m.focusedModel {
	case FocusSearchTracksView:
		m.focusedModel = FocusSearchAlbumsView
	case FocusSearchPlaylistsView:
		m.focusedModel = FocusSearchTracksView
	case FocusSearchArtistsView:
		m.focusedModel = FocusSearchPlaylistsView
	case FocusSearchAlbumsView:
		m.focusedModel = FocusSearchArtistsView
	default:
		m.focusedModel = FocusSearchTracksView
	}
}

func (m *applicationModel) cycleFocus() {
	switch m.focusedModel {
	case FocusLibrary:
		// From Library, go to either PlaylistView or SearchView depending on search state
		if m.searchBar.searching {
			m.focusedModel = m.getDefaultSearchView()
		} else {
			m.focusedModel = FocusPlaylistView
		}
	case FocusPlaylistView:
		// From PlaylistView, go to Library
		m.focusedModel = FocusLibrary
	case FocusSearchTracksView, FocusSearchPlaylistsView, FocusSearchArtistsView, FocusSearchAlbumsView:
		// If we're in a search view, cycle through the search views
		m.cycleSearchViews()
	case FocusSearchBar:
		// From SearchBar, go to the appropriate content view
		if m.searchBar.searching {
			m.focusedModel = m.getDefaultSearchView()
		} else {
			m.focusedModel = FocusPlaylistView
		}
	}
}

func (m *applicationModel) cycleFocusBackward() {
	switch m.focusedModel {
	case FocusLibrary:
		// From Library, go to either PlaylistView or SearchView depending on search state
		if m.searchBar.searching {
			m.focusedModel = m.getDefaultSearchView()
		} else {
			m.focusedModel = FocusPlaylistView
		}
	case FocusPlaylistView:
		// From PlaylistView, go to Library
		m.focusedModel = FocusLibrary
	case FocusSearchTracksView, FocusSearchPlaylistsView, FocusSearchArtistsView, FocusSearchAlbumsView:
		// If we're in a search view, cycle through the search views in reverse
		m.cycleSearchViewsBackward()
	case FocusSearchBar:
		// From SearchBar, go to Library (since it's at the top)
		m.focusedModel = FocusLibrary
	}
}

func (m applicationModel) handleGlobalKeys(msg tea.KeyMsg) (applicationModel, tea.Cmd, bool) {
	var cmd tea.Cmd

	log.Println("Handling global key:", msg)

	// These keys should always work
	switch {
	case key.Matches(msg, DefaultKeyMap.Quit):
		return m, tea.Quit, true
	case key.Matches(msg, DefaultKeyMap.Help):
		m.showHelp = !m.showHelp
		return m, nil, true
	}

	// If we're in help mode, check for Return key to exit help
	if m.showHelp {
		if key.Matches(msg, DefaultKeyMap.Return) {
			m.showHelp = false
			return m, nil, true
		}
		// Let other keys pass through when in help mode
		return m, nil, false
	}

	// Handle other global keys
	switch {
	case key.Matches(msg, DefaultKeyMap.Search):
		// Toggle search mode
		if m.searchBar.searching {
			if m.isSearchViewFocus() {
				return m, NavigateToSearch(), true
			}
		} else {
			return m, NavigateToSearch(), true
		}
	case key.Matches(msg, DefaultKeyMap.Return) && m.focusedModel == FocusSearchBar:
		m.searchBar.ExitSearchMode()
		return m, NavigateToLibrary(), true
	case key.Matches(msg, DefaultKeyMap.Return) && m.focusedModel != FocusSearchBar:
		return m, NavigateToLibrary(), true
	}

	if m.searchBar.searching {
		return m, cmd, false
	}

	//Playback controls globals
	switch {
	case key.Matches(msg, DefaultKeyMap.PlayPause):
		if m.spotifyState.PlayerState.Playing {
			return m, m.spotifyState.PausePlayback(), true
		}
		return m, m.spotifyState.StartPlayback(), true
	case key.Matches(msg, DefaultKeyMap.Next):
		return m, m.spotifyState.NextTrack(), true
	case key.Matches(msg, DefaultKeyMap.Previous):
		return m, m.spotifyState.PreviousTrack(), true
	case key.Matches(msg, DefaultKeyMap.Shuffle):
		return m, m.spotifyState.ToggleShuffleMode(), true
	case key.Matches(msg, DefaultKeyMap.Repeat):
		return m, m.spotifyState.ToggleRepeatMode(), true
	case key.Matches(msg, DefaultKeyMap.VolumeUp):
		return m, m.spotifyState.IncreaseVolume(), true
	case key.Matches(msg, DefaultKeyMap.VolumeDown):
		return m, m.spotifyState.DecreaseVolume(), true

	}
	log.Println("Unhandled key:", msg)

	return m, cmd, false
}

func (m applicationModel) updateFocusedModel(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch m.focusedModel {
	case FocusLibrary:
		library, cmd := m.library.Update(msg)
		m.library = library.(libraryModel)
		cmds = append(cmds, cmd)

	case FocusPlaylistView:
		playlistView, cmd := m.playlistView.Update(msg)
		m.playlistView = playlistView.(playlistViewModel)
		cmds = append(cmds, cmd)

	case FocusSearchTracksView, FocusSearchPlaylistsView, FocusSearchArtistsView, FocusSearchAlbumsView:
		m.searchView.SetActiveList(m.focusedModel)
		searchView, cmd := m.searchView.Update(msg)
		m.searchView = searchView.(searchViewModel)
		cmds = append(cmds, cmd)

	case FocusSearchBar:
		searchBar, cmd := m.searchBar.Update(msg)
		m.searchBar = searchBar.(searchbarModel)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// Helper function to get border style based on focus state
func getBorderStyle(isFocused bool) lipgloss.Color {
	if isFocused {
		return lipgloss.Color(PrimaryColor)
	}
	return lipgloss.Color(BorderColor)
}
