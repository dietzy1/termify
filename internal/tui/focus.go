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
	FocusDeviceSelector
	FocusQueue
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

func (m *applicationModel) cycleFocus() tea.Cmd {
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
		return m.spotifyState.SelectPlaylist(string(m.library.list.SelectedItem().(playlist).uri))

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
	case FocusDeviceSelector:
		m.focusedModel = FocusLibrary
		m.deviceSelector.Blur()
	case FocusQueue:
		m.focusedModel = FocusLibrary
		return tea.WindowSize()
	}

	return nil
}

func (m *applicationModel) cycleFocusBackward() tea.Cmd {
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
		return m.spotifyState.SelectPlaylist(string(m.library.list.SelectedItem().(playlist).uri))
	case FocusSearchTracksView, FocusSearchPlaylistsView, FocusSearchArtistsView, FocusSearchAlbumsView:
		// If we're in a search view, cycle through the search views in reverse
		m.cycleSearchViewsBackward()
	case FocusSearchBar:
		// From SearchBar, go to Library (since it's at the top)
		m.focusedModel = FocusLibrary
	}
	return nil
}

func (m applicationModel) handleGlobalKeys(msg tea.KeyMsg) (applicationModel, tea.Cmd, bool) {
	var cmd tea.Cmd
	log.Println("Handling global key:", msg)

	// All return key cases
	switch {
	case key.Matches(msg, DefaultKeyMap.Return) && m.focusedModel == FocusSearchBar:
		m.searchBar.ExitSearchMode()
		m.deviceSelector.Blur()

		return m, tea.Sequence(
			navigateToLibrary(),
			tea.WindowSize(),
		), true

	case key.Matches(msg, DefaultKeyMap.Return) && m.activeViewport == HelpView:
		m.activeViewport = MainView
		m.deviceSelector.Blur()
		return m, nil, false

	case key.Matches(msg, DefaultKeyMap.Return) && m.focusedModel != FocusSearchBar:
		m.searchBar.ExitSearchMode()
		m.deviceSelector.Blur()

		return m, tea.Sequence(
			m.spotifyState.SelectPlaylist(string(m.library.list.SelectedItem().(playlist).uri)),
			navigateToLibrary(),
			tea.WindowSize(),
		), true
	}

	if m.searchBar.searching && m.focusedModel == FocusSearchBar {
		//Redirect all remaining keys while searching
		return m, cmd, false
	}

	// Global view changers
	switch {
	case key.Matches(msg, DefaultKeyMap.Search):
		return m, tea.Batch(
			navigateToSearch(),
			tea.WindowSize(),
		), true

	case key.Matches(msg, DefaultKeyMap.Help):
		m.toggleHelpView()
		return m, nil, true

	case key.Matches(msg, DefaultKeyMap.ViewQueue):
		m.activeViewport = MainView
		m.focusedModel = FocusQueue
		return m, tea.WindowSize(), true

	case key.Matches(msg, DefaultKeyMap.Device):
		m.focusedModel = FocusDeviceSelector
		m.deviceSelector.Focus()
		return m, tea.WindowSize(), true
	}

	//Playback controls globals
	switch {
	case key.Matches(msg, DefaultKeyMap.PlayPause):
		playerState := m.spotifyState.GetPlayerState()
		if playerState.Playing {
			return m, m.spotifyState.PausePlayback(m.ctx), true
		}
		return m, m.spotifyState.StartPlayback(m.ctx), true
	case key.Matches(msg, DefaultKeyMap.Next):
		model, cmd := m.handleAutoplay()
		return model, cmd, true
	case key.Matches(msg, DefaultKeyMap.Previous):
		return m, m.spotifyState.PreviousTrack(m.ctx), true
	case key.Matches(msg, DefaultKeyMap.Shuffle):
		return m, m.spotifyState.ToggleShuffleMode(m.ctx), true
	case key.Matches(msg, DefaultKeyMap.Repeat):
		return m, m.spotifyState.ToggleRepeatMode(m.ctx), true
	case key.Matches(msg, DefaultKeyMap.VolumeUp):
		return m, m.spotifyState.IncreaseVolume(m.ctx), true
	case key.Matches(msg, DefaultKeyMap.VolumeDown):
		return m, m.spotifyState.DecreaseVolume(m.ctx), true
	case key.Matches(msg, DefaultKeyMap.VolumeMute):
		return m, m.spotifyState.Mute(m.ctx), true
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

	case FocusDeviceSelector:
		deviceSelector, cmd := m.deviceSelector.Update(msg)
		m.deviceSelector = deviceSelector.(deviceSelectorModel)
		cmds = append(cmds, cmd)

	case FocusQueue:
		queue, cmd := m.queueView.Update(msg)
		m.queueView = queue.(queueModel)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *applicationModel) toggleHelpView() {
	if m.activeViewport == HelpView {
		m.activeViewport = MainView
	} else {
		m.activeViewport = HelpView
	}
}

// Helper function to get border style based on focus state
func getBorderStyle(isFocused bool) lipgloss.Color {
	if isFocused {
		return lipgloss.Color(PrimaryColor)
	}
	return lipgloss.Color(BorderColor)
}
