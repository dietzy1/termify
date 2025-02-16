package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ tea.Model = (*playbackControlsModel)(nil)

type playbackControlsModel struct {
	width         int
	currentButton int
	volume        float64

	spotifyState *SpotifyState
}

func newPlaybackControlsModel(spotifyState *SpotifyState) playbackControlsModel {
	return playbackControlsModel{
		currentButton: 2,
		volume:        0.5,
		width:         0,
		spotifyState:  spotifyState,
	}
}

func (m playbackControlsModel) Init() tea.Cmd {
	return nil
}

func (m playbackControlsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case PlayerStateUpdatedMsg:
		log.Println("Playback controls recieved player state update")
		if msg.Err != nil {
			log.Println("Error updating player state")
			return m, nil
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Left):
			m.currentButton = (m.currentButton - 1 + 5) % 5
		case key.Matches(msg, DefaultKeyMap.Right):
			m.currentButton = (m.currentButton + 1) % 5
		case key.Matches(msg, DefaultKeyMap.Select):
			switch m.currentButton {
			case 0:
				return m, m.spotifyState.ToggleShuffleMode()
			case 1:
				return m, m.spotifyState.PreviousTrack()
			case 2:
				if m.spotifyState.playerState.Playing {
					return m, m.spotifyState.PausePlayback()
				}
				return m, m.spotifyState.StartPlayback()
			case 3:
				return m, m.spotifyState.NextTrack()
			case 4:
				return m, m.spotifyState.ToggleRepeatMode()
			}
		}
	}
	return m, nil
}

func (m playbackControlsModel) View() string {

	var baseStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor)).
		Bold(true).
		Padding(0, 3).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(BorderColor))

	var activeStyle = baseStyle.
		Foreground(lipgloss.Color(PrimaryColor)).
		BorderForeground(lipgloss.Color(PrimaryColor))

	if m.width < SHRINKWIDTH {
		baseStyle = baseStyle.Padding(0, 1)
		activeStyle = activeStyle.Padding(0, 1)
	}

	shuffleButton := "⇄"
	prevButton := "⏮"
	playPauseButton := "▶"
	nextButton := "⏭"
	repeatButton := "↺"
	playPauseHelper := "Play"

	// Get current states (changed from string comparison to boolean)
	shuffleState := m.spotifyState.playerState.ShuffleState
	repeatState := m.spotifyState.playerState.RepeatState

	if m.spotifyState.playerState.Playing {
		playPauseButton = "⏸"
		playPauseHelper = "Pause"
	}

	buttons := []string{shuffleButton, prevButton, playPauseButton, nextButton, repeatButton}
	helperTexts := []string{"Shuffle", "Previous", playPauseHelper, "Next", "Repeat"}
	var renderedButtons []string

	for i, button := range buttons {
		helperContent := lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextColor)).
			Width(8).
			Render(" ")

		if i == m.currentButton {
			helperContent = lipgloss.NewStyle().
				Foreground(lipgloss.Color(TextColor)).
				Width(8).
				Align(lipgloss.Center).
				Render(helperTexts[i])
		}

		var btn string
		switch {
		case i == m.currentButton:
			btn = activeStyle.Render(button)
		case (i == 0 && shuffleState) || (i == 4 && repeatState == "context"):
			btn = baseStyle.Foreground(lipgloss.Color(PrimaryColor)).Render(button)
		default:
			btn = baseStyle.Render(button)
		}

		// Maintain consistent dimensions for all buttons
		renderedButton := lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Padding(0, 1).Render(btn), // Add horizontal padding
			helperContent,
		)
		renderedButtons = append(renderedButtons, renderedButton)
	}

	playbackControls := lipgloss.JoinHorizontal(lipgloss.Bottom, renderedButtons...)

	containerStyle := lipgloss.NewStyle().
		/* Width(m.width-30). */
		Foreground(lipgloss.Color(TextColor)).
		Align(lipgloss.Center, lipgloss.Center)

	return containerStyle.Render(
		playbackControls,
	)
}
