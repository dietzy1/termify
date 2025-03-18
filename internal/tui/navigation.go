package tui

import tea "github.com/charmbracelet/bubbletea"

type NavigationMsg struct {
	Target     FocusedModel
	ExitSearch bool
	PlaylistID string // Optional: playlist ID to load when navigating to playlist view
}

func NavigateCmd(target FocusedModel, exitSearch bool, playlistID string) tea.Cmd {
	return func() tea.Msg {
		return NavigationMsg{
			Target:     target,
			ExitSearch: exitSearch,
			PlaylistID: playlistID,
		}
	}
}

// Helper functions for common navigation patterns
func NavigateToPlaylistView(playlistID string) tea.Cmd {
	return NavigateCmd(FocusPlaylistView, true, playlistID)
}

func NavigateToLibrary() tea.Cmd {
	return NavigateCmd(FocusLibrary, false, "")
}

func NavigateToSearch() tea.Cmd {
	return NavigateCmd(FocusSearchBar, true, "")
}

func NavigateToSearchResults() tea.Cmd {
	return NavigateCmd(FocusSearchTracksView, true, "")
}
