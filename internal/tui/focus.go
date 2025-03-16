package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FocusedModel int

const (
	FocusLibrary FocusedModel = iota
	FocusSearchBar

	FocusPlaylistView
	FocusSearchView
)

//How we want focus to work:
// Tab key cycles focus between main components:
// - Library <-> PlaylistView/SearchView (depending on search state)
// - Press / to enter search mode and focus the SearchBar
// - In SearchBar, press Enter to submit search and focus SearchView
// - In SearchBar, press Esc to exit search mode and return to previous focus
// - In SearchBar, press Up/Down to move focus to the view below (PlaylistView or SearchView)
// - Shift+Tab reverses the tab order

//const focusModelCount = 2

func (m *applicationModel) cycleFocus() {
	switch m.focusedModel {
	case FocusLibrary:
		// From Library, go to either PlaylistView or SearchView depending on search state
		if m.searchBar.searching {
			m.focusedModel = FocusSearchView
		} else {
			m.focusedModel = FocusPlaylistView
		}
	case FocusPlaylistView, FocusSearchView:
		// From content views, go to Library
		m.focusedModel = FocusLibrary
	case FocusSearchBar:
		// From SearchBar, go to the appropriate content view
		if m.searchBar.searching {
			m.focusedModel = FocusSearchView
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
			m.focusedModel = FocusSearchView
		} else {
			m.focusedModel = FocusPlaylistView
		}
	case FocusPlaylistView, FocusSearchView:
		// From content views, go to Library
		m.focusedModel = FocusLibrary
	case FocusSearchBar:
		// From SearchBar, go to Library (since it's at the top)
		m.focusedModel = FocusLibrary
	}
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

	case FocusSearchView:
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
