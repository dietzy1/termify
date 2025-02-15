package tui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zmb3/spotify/v2"
)

type applicationModel struct {
	spotifyState *SpotifyState

	focusedModel FocusedModel
	navbar       navbarModel
	/* searchBar       searchbarModel */
	library         libraryModel
	viewport        viewportModel
	playbackControl playbackControlsModel
	audioPlayer     audioPlayerModel
}

func (m applicationModel) Init() tea.Cmd {
	log.Println("Application: Initializing application model")
	return tea.Batch(
		tea.WindowSize(),
		m.spotifyState.FetchPlaylists(),
		m.spotifyState.GetPlaybackState(),
		m.audioPlayer.Init(),
	)
}

func newApplication(client *spotify.Client) applicationModel {
	log.Printf("Application: Creating new application with client: %v", client != nil)

	spotifyState := NewSpotifyState(client)
	log.Printf("Application: Created SpotifyState instance: %v", spotifyState != nil)

	navbar := newNavbar()
	/* searchBar := newSearchbar() */
	library := newLibrary(spotifyState)
	viewport := newViewport(spotifyState)
	playbackControl := newPlaybackControlsModel(spotifyState)
	audioPlayer := newAudioPlayer(spotifyState)

	return applicationModel{
		spotifyState: spotifyState,
		navbar:       navbar,
		/* searchBar:       searchBar, */
		library:         library,
		viewport:        viewport,
		playbackControl: playbackControl,
		audioPlayer:     audioPlayer,
	}
}

func (m applicationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case PlayerStateUpdatedMsg:
		// Update playbackControl with new state
		if updatedPlaybackControl, cmd, ok := updateSubmodel(m.playbackControl, msg, m.playbackControl); ok {
			m.playbackControl = updatedPlaybackControl
			cmds = append(cmds, cmd)
		}
		// Update audioPlayer with new state
		if updatedAudioPlayer, cmd, ok := updateSubmodel(m.audioPlayer, msg, m.audioPlayer); ok {
			m.audioPlayer = updatedAudioPlayer
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case PlaylistsUpdatedMsg:
		if updatedLibrary, cmd, ok := updateSubmodel(m.library, msg, m.library); ok {
			m.library = updatedLibrary
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case TracksUpdatedMsg:
		// Update viewport with new tracks
		if updatedViewport, cmd, ok := updateSubmodel(m.viewport, msg, m.viewport); ok {
			m.viewport = updatedViewport
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case PlaylistSelectedMsg:
		cmds = append(cmds, m.spotifyState.FetchPlaylistTracks(msg.PlaylistID))
		return m, tea.Batch(cmds...)

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
	case tickMsg:
		// Ensure tick and progress messages are passed to audioPlayer
		if updatedAudioPlayer, cmd, ok := updateSubmodel(m.audioPlayer, msg, m.audioPlayer); ok {
			m.audioPlayer = updatedAudioPlayer
			cmds = append(cmds, cmd)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m applicationModel) View() string {
	libraryStyle := applyFocusStyle(m.focusedModel == FocusLibrary)
	viewportStyle := applyFocusStyle(m.focusedModel == FocusViewport)
	combinedPlaybackSectionStyle := applyFocusStyle(m.focusedModel == FocusPlaybackControl)

	viewport := viewportStyle.Render(m.viewport.View())
	library := libraryStyle.Render(m.library.View())
	playback := m.playbackControl.View()
	audioPlayer := m.audioPlayer.View()

	return lipgloss.JoinVertical(
		lipgloss.Center,
		m.navbar.View(),
		lipgloss.JoinHorizontal(lipgloss.Top, library, viewport),
		combinedPlaybackSectionStyle.Render(lipgloss.JoinVertical(lipgloss.Center, playback, audioPlayer)),
	)
}

func (m applicationModel) handleWindowSizeMsg(msg tea.WindowSizeMsg) (applicationModel, tea.Cmd) {
	var cmds []tea.Cmd
	msg.Height -= lipgloss.Height(m.navbar.View()) + lipgloss.Height(m.playbackControl.View()) + lipgloss.Height(m.audioPlayer.View()) + 4

	if updatedNavbar, cmd, ok := updateSubmodel(m.navbar, tea.WindowSizeMsg{
		Width: msg.Width - 2,
	}, m.navbar); ok {
		m.navbar = updatedNavbar
		cmds = append(cmds, cmd)
	} else {
		return m, tea.Quit
	}

	if updatedLibrary, cmd, ok := updateSubmodel(m.library, tea.WindowSizeMsg{
		Width:  28,
		Height: msg.Height,
	}, m.library); ok {
		m.library = updatedLibrary
		cmds = append(cmds, cmd)
	} else {
		return m, tea.Quit
	}

	if updatedViewport, cmd, ok := updateSubmodel(m.viewport, tea.WindowSizeMsg{
		Width:  msg.Width - 4 - m.library.list.Width(),
		Height: msg.Height,
	}, m.viewport); ok {
		m.viewport = updatedViewport
		cmds = append(cmds, cmd)
	} else {
		return m, tea.Quit
	}

	if updatedPlaybackControl, cmd, ok := updateSubmodel(m.playbackControl, tea.WindowSizeMsg{
		Width: msg.Width - 2,
	}, m.playbackControl); ok {
		m.playbackControl = updatedPlaybackControl
		cmds = append(cmds, cmd)
	} else {
		return m, tea.Quit
	}

	if updatedAudioPlayer, cmd, ok := updateSubmodel(m.audioPlayer, tea.WindowSizeMsg{
		Width: msg.Width - 2,
	}, m.audioPlayer); ok {
		m.audioPlayer = updatedAudioPlayer
		cmds = append(cmds, cmd)
	} else {
		return m, tea.Quit
	}

	return m, tea.Batch(cmds...)
}
