package tui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zmb3/spotify/v2"
)

// Message types for Spotify events
type playlistSelectedMsg struct {
	playlistID string
}

type tracksLoadedMsg struct {
	tracks []spotify.PlaylistTrack
	err    error
}

func (m model) updateApplication(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle messages that need to be propagated to specific models
	switch msg := msg.(type) {
	case playlistsUpdatedMsg:
		// Propagate playlist updates to library model
		if updatedLibrary, cmd, ok := updateSubmodel(m.library, msg, m.library); ok {
			m.library = updatedLibrary
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case playlistSelectedMsg:
		// When a playlist is selected, fetch its tracks
		log.Printf("Playlist selected: %s", msg.playlistID)
		return m, func() tea.Msg {
			tracks, err := m.spotifyState.FetchPlaylistTracks(msg.playlistID)
			return tracksLoadedMsg{tracks: tracks, err: err}
		}

	case tracksLoadedMsg:
		if msg.err != nil {
			log.Printf("Error loading tracks: %v", msg.err)
			return m, nil
		}
		// Update viewport with new tracks
		if updatedViewport, cmd, ok := updateSubmodel(m.viewport, msg, m.viewport); ok {
			m.viewport = updatedViewport
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	// Handle general application messages
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.cycleFocus()
		case "shift+tab":
			m.cycleFocusBackward()
		default:
			return m.updateFocusedModel(msg)
		}
	case tea.WindowSizeMsg:
		log.Printf("Outer viewport width: %d, height: %d", msg.Width, msg.Height)
		return m.handleWindowSizeMsg(msg)
	/* case authSuccessMsg:
	log.Printf("Handling auth success in application update")
	// Initialize spotify state and library
	m.spotifyClient = msg.client
	m.spotifyState = NewSpotifyState(msg.client)
	m.library = newLibrary(msg.client)
	return m, m.library.Init() */
	case tickMsg:
		// Ensure tick and progress messages are passed to audioPlayer
		if updatedAudioPlayer, cmd, ok := updateSubmodel(m.audioPlayer, msg, m.audioPlayer); ok {
			m.audioPlayer = updatedAudioPlayer
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) viewMain() string {
	libraryStyle := applyFocusStyle(m.focusedModel == FocusLibrary)
	viewportStyle := applyFocusStyle(m.focusedModel == FocusViewport)
	combinedPlaybackSectionStyle := applyFocusStyle(m.focusedModel == FocusPlaybackControl)

	viewport := viewportStyle.Render(m.viewport.View())
	library := libraryStyle.Render(m.library.View())
	playback := m.playbackControl.View()
	audioPlayer := m.audioPlayer.View()

	// Join all sections vertically
	return lipgloss.JoinVertical(
		lipgloss.Center,
		m.navbar.View(),
		lipgloss.JoinHorizontal(lipgloss.Top, library, viewport),
		combinedPlaybackSectionStyle.Render(lipgloss.JoinVertical(lipgloss.Center, playback, audioPlayer)),
	)
}
