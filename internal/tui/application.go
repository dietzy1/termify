package tui

import (
	"log"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
	"github.com/zmb3/spotify/v2"
)

// ErrorMsg represents an error that should be displayed to the user
type errorMsg struct {
	title   string
	message string
}

// ShowErrorMsg is a message type to show an error to the user
type ShowErrorMsg struct {
	Title   string
	Message string
}

// ErrorTimerExpiredMsg is sent when an error's display time has expired
type ErrorTimerExpiredMsg struct{}

type applicationModel struct {
	width, height int

	spotifyState *state.SpotifyState

	focusedModel FocusedModel

	errorBar          errorMsg
	errorDisplayTimer time.Time // Tracks when the error should be hidden
	navbar            navbarModel
	library           libraryModel

	searchBar       searchbarModel
	playlistView    playlistViewModel
	searchView      searchViewModel
	playbackControl playbackControlsModel
	audioPlayer     audioPlayerModel

	showHelp bool
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
		spotifyState:      spotifyState,
		focusedModel:      FocusLibrary,
		navbar:            newNavbar(spotifyState),
		library:           newLibrary(spotifyState),
		searchBar:         newSearchbar(spotifyState),
		playlistView:      NewPlaylistView(spotifyState),
		searchView:        NewSearchView(spotifyState),
		playbackControl:   newPlaybackControlsModel(spotifyState),
		audioPlayer:       newAudioPlayer(spotifyState),
		errorBar:          errorMsg{},
		errorDisplayTimer: time.Time{}, // Zero time (indicates no active timer)
		showHelp:          false,
	}
}

func (m applicationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case ShowErrorMsg:
		m.errorBar.title = msg.Title
		m.errorBar.message = msg.Message
		m.errorDisplayTimer = time.Now().Add(5 * time.Second)
		return m, tea.Batch(
			tea.WindowSize(),
			tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
				return ErrorTimerExpiredMsg{}
			}),
		)
	case ErrorTimerExpiredMsg:
		if !m.errorDisplayTimer.IsZero() && time.Now().After(m.errorDisplayTimer) {
			m.errorBar.title = ""
			m.errorBar.message = ""
			m.errorDisplayTimer = time.Time{}
		}
		return m, tea.WindowSize()

	case debouncedSearch:
		if updatedSearchBar, cmd, ok := updateSubmodel(m.searchBar, msg, m.searchBar); ok {
			m.searchBar = updatedSearchBar
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case NavigationMsg:
		m.focusedModel = msg.Target

		// If this navigation requires exiting search mode
		if msg.ExitSearch {
			m.searchBar.searching = false
			m.searchBar.textInput.Blur()
			m.searchBar.textInput.SetValue("")
		}

		// If we're navigating to the search bar, put it in search mode
		if msg.Target == FocusSearchBar {
			m.searchBar.EnterSearchMode()
		}

		// If a playlist ID is provided, fetch its tracks
		if msg.selectedID != "" {
			// Here we can switch on what kind of view we're navigating to

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

func (m applicationModel) View() string {
	if m.showHelp {
		return m.viewHelp()
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

	// Add a simple navigation help at the bottom
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

func (m applicationModel) renderPlaybackSection() string {
	// Get the song info and volume control views
	songInfoView := m.audioPlayer.songInfoView()
	volumeControlView := m.audioPlayer.volumeControlView()

	// Calculate the available width for the center section
	availableWidth := m.width - lipgloss.Width(songInfoView) - lipgloss.Width(volumeControlView) - 2

	// Style for the playback section
	combinedPlaybackSectionStyle := lipgloss.NewStyle().MaxWidth(m.width)

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

	return combinedPlaybackSectionStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Bottom,
			songInfoView,
			centerSection,
			volumeControlView),
	)
}

func (m applicationModel) renderErrorBar() string {
	if m.errorBar.title == "" || m.errorBar.message == "" {
		return ""
	}

	errorBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff4444")).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#ff4444")).
		Width(m.width-2).
		Height(2).
		Padding(0, 1)

	return errorBar.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().
			Bold(true).
			Render("Error: "+m.errorBar.title),
		"Details: "+m.errorBar.message,
	))
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

	msg.Height -= lipgloss.Height(m.navbar.View()) + lipgloss.Height(m.playbackControl.View()) + lipgloss.Height(m.audioPlayer.View()) + 1 + errorHeight

	if updatedNavbar, cmd, ok := updateSubmodel(m.navbar, tea.WindowSizeMsg{
		Width: msg.Width,
	}, m.navbar); ok {
		m.navbar = updatedNavbar
		cmds = append(cmds, cmd)
	}

	if updatedLibrary, cmd, ok := updateSubmodel(m.library, tea.WindowSizeMsg{
		Height: msg.Height,
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
		Height: msg.Height - lipgloss.Height(m.searchBar.View()),
	}, m.playlistView); ok {
		m.playlistView = updatedPlaylistView
		cmds = append(cmds, cmd)
	}

	if updatedSearchView, cmd, ok := updateSubmodel(m.searchView, tea.WindowSizeMsg{
		Width:  msg.Width - lipgloss.Width(m.library.View()),
		Height: msg.Height - lipgloss.Height(m.searchBar.View()),
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

func (m applicationModel) renderActionFeedback(action string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Align(lipgloss.Center).
		Width(m.width).
		Render(action)
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

// ShowError creates a command to show an error message
func ShowError(title, message string) tea.Cmd {
	return func() tea.Msg {
		return ShowErrorMsg{
			Title:   title,
			Message: message,
		}
	}
}
