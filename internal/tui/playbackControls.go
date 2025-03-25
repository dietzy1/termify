package tui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
		if msg.Err != nil {
			log.Println("Error updating player state")
			return m, ShowError("Error updating player state", "Illegal action")
		}
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
		BorderForeground(BorderColor)

	var activeStyle = baseStyle.
		Foreground(PrimaryColor)

	if m.width < SHRINKWIDTH {
		baseStyle = baseStyle.Padding(0, 1)
		activeStyle = activeStyle.Padding(0, 1)
	}

	shuffleButton := "⇄"
	prevButton := "⏮"
	playPauseButton := "▶"
	nextButton := "⏭"
	repeatButton := "↺"

	playerState := m.spotifyState.GetPlayerState()

	// Get current states (changed from string comparison to boolean)
	shuffleState := playerState.ShuffleState
	repeatState := playerState.RepeatState

	//TODO: This is a daterace we need to write a function which accesses these things using the mutex to fix
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
			Foreground(lipgloss.Color(TextColor)).
			Align(lipgloss.Center).Foreground(lipgloss.Color(TextColor)).MaxHeight(1).
			Render(keybindTexts[i])

		var btn string
		switch {
		case (i == 0 && shuffleState) || (i == 4 && repeatState == "context") || (i == 2 && playerState.Playing):
			btn = activeStyle.Render(button)
		default:
			btn = baseStyle.Render(button)
		}

		renderedButton := lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Render(btn),
			helperContent,
		)
		renderedButtons = append(renderedButtons, renderedButton)
	}

	return lipgloss.JoinHorizontal(lipgloss.Bottom, renderedButtons...)
}
