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

	errorToast      errorToastModel
	navbar          navbarModel
	library         libraryModel
	searchBar       searchbarModel
	playlistView    playlistViewModel
	searchView      searchViewModel
	queueView       queueModel
	playbackControl playbackControlsModel
	audioPlayer     audioPlayerModel
	deviceSelector  deviceSelectorModel
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
		playlistView:    newPlaylistView(spotifyState),
		searchView:      newSearchView(spotifyState),
		queueView:       newQueue(spotifyState),
		playbackControl: newPlaybackControlsModel(spotifyState),
		audioPlayer:     newAudioPlayer(spotifyState),
		deviceSelector:  NewDeviceSelector(spotifyState),
		errorToast:      newErrorToast(),
		activeViewport:  MainView,
	}
}

func (m applicationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case state.QueueUpdatedMsg:
		if updatedQueueView, cmd, ok := updateSubmodel(m.queueView, msg,
			m.queueView); ok {
			m.queueView = updatedQueueView
			cmds = append(cmds, cmd)
		}
		queueCount := m.spotifyState.GetQueue()
		m.navbar.queueCount = len(queueCount)

	case AutoplayNextTrackMsg:
		return m.handleAutoplay()

	case clearQueuedHighlightMsg:
		if updatedPlaylistView, cmd, ok := updateSubmodel(m.playlistView, msg, m.playlistView); ok {
			m.playlistView = updatedPlaylistView
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case state.ErrorMsg:
		if updatedErrorToast, cmd, ok := updateSubmodel(m.errorToast, ShowToastMsg{
			Title:   msg.Title,
			Message: msg.Message,
		}, m.errorToast); ok {
			m.errorToast = updatedErrorToast
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case ShowToastMsg, ErrorTimerExpiredMsg:
		if updatedErrorToast, cmd, ok := updateSubmodel(m.errorToast, msg, m.errorToast); ok {
			m.errorToast = updatedErrorToast
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

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

	case state.DevicesUpdatedMsg:
		if updatedDevices, cmd, ok := updateSubmodel(m.deviceSelector, msg, m.deviceSelector); ok {
			m.deviceSelector = updatedDevices
			cmds = append(cmds, cmd)
		}

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
	MainView viewport = iota
	HelpView
)

func (m applicationModel) View() string {

	var viewportContent string
	switch m.activeViewport {
	case HelpView:
		viewportContent = m.renderHelp()
	case MainView:
		viewportContent = m.viewMain()
	}

	playbackSection := m.renderPlaybackSection()

	var navContent []string
	navContent = append(navContent, m.navbar.View())
	errorToastView := m.errorToast.View()
	if errorToastView != "" {
		navContent = append(navContent, errorToastView)
	}

	navigationHelp := m.renderNavigationHelp()

	return lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Top, navContent...),
		viewportContent,
		playbackSection,
		navigationHelp,
	)
}

func (m applicationModel) viewMain() string {
	m.library.isFocused = m.focusedModel == FocusLibrary
	m.searchBar.isFocused = m.focusedModel == FocusSearchBar
	m.playlistView.isFocused = m.focusedModel == FocusPlaylistView
	m.queueView.isFocused = m.focusedModel == FocusQueue

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
	queue := ""
	if m.focusedModel == FocusQueue {
		queue = m.queueView.View()
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		library,
		lipgloss.JoinVertical(
			lipgloss.Top,
			m.searchBar.View(),
			viewport,
		),
		queue,
	)
}

func (m applicationModel) handleWindowSizeMsg(msg tea.WindowSizeMsg) (applicationModel, tea.Cmd) {
	var cmds []tea.Cmd
	m.width = msg.Width
	m.height = msg.Height

	if updatedErrorToast, cmd, ok := updateSubmodel(m.errorToast, tea.WindowSizeMsg{
		Width: msg.Width,
	}, m.errorToast); ok {
		m.errorToast = updatedErrorToast
		cmds = append(cmds, cmd)
	}

	errorHeight := m.errorToast.Height()
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

	playbackSectionHeight := lipgloss.Height(m.renderPlaybackSection())
	navHelpHeight := 1

	viewportHeight := msg.Height - navbarHeight - errorHeight - playbackSectionHeight - navHelpHeight

	if updatedLibrary, cmd, ok := updateSubmodel(m.library, tea.WindowSizeMsg{
		Height: viewportHeight,
	}, m.library); ok {
		m.library = updatedLibrary
		cmds = append(cmds, cmd)
	}

	libraryWidth := lipgloss.Width(m.library.View())

	queueWidth := 0
	if m.focusedModel == FocusQueue {
		queueWidth = lipgloss.Width(m.library.View())
	}

	mainContentViewWidth := msg.Width - libraryWidth - queueWidth

	if updatedSearchBar, cmd, ok := updateSubmodel(m.searchBar, tea.WindowSizeMsg{
		Width: mainContentViewWidth,
	}, m.searchBar); ok {
		m.searchBar = updatedSearchBar
		cmds = append(cmds, cmd)
	}

	searchBarHeight := lipgloss.Height(m.searchBar.View())
	mainViewportContentHeight := viewportHeight - searchBarHeight

	if updatedPlaylistView, cmd, ok := updateSubmodel(m.playlistView, tea.WindowSizeMsg{
		Width:  mainContentViewWidth,
		Height: mainViewportContentHeight,
	}, m.playlistView); ok {
		m.playlistView = updatedPlaylistView
		cmds = append(cmds, cmd)
	}

	if updatedSearchView, cmd, ok := updateSubmodel(m.searchView, tea.WindowSizeMsg{
		Width:  mainContentViewWidth,
		Height: mainViewportContentHeight,
	}, m.searchView); ok {
		m.searchView = updatedSearchView
		cmds = append(cmds, cmd)
	}

	if updatedQueueView, cmd, ok := updateSubmodel(m.queueView, tea.WindowSizeMsg{
		Height: viewportHeight,
	}, m.queueView); ok {
		m.queueView = updatedQueueView
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
