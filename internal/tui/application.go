package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
	"github.com/zmb3/spotify/v2"
)

type applicationModel struct {
	width, height int

	spotifyState *state.SpotifyState

	focusedModel   FocusedModel
	activeViewport viewport

	errorBar errorMsg
	navbar   navbarModel
	library  libraryModel

	searchBar       searchbarModel
	playlistView    playlistViewModel
	searchView      searchViewModel
	playbackControl playbackControlsModel
	audioPlayer     audioPlayerModel
	deviceSelector  DeviceSelectorModel
}

func (m applicationModel) Init() tea.Cmd {
	log.Println("Application: Initializing application model")
	return tea.Batch(
		tea.WindowSize(),
		m.searchBar.Init(),
		m.audioPlayer.Init(),
		m.spotifyState.FetchPlaylists(),
		m.spotifyState.FetchPlaybackState(),
		m.spotifyState.FetchQueue(),
		m.spotifyState.FetchDevices(),
	)
}

func newApplication(client *spotify.Client) applicationModel {
	spotifyState := state.NewSpotifyState(client)
	log.Printf("Application: Created SpotifyState instance: %v", spotifyState != nil)

	return applicationModel{
		spotifyState:    spotifyState,
		focusedModel:    FocusLibrary,
		navbar:          newNavbar(),
		library:         newLibrary(spotifyState),
		searchBar:       newSearchbar(spotifyState),
		playlistView:    NewPlaylistView(spotifyState),
		searchView:      NewSearchView(spotifyState),
		playbackControl: newPlaybackControlsModel(spotifyState),
		audioPlayer:     newAudioPlayer(spotifyState),
		errorBar:        errorMsg{},
		activeViewport:  MainView,
	}
}

func (m applicationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case AutoplayNextTrackMsg:
		return m.handleAutoplay()

	case clearQueuedHighlightMsg:
		if updatedPlaylistView, cmd, ok := updateSubmodel(m.playlistView, msg, m.playlistView); ok {
			m.playlistView = updatedPlaylistView
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case ShowToastMsg:
		return m.handleShowErrorMsg(msg)
	case ErrorTimerExpiredMsg:
		m.errorBar.title = ""
		m.errorBar.message = ""
		return m, tea.WindowSize()

	case debouncedSearch:
		if updatedSearchBar, cmd, ok := updateSubmodel(m.searchBar, msg, m.searchBar); ok {
			m.searchBar = updatedSearchBar
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case state.PlayerStateUpdatedMsg:
		if updatedPlaybackControl, cmd, ok := updateSubmodel(m.playbackControl, msg, m.playbackControl); ok {
			m.playbackControl = updatedPlaybackControl
			cmds = append(cmds, cmd)
		}

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
		if updatedPlaylistView, cmd, ok := updateSubmodel(m.playlistView, msg, m.playlistView); ok {
			m.playlistView = updatedPlaylistView
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case state.SearchResultsUpdatedMsg:

		if updatedSearchView, cmd, ok := updateSubmodel(m.searchView, msg, m.searchView); ok {
			m.searchView = updatedSearchView
			cmds = append(cmds, cmd)
		}

		return m, tea.Batch(cmds...)

	case state.PlaylistSelectedMsg:
		cmds = append(cmds, m.spotifyState.FetchPlaylistTracks(spotify.ID(msg.PlaylistID)))
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
			cmd := m.cycleFocus()
			return m, cmd
		case key.Matches(msg, DefaultKeyMap.CycleFocusBackward):
			cmd := m.cycleFocusBackward()
			return m, cmd
		default:
			return m.updateFocusedModel(msg)
		}

	case tea.WindowSizeMsg:
		log.Printf("Outer viewport width: %d, height: %d", msg.Width, msg.Height)
		return m.handleWindowSizeMsg(msg)

	case NavigationMsg:
		return m.handleNavigationMsg(msg)
	case tickMsg:
		// Ensure tick and progress messages are passed to audioPlayer
		if updatedAudioPlayer, cmd, ok := updateSubmodel(m.audioPlayer, msg, m.audioPlayer); ok {
			m.audioPlayer = updatedAudioPlayer
			cmds = append(cmds, cmd)
		}
	case cursor.BlinkMsg:
		if m.focusedModel != FocusSearchBar {
			return m, nil
		}
		if searchbarModel, cmd, ok := updateSubmodel(m.searchBar, msg, m.searchBar); ok {
			m.searchBar = searchbarModel
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

type viewport int

const (
	// Normal view
	MainView viewport = iota
	// Queue view
	QueueView
	// Device view
	DeviceView
	// Help view
	HelpView
)

func (m applicationModel) View() string {

	switch m.activeViewport {
	case QueueView:
		return "Queue"
	case DeviceView:
		return m.viewDevice()
	case HelpView:
		return m.viewHelp()
	case MainView:
	}

	// Update focus state for components
	m.library.isFocused = m.focusedModel == FocusLibrary
	m.searchBar.isFocused = m.focusedModel == FocusSearchBar
	m.playlistView.isFocused = m.focusedModel == FocusPlaylistView

	// Set search view focus state
	m.searchView.isFocused = m.isSearchViewFocus()

	// If search view is focused, set the active list
	if m.isSearchViewFocus() {
		m.searchView.SetActiveList(m.focusedModel)
	}

	// Determine which view to show based on search state
	var viewport string
	if m.searchBar.searching {
		viewport = m.searchView.View()
	} else {
		viewport = m.playlistView.View()
	}

	library := m.library.View()
	playbackSection := m.renderPlaybackSection()

	var navContent []string
	navContent = append(navContent, m.navbar.View())
	errorBar := m.renderErrorBar()
	if errorBar != "" {
		navContent = append(navContent, errorBar)
	}

	navigationHelp := m.renderNavigationHelp()

	return lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Top, navContent...),
		lipgloss.JoinHorizontal(lipgloss.Top,
			library,
			lipgloss.JoinVertical(lipgloss.Top,
				m.searchBar.View(),
				viewport)),
		playbackSection,
		navigationHelp,
	)
}

func (m applicationModel) handleWindowSizeMsg(msg tea.WindowSizeMsg) (applicationModel, tea.Cmd) {
	var cmds []tea.Cmd
	m.width = msg.Width
	m.height = msg.Height

	errorHeight := 0
	if m.errorBar.title != "" {
		errorHeight = lipgloss.Height(m.renderErrorBar())
		log.Println("Error height:", errorHeight)
	}

	const SHRINKHEIGHT = 40
	var navbarHeight = 1
	if m.height > SHRINKHEIGHT {
		navbarHeight = 5
	}

	if updatedNavbar, cmd, ok := updateSubmodel(m.navbar, tea.WindowSizeMsg{
		Width:  msg.Width,
		Height: navbarHeight,
	}, m.navbar); ok {
		m.navbar = updatedNavbar
		cmds = append(cmds, cmd)
	}

	viewportHeight := msg.Height - navbarHeight - lipgloss.Height(m.playbackControl.View()) - lipgloss.Height(m.audioPlayer.View()) - 1 - errorHeight

	if updatedLibrary, cmd, ok := updateSubmodel(m.library, tea.WindowSizeMsg{
		Height: viewportHeight,
	}, m.library); ok {
		m.library = updatedLibrary
		cmds = append(cmds, cmd)
	}

	if updatedSearchBar, cmd, ok := updateSubmodel(m.searchBar, tea.WindowSizeMsg{
		Width: msg.Width - lipgloss.Width(m.library.View()),
	}, m.searchBar); ok {
		m.searchBar = updatedSearchBar
		cmds = append(cmds, cmd)
	}

	if updatedPlaylistView, cmd, ok := updateSubmodel(m.playlistView, tea.WindowSizeMsg{
		Width:  msg.Width - lipgloss.Width(m.library.View()),
		Height: viewportHeight - lipgloss.Height(m.searchBar.View()),
	}, m.playlistView); ok {
		m.playlistView = updatedPlaylistView
		cmds = append(cmds, cmd)
	}

	if updatedSearchView, cmd, ok := updateSubmodel(m.searchView, tea.WindowSizeMsg{
		Width:  msg.Width - lipgloss.Width(m.library.View()),
		Height: viewportHeight - lipgloss.Height(m.searchBar.View()),
	}, m.searchView); ok {
		m.searchView = updatedSearchView
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

	return m, tea.Batch(cmds...)
}
