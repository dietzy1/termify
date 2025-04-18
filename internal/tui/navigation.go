package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zmb3/spotify/v2"
)

// Trackview can consit of playlistTracks, artist top tracks, album tracks
const (
	playlistView viewport = iota
	artistTopTracksView
	albumTracksView
)

type NavigationMsg struct {
	Target     FocusedModel
	ExitSearch bool
	selectedID spotify.ID // Optional: playlist ID to load when navigating to playlist view
	viewport   viewport
}

func NavigateCmd(target FocusedModel, exitSearch bool, selectedID spotify.ID, vp viewport) tea.Cmd {
	return func() tea.Msg {
		return NavigationMsg{
			Target:     target,
			ExitSearch: exitSearch,
			selectedID: selectedID,
			viewport:   vp,
		}
	}
}

// Helper functions for common navigation patterns
func NavigateToPlaylistView(selectedID spotify.ID, vp viewport) tea.Cmd {
	return NavigateCmd(FocusPlaylistView, true, selectedID, vp)
}

func NavigateToLibrary() tea.Cmd {
	return NavigateCmd(FocusLibrary, false, "", playlistView)
}

func NavigateToSearch() tea.Cmd {
	return NavigateCmd(FocusSearchBar, true, "", playlistView)
}

func NavigateToSearchResults() tea.Cmd {
	return NavigateCmd(FocusSearchTracksView, true, "", playlistView)
}

func (m applicationModel) handleNavigationMsg(msg NavigationMsg) (applicationModel, tea.Cmd) {
	var cmds []tea.Cmd
	m.focusedModel = msg.Target

	if msg.ExitSearch {
		m.searchBar.searching = false
		m.searchBar.textInput.Blur()
		m.searchBar.textInput.SetValue("")
	}

	if msg.Target == FocusSearchBar {
		m.searchBar.EnterSearchMode()
	}

	if msg.selectedID != "" {
		switch msg.viewport {
		case playlistView:
			cmds = append(cmds, m.spotifyState.FetchPlaylistTracks(msg.selectedID))
		case artistTopTracksView:
			cmds = append(cmds, m.spotifyState.FetchTopTracks(spotify.ID(msg.selectedID)))
		case albumTracksView:
			cmds = append(cmds, m.spotifyState.FetchAlbumTracks(msg.selectedID))
		}
	}
	return m, tea.Batch(cmds...)
}

// renderNavigationHelp shows a simple help message for navigation
func (m applicationModel) renderNavigationHelp() string {
	var focusName string
	var helpText string

	switch m.focusedModel {
	case FocusLibrary:
		focusName = "Library"
		helpText = "Tab: Switch to content view | /: Search"
	case FocusPlaylistView:
		focusName = "Playlist"
		helpText = "Tab: Switch to library | /: Search"
	case FocusSearchTracksView:
		focusName = "Search Tracks"
		helpText = "Tab: Cycle search views | /: Search"
	case FocusSearchPlaylistsView:
		focusName = "Search Playlists"
		helpText = "Tab: Cycle search views | /: Search"
	case FocusSearchArtistsView:
		focusName = "Search Artists"
		helpText = "Tab: Cycle search views | /: Search"
	case FocusSearchAlbumsView:
		focusName = "Search Albums"
		helpText = "Tab: Cycle search views | /: Search"
	case FocusSearchBar:
		focusName = "Search"
		helpText = "Esc: Exit search | tab: Navigate to content"
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Align(lipgloss.Center).
		Width(m.width)

	return helpStyle.Render("Focus: " + focusName + " | " + helpText)
}
