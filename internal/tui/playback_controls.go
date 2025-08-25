package tui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/dietzy1/termify/internal/state"
)

var _ tea.Model = (*playbackControlsModel)(nil)

type playbackControlsModel struct {
	width int

	spotifyState *state.SpotifyState
}

func newPlaybackControlsModel(spotifyState *state.SpotifyState) playbackControlsModel {
	return playbackControlsModel{
		spotifyState: spotifyState,
	}
}

func (m playbackControlsModel) Init() tea.Cmd {
	return nil
}

func (m playbackControlsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case state.PlayerStateUpdatedMsg:
		log.Println("Playback controls recieved player state update")
		/* if msg.Err != nil {
			log.Println("Error updating player state")
			return m, ShowErrorToast("Error updating player state", "Illegal action")
		} */
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
	}
	return m, nil
}

func (m playbackControlsModel) View() string {
	var baseStyle = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 3).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).Background(BackgroundColor).BorderBackground(BackgroundColor)

	var activeStyle = baseStyle.
		Foreground(PrimaryColor)

	if m.width < SHRINKWIDTH {
		baseStyle = baseStyle.Padding(0, 1)
		activeStyle = activeStyle.Padding(0, 1)
		return ""
	}

	shuffleButton := "⇄"
	prevButton := "⏮"
	playPauseButton := "▶"
	nextButton := "⏭"
	repeatButton := "↺"

	playerState := m.spotifyState.GetPlayerState()

	shuffleState := playerState.ShuffleState
	repeatState := playerState.RepeatState

	if playerState.Playing {
		playPauseButton = "⏸"
	}

	buttons := []string{shuffleButton, prevButton, playPauseButton, nextButton, repeatButton}

	// Replace helper texts with keybinds
	keybindTexts := []string{
		DefaultKeyMap.Shuffle.Help().Key,
		DefaultKeyMap.Previous.Help().Key,
		DefaultKeyMap.PlayPause.Help().Key,
		DefaultKeyMap.Next.Help().Key,
		DefaultKeyMap.Repeat.Help().Key,
	}

	var renderedButtons []string

	for i, button := range buttons {
		helperContent := lipgloss.NewStyle().
			Foreground(TextColor).
			Background(BackgroundColor).
			MaxHeight(1).
			Width(9).
			Align(lipgloss.Center).
			Render(keybindTexts[i])

		var btn string
		switch {
		case (i == 0 && shuffleState) || (i == 4 && repeatState == "context") || (i == 2 && playerState.Playing):
			btn = activeStyle.Render(button)
		default:
			btn = baseStyle.Render(button)
		}

		renderedButton := lipgloss.JoinVertical(lipgloss.Center,
			btn,
			helperContent,
		)
		renderedButtons = append(renderedButtons, renderedButton)
	}

	return lipgloss.JoinHorizontal(lipgloss.Bottom, renderedButtons...)
}

func (m applicationModel) renderPlaybackSection() string {
	// Get the song info and volume control views
	songInfoView := m.audioPlayer.songInfoView()

	device := m.deviceSelector.View()
	volumeControlView := m.audioPlayer.volumeControlView()

	// Calculate the available width for the center section
	availableWidth := m.width - lipgloss.Width(songInfoView) - lipgloss.Width(volumeControlView) // was -2 before here

	// Style for the playback section
	combinedPlaybackSectionStyle := lipgloss.NewStyle().
		MaxWidth(m.width)

	// Center both components individually
	centeredPlaybackControls := lipgloss.NewStyle().
		Width(availableWidth).
		Align(lipgloss.Center).
		MaxHeight(4).Background(BackgroundColor).
		Render(m.playbackControl.View())

	centeredAudioPlayer := lipgloss.NewStyle().
		Width(availableWidth).
		Align(lipgloss.Center).
		MaxHeight(1).
		Render(m.audioPlayer.View())

	// Join them vertically
	centerSection := lipgloss.JoinVertical(
		lipgloss.Bottom,
		centeredPlaybackControls,
		centeredAudioPlayer,
	)

	rightSection := lipgloss.JoinVertical(
		lipgloss.Right,
		device,
		volumeControlView,
	)

	return combinedPlaybackSectionStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Bottom,
			songInfoView,
			centerSection,
			rightSection),
	)
}
