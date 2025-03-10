package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
	"github.com/zmb3/spotify/v2"
)

type applicationModel struct {
	width, height int

	spotifyState *state.SpotifyState

	focusedModel    FocusedModel
	navbar          navbarModel
	library         libraryModel
	viewport        viewportModel
	playbackControl playbackControlsModel
	audioPlayer     audioPlayerModel

	helpModel helpModel
	showHelp  bool
	//showError bool
}

func (m applicationModel) Init() tea.Cmd {
	log.Println("Application: Initializing application model")
	return tea.Batch(
		tea.WindowSize(),
		m.spotifyState.FetchPlaylists(),
		m.spotifyState.FetchPlaybackState(),
		m.spotifyState.FetchDevices(),
		m.audioPlayer.Init(),
	)
}

func newApplication(client *spotify.Client) applicationModel {
	spotifyState := state.NewSpotifyState(client)
	log.Printf("Application: Created SpotifyState instance: %v", spotifyState != nil)

	return applicationModel{
		spotifyState:    spotifyState,
		navbar:          newNavbar(spotifyState),
		library:         newLibrary(spotifyState),
		viewport:        newViewport(spotifyState),
		playbackControl: newPlaybackControlsModel(spotifyState),
		audioPlayer:     newAudioPlayer(spotifyState),
		helpModel:       newHelp(),
		//showError:       true,
	}
}

func (m applicationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case state.PlayerStateUpdatedMsg:
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

	case state.PlaylistsUpdatedMsg:
		if updatedLibrary, cmd, ok := updateSubmodel(m.library, msg, m.library); ok {
			m.library = updatedLibrary
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case state.TracksUpdatedMsg:
		// Update viewport with new tracks
		if updatedViewport, cmd, ok := updateSubmodel(m.viewport, msg, m.viewport); ok {
			m.viewport = updatedViewport
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case state.PlaylistSelectedMsg:
		cmds = append(cmds, m.spotifyState.FetchPlaylistTracks(msg.PlaylistID))
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:

		log.Println("Updating model with message type", msg)
		updatedModel, cmd, handled := m.handleGlobalKeys(msg)
		m = updatedModel
		if handled {
			return m, cmd
		}

		switch {
		case key.Matches(msg, DefaultKeyMap.CycleFocusForward):
			m.cycleFocus()
		case key.Matches(msg, DefaultKeyMap.CycleFocusBackward):
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

	//const error = true

	if m.showHelp {
		return m.viewHelp()
	}

	libraryStyle := applyFocusStyle(m.focusedModel == FocusLibrary)
	viewportStyle := applyFocusStyle(m.focusedModel == FocusViewport)
	combinedPlaybackSectionStyle := lipgloss.NewStyle().MaxWidth(m.width)

	viewport := viewportStyle.Render(m.viewport.View())
	library := libraryStyle.Render(m.library.View())

	// Get the song info and volume control views
	songInfoView := m.audioPlayer.songInfoView()
	volumeControlView := m.audioPlayer.volumeControlView()

	// Calculate the available width for the center section
	availableWidth := m.width - lipgloss.Width(songInfoView) - lipgloss.Width(volumeControlView) - 2

	// Center both components individually
	centeredPlaybackControls := lipgloss.NewStyle().
		Width(availableWidth).
		Align(lipgloss.Center).
		Render(m.playbackControl.View())

	centeredAudioPlayer := lipgloss.NewStyle().
		Width(availableWidth).
		Align(lipgloss.Center).
		Render(m.audioPlayer.View())

	// Join them vertically
	centerSection := lipgloss.JoinVertical(
		lipgloss.Center,
		centeredPlaybackControls,
		centeredAudioPlayer,
	)

	/* errorBar := lipgloss.NewStyle().
	Foreground(lipgloss.Color("#ff4444")).
	Border(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#ff4444")).
	Width(m.width-2).
	Height(2).
	Padding(0, 1).
	Render(lipgloss.JoinVertical(
		lipgloss.Left,
		"Error: Critical System Failure",
		"Details: Unable to connect to Spotify API. Please check your internet connection.",
	)) */

	return lipgloss.JoinVertical(
		lipgloss.Center,

		m.navbar.View(),
		/* errorBar, */
		lipgloss.JoinHorizontal(lipgloss.Top, library, viewport),
		combinedPlaybackSectionStyle.Render(
			lipgloss.JoinHorizontal(lipgloss.Bottom,
				songInfoView,
				centerSection,
				volumeControlView),
		),
		"\r",
	)
}

func (m applicationModel) viewHelp() string {

	combinedPlaybackSectionStyle := lipgloss.NewStyle().MaxWidth(m.width)
	songInfoView := m.audioPlayer.songInfoView()
	volumeControlView := m.audioPlayer.volumeControlView()

	// Calculate the available width for the center section
	availableWidth := m.width - lipgloss.Width(songInfoView) - lipgloss.Width(volumeControlView) - 2

	centeredPlaybackControls := lipgloss.NewStyle().
		Width(availableWidth).
		Align(lipgloss.Center).
		Render(m.playbackControl.View())

	centeredAudioPlayer := lipgloss.NewStyle().
		Width(availableWidth).
		Align(lipgloss.Center).
		Render(m.audioPlayer.View())

	// Join them vertically
	centerSection := lipgloss.JoinVertical(
		lipgloss.Center,
		centeredPlaybackControls,
		centeredAudioPlayer,
	)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		m.navbar.View(),
		lipgloss.NewStyle().Height(m.height-lipgloss.Height(m.navbar.View())-lipgloss.Height(centerSection)).Render(m.helpModel.View()),
		combinedPlaybackSectionStyle.Render(
			lipgloss.JoinHorizontal(lipgloss.Bottom,
				songInfoView,
				centerSection,
				volumeControlView),
		),
		"\r",
	)
}

func (m applicationModel) handleWindowSizeMsg(msg tea.WindowSizeMsg) (applicationModel, tea.Cmd) {
	var cmds []tea.Cmd
	m.width = msg.Width
	m.height = msg.Height // Yet to use this for anything really
	//Check if errorMsg exists

	var errorSubstracter int = 0
	/* if m.showError {
		errorSubstracter = 4
	} */

	msg.Height -= lipgloss.Height(m.navbar.View()) + lipgloss.Height(m.playbackControl.View()) + lipgloss.Height(m.audioPlayer.View()) + 3 + errorSubstracter

	if updatedNavbar, cmd, ok := updateSubmodel(m.navbar, tea.WindowSizeMsg{
		Width: msg.Width,
	}, m.navbar); ok {
		m.navbar = updatedNavbar
		cmds = append(cmds, cmd)
	}

	if updatedLibrary, cmd, ok := updateSubmodel(m.library, tea.WindowSizeMsg{
		Width:  28,
		Height: msg.Height,
	}, m.library); ok {
		m.library = updatedLibrary
		cmds = append(cmds, cmd)
	}

	if updatedViewport, cmd, ok := updateSubmodel(m.viewport, tea.WindowSizeMsg{
		Width:  msg.Width - 4 - m.library.list.Width(),
		Height: msg.Height,
	}, m.viewport); ok {
		m.viewport = updatedViewport
		cmds = append(cmds, cmd)
	}

	if updatedPlaybackControl, cmd, ok := updateSubmodel(m.playbackControl, tea.WindowSizeMsg{
		Width: msg.Width,
	}, m.playbackControl); ok {
		m.playbackControl = updatedPlaybackControl
		cmds = append(cmds, cmd)
	}

	if updatedAudioPlayer, cmd, ok := updateSubmodel(m.audioPlayer, tea.WindowSizeMsg{
		Width: msg.Width,
	}, m.audioPlayer); ok {
		m.audioPlayer = updatedAudioPlayer
		cmds = append(cmds, cmd)
	}

	if updatedHelp, cmd, ok := updateSubmodel(m.helpModel, tea.WindowSizeMsg{
		Width:  msg.Width,
		Height: msg.Height + 2,
	}, m.helpModel); ok {
		m.helpModel = updatedHelp
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m applicationModel) handleGlobalKeys(msg tea.KeyMsg) (applicationModel, tea.Cmd, bool) {
	var cmd tea.Cmd

	log.Println("Handling global key:", msg)
	switch {
	case key.Matches(msg, DefaultKeyMap.Quit):
		return m, tea.Quit, true
	case key.Matches(msg, DefaultKeyMap.Help):
		m.showHelp = !m.showHelp
		return m, nil, true

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

	// If we're in help mode, check for Return key to exit help
	if m.showHelp {
		if key.Matches(msg, DefaultKeyMap.Return) {
			m.showHelp = false
			return m, nil, true
		}
	}

	log.Println("Unhandled key:", msg)

	return m, cmd, false
}
