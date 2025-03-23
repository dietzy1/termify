package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

type viewport int

// Trackview can consit of playlistTracks, artist top tracks, album tracks
const (
	playlistView = iota
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
