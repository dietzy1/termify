package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
)

var _ tea.Model = (*playbackControlsModel)(nil)

type playbackControlsModel struct {
	width         int
	currentButton int
	volume        float64

	spotifyState *state.SpotifyState
}

func newPlaybackControlsModel(spotifyState *state.SpotifyState) playbackControlsModel {
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
	case state.PlayerStateUpdatedMsg:
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
				if m.spotifyState.PlayerState.Playing {
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
	shuffleState := m.spotifyState.PlayerState.ShuffleState
	repeatState := m.spotifyState.PlayerState.RepeatState

	if m.spotifyState.PlayerState.Playing {
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
			Align(lipgloss.Center).
			Render(" ")

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

		renderedButton := lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Padding(0, 1).Render(btn),
			helperContent,
		)
		renderedButtons = append(renderedButtons, renderedButton)
	}

	return lipgloss.JoinHorizontal(lipgloss.Bottom, renderedButtons...)
}
